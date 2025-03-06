// Copyright 2025 Ian Lewis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syn

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/transform"

	"github.com/ianlewis/go-stardict/internal/index"
)

// Word is a .syn file entry.
type Word struct {
	// Word is the synonym word.
	Word string

	// OriginalWordIndex is the index into the .idx index.
	OriginalWordIndex uint32
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
	// Folder returns a [transform.Transformer] that performs folding (e.g.
	// case folding, whitespace folding, etc.) on index entries.
	Folder func() transform.Transformer
}

// DefaultOptions is the default options for a Syn.
var DefaultOptions = &Options{
	Folder: func() transform.Transformer {
		return transform.Nop
	},
}

// Syn is is the synonym index. It is largely a map of synonym words to related
// index entries.
type Syn struct {
	// index is sorted by the folded word value.
	index *index.Index[*foldedWord]

	// foldTransformer performs folding on text.
	foldTransformer func() transform.Transformer
}

// New returns a new Syn by reading the data from r.
func New(r io.ReadCloser, options *Options) (*Syn, error) {
	if options == nil {
		options = DefaultOptions
	}

	syn := Syn{
		foldTransformer: DefaultOptions.Folder,
	}
	if options.Folder != nil {
		syn.foldTransformer = options.Folder
	}

	i := 0
	s, err := NewScanner(r)
	if err != nil {
		return nil, fmt.Errorf("creating synonym index scanner: %w", err)
	}

	var words []*foldedWord
	for s.Scan() {
		word := s.Word()
		folded, _, err := transform.String(syn.foldTransformer(), word.Word)
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
		return nil, fmt.Errorf("scanning synonym index %w", err)
	}

	// We need to re-sort based on the folded word.
	syn.index = index.NewIndex(words, strings.Compare)

	return &syn, nil
}

// NewFromIfoPath returns a new in-memory index.
func NewFromIfoPath(ifoPath string, options *Options) (*Syn, error) {
	var r io.ReadCloser
	f, err := Open(ifoPath)
	if err != nil {
		return nil, err
	}
	r = f

	idxExt := strings.ToLower(filepath.Ext(f.Name()))
	if idxExt == ".gz" || idxExt == ".dz" {
		r, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("creating .ifo gzip reader: %w", err)
		}
	}

	return New(r, options)
}

// Open opens the .syn file given the path to the .ifo file.
func Open(ifoPath string) (*os.File, error) {
	baseName := strings.TrimSuffix(ifoPath, filepath.Ext(ifoPath))

	synExts := []string{
		".syn",
		".syn.gz",
		".syn.GZ",
		".syn.dz",
		".syn.DZ",
		".SYN",
		".SYN.gz",
		".SYN.GZ",
		".SYN.dz",
		".SYN.DZ",
	}
	var f *os.File
	var err error
	for _, ext := range synExts {
		f, err = os.Open(baseName + ext)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("opening .syn file: %w", err)
		}
	}

	// Catch the case when no .syn file was found.
	if err != nil {
		return nil, fmt.Errorf("opening .syn file: %w", err)
	}

	return f, nil
}

// Search performs a query of the index and returns matching words.
func (syn *Syn) Search(query string) ([]*Word, error) {
	foldedQuery, _, err := transform.String(syn.foldTransformer(), query)
	if err != nil {
		return nil, fmt.Errorf("folding query %q: %w", query, err)
	}

	result := syn.index.Search(foldedQuery)

	var words []*Word
	for _, w := range result {
		words = append(words, w.word)
	}

	return words, nil
}
