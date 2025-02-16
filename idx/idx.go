// Copyright 2021 Google LLC
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
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/ianlewis/go-stardict/internal/folding"
)

var errIndex = errors.New("invalid index")

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

// Options are options for the idx data.
type Options struct {
	// OffsetBits are the number of bits in the offset fields.
	OffsetBits int
}

// DefaultOptions is the default options for an Idx.
var DefaultOptions = &Options{
	OffsetBits: 32,
}

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// Scanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	// words is indexed by the original file index.
	words []*Word

	// folded is sorted by the folded word value.
	folded []*foldedWord

	// foldTransformer performs folding on text.
	foldTransformer transform.Transformer
}

// New returns a new in-memory index.
func New(r io.ReadCloser, options *Options) (*Idx, error) {
	if options == nil {
		options = DefaultOptions
	}

	idx := &Idx{
		foldTransformer: transform.Chain(
			// Unicode Normalization Form D (Canonical Decomposition.
			norm.NFD,
			// Perform case folding.
			cases.Fold(),
			// Perform whitespace folding.
			&folding.WhitespaceFolder{},
			// Remove Non-spacing marks ([, ] {, }, etc.).
			runes.Remove(runes.In(unicode.Mn)),
			// Remove punctuation.
			runes.Remove(runes.In(unicode.P)),
			// Unicode Normalization Form C (Canonical Decomposition, followed by Canonical Composition)
			// NOTE: Case folding does not normalize the input and may not
			// preserve a normal form. Canonical Decomposition is thus necessary
			// to be performed a second time.
			norm.NFC,
		),
	}

	i := 0
	s, err := NewScanner(r, options)
	if err != nil {
		return nil, fmt.Errorf("creating index scanner: %w", err)
	}
	for s.Scan() {
		word := s.Word()
		folded, _, err := transform.String(idx.foldTransformer, word.Word)
		if err != nil {
			return nil, fmt.Errorf("folding word %q: %w", word.Word, err)
		}
		idx.words = append(idx.words, word)
		idx.folded = append(idx.folded, &foldedWord{
			folded: folded,
			word:   word,
		})
		i++
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning index %w", err)
	}

	// We need to re-sort based on the folded word.
	sort.Slice(idx.folded, func(i, j int) bool {
		return idx.folded[i].folded < idx.folded[j].folded
	})

	return idx, nil
}

func openIdxFile(ifoPath string) (*os.File, error) {
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
	var r io.ReadCloser
	f, err := openIdxFile(ifoPath)
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

// ByIndex returns the word at the original index in the .idx file.
func (idx *Idx) ByIndex(i uint32) (*Word, error) {
	intIndex := int(i)
	if intIndex >= len(idx.words) {
		return nil, fmt.Errorf("%w: %d", errIndex, intIndex)
	}
	return idx.words[intIndex], nil
}

// Search performs a query of the index and returns matching words.
func (idx *Idx) Search(query string) ([]*Word, error) {
	foldedQuery, _, err := transform.String(idx.foldTransformer, query)
	if err != nil {
		return nil, fmt.Errorf("folding query %q: %w", query, err)
	}

	start := 0
	end := len(idx.folded) - 1
	for start <= end {
		pivot := (start + end) / 2
		switch {
		case idx.folded[pivot].folded < foldedQuery:
			start = pivot + 1
		case idx.folded[pivot].folded > foldedQuery:
			end = pivot - 1
		default:
			// We have found a matching word.
			// Multiple word entries may have the same value we must find the
			// first and iterate over the index until we have found all matches.
			i := pivot
			for i > 0 && idx.folded[i-1].folded == foldedQuery {
				i--
			}
			j := pivot
			for j+1 < len(idx.folded) && idx.folded[j+1].folded == foldedQuery {
				j++
			}

			var result []*Word
			for ; i < j+1; i++ {
				result = append(result, idx.folded[i].word)
			}

			return result, nil
		}
	}

	return nil, nil
}
