// Copyright 2021 Google LLC
// Copyright 2025 Ian Lewis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package idx

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/gobwas/glob/syntax"
	"golang.org/x/text/transform"

	"github.com/ianlewis/go-stardict/internal/index"
	"github.com/ianlewis/go-stardict/syn"
)

var (
	// ErrPrefix indicates that the query must not start with a glob wildcard.
	ErrPrefix = errors.New("search query must not start with wildcard")

	// ErrGlob indicates an error with a glob search query.
	ErrGlob = errors.New("invalid glob query")
)

// Word is an .idx file entry.
type Word struct {
	// Word is the word as it appears in the index.
	Word string

	// Offset is the offset in the .dict file that the corresponding entry appears.
	Offset uint64

	// Size is the total size of the corresponding .dict file entry.
	Size uint32
}

type foldedWord struct {
	folded string
	word   *Word
}

func (w *foldedWord) String() string {
	return w.folded
}

// Options are options for the idx data.
type Options struct {
	// Folder is the transformer that performs folding on index entries.
	Folder transform.Transformer

	// ScannerOptions are the options to use when reading the .idx file.
	ScannerOptions *ScannerOptions
}

// DefaultOptions is the default options for an Idx.
var DefaultOptions = &Options{
	Folder: transform.Nop,
	ScannerOptions: &ScannerOptions{
		OffsetBits: 32,
	},
}

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// Scanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	// index is sorted by the folded word value.
	index *index.Index[*foldedWord]

	// foldTransformer performs folding on text.
	foldTransformer transform.Transformer
}

// New returns a new in-memory index.
func New(r io.ReadCloser, options *Options) (*Idx, error) {
	return NewWithSyn(r, nil, options)
}

// New returns a new in-memory index with synonyms merged in.
func NewWithSyn(idxReader, synReader io.ReadCloser, options *Options) (*Idx, error) {
	if options == nil {
		options = DefaultOptions
	}

	idx := &Idx{
		foldTransformer: DefaultOptions.Folder,
	}
	if options.Folder != nil {
		idx.foldTransformer = options.Folder
	}

	i := 0
	s, err := NewScanner(idxReader, options.ScannerOptions)
	if err != nil {
		return nil, fmt.Errorf("creating index scanner: %w", err)
	}

	var words []*foldedWord
	for s.Scan() {
		word := s.Word()
		folded, _, err := transform.String(idx.foldTransformer, word.Word)
		if err != nil {
			return nil, fmt.Errorf("folding word %q: %w", word.Word, err)
		}
		words = append(words, &foldedWord{
			folded: folded,
			word:   word,
		})
		i++
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning index: %w", err)
	}

	// Merge in options.Syn.
	if synReader != nil {
		synScanner, err := syn.NewScanner(synReader)
		if err != nil {
			return nil, fmt.Errorf("scanning synonym index: %w", err)
		}
		for synScanner.Scan() {
			word := synScanner.Word()
			folded, _, err := transform.String(idx.foldTransformer, word.Word)
			if err != nil {
				return nil, fmt.Errorf("folding word %q: %w", word.Word, err)
			}
			words = append(words, &foldedWord{
				folded: folded,
				word:   words[word.OriginalWordIndex].word,
			})
		}
	}

	idx.index = index.NewIndex(words, prefixCmp)

	return idx, nil
}

// Open opens the .idx file given the path to the .ifo file.
func Open(ifoPath string) (*os.File, error) {
	baseName := strings.TrimSuffix(ifoPath, filepath.Ext(ifoPath))

	idxExts := []string{
		".idx",
		".idx.gz",
		".idx.GZ",
		".idx.dz",
		".idx.DZ",
		".IDX",
		".IDX.gz",
		".IDX.GZ",
		".IDX.dz",
		".IDX.DZ",
	}
	var f *os.File
	var err error
	for _, ext := range idxExts {
		f, err = os.Open(baseName + ext)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("opening .idx file: %w", err)
		}
	}

	// Catch the case when no .idx file was found.
	if err != nil {
		return nil, fmt.Errorf("opening .idx file: %w", err)
	}

	return f, nil
}

// NewFromIfoPath returns a new in-memory index.
func NewFromIfoPath(ifoPath string, options *Options) (*Idx, error) {
	var idxReader, synReader io.ReadCloser
	idxFile, err := Open(ifoPath)
	if err != nil {
		return nil, err
	}
	idxReader = idxFile

	idxExt := strings.ToLower(filepath.Ext(idxFile.Name()))
	if idxExt == ".gz" || idxExt == ".dz" {
		idxReader, err = gzip.NewReader(idxReader)
		if err != nil {
			return nil, fmt.Errorf("creating .idx gzip reader: %w", err)
		}
	}

	synFile, err := syn.Open(ifoPath)
	if !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			//nolint:wrapcheck // it isn't necessary to wrap this error.
			return nil, err
		}
		synReader = synFile

		synExt := strings.ToLower(filepath.Ext(synFile.Name()))
		if synExt == ".gz" || synExt == ".dz" {
			synReader, err = gzip.NewReader(synReader)
			if err != nil {
				return nil, fmt.Errorf("creating .syn gzip reader: %w", err)
			}
		}
	}

	return NewWithSyn(idxReader, synReader, options)
}

// Search performs a query of the index and returns matching words. The query
// supports glob patterns whose pattern syntax is:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` ] { character-range } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	character-range:
//	    c           matches character c (c != `\\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
//
// The pattern is folded using the given folding transformer and matches the
// folded word in the index.
func (idx *Idx) Search(query string) ([]*Word, error) {
	foldedQuery, err := idx.foldGlob(query)
	if err != nil {
		return nil, err
	}

	g, err := glob.Compile(foldedQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrGlob, query)
	}

	// Get the prefix to the query.
	var b strings.Builder
	for _, r := range foldedQuery {
		if syntax.Special(byte(r)) {
			break
		}
		b.WriteRune(r)
	}
	prefix := b.String()
	if prefix == "" {
		return nil, fmt.Errorf("%w: %q", ErrPrefix, query)
	}

	// Get all results with the static prefix.
	result := idx.index.Search(prefix)

	var words []*Word
	for _, w := range result {
		if g.Match(w.folded) {
			words = append(words, w.word)
		}
	}

	return words, nil
}

// foldGlob performs folding on glob non-special characters.
func (idx *Idx) foldGlob(q string) (string, error) {
	var s []string

	var b strings.Builder
	var isSpecial bool
	for i := 0; i < len(q); i++ {
		c := q[i]
		if syntax.Special(c) {
			if !isSpecial {
				if b.Len() > 0 {
					w, _, err := transform.String(idx.foldTransformer, b.String())
					if err != nil {
						return "", fmt.Errorf("folding query %q: %w", q, err)
					}

					s = append(s, w)
				}
				isSpecial = true
				b.Reset()
			}
		} else {
			if isSpecial {
				if b.Len() > 0 {
					s = append(s, b.String())
				}
				isSpecial = false
				b.Reset()
			}
		}

		b.WriteByte(c)
	}

	w := b.String()
	if !isSpecial {
		// fold the word
		fw, _, err := transform.String(idx.foldTransformer, w)
		if err != nil {
			return "", fmt.Errorf("folding query %q: %w", q, err)
		}
		w = fw
	}
	if b.Len() > 0 {
		s = append(s, w)
	}

	return strings.Join(s, ""), nil
}

func prefixCmp(a, b string) int {
	if strings.HasPrefix(b, a) {
		return 0
	}

	return strings.Compare(a, b)
}
