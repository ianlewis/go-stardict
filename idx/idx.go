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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

// Options are options for the idx data.
type Options struct {
	// OffsetBits are the number of bits in the offset fields.
	OffsetBits int
}

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// Scanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	words           []*foldedWord
	foldTransformer transform.Transformer
}

// New returns a new in-memory index.
func New(r io.ReadCloser, options *Options) (*Idx, error) {
	if options == nil {
		options = &Options{
			OffsetBits: 32,
		}
	}

	idx := &Idx{
		foldTransformer: transform.Chain(
			norm.NFD,
			cases.Fold(),
			runes.Remove(runes.In(unicode.Mn)),
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

		idx.words = append(idx.words, &foldedWord{
			folded: folded,
			word:   word,
		})
		i++
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning index %w", err)
	}

	// We need to re-sort based on the folded word.
	slices.SortFunc(idx.words, func(a, b *foldedWord) int {
		return strings.Compare(a.folded, b.folded)
	})

	return idx, nil
}

func openIdxFile(ifoPath string) (*os.File, error) {
	baseName := strings.TrimSuffix(ifoPath, filepath.Ext(ifoPath))

	idxExts := []string{".idx.gz", ".idx", ".IDX", ".IDX.gz", ".IDX.GZ"}
	var f *os.File
	var err error
	for _, ext := range idxExts {
		f, err = os.Open(baseName + ext)
		// TODO: check for os.ErrNotExist
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("opening .idx file: %w", err)
	}

	return f, err
}

// NewFromIfoPath returns a new in-memory index.
func NewFromIfoPath(ifoPath string, options *Options) (*Idx, error) {
	var r io.ReadCloser
	f, err := openIdxFile(ifoPath)
	if err != nil {
		return nil, err
	}
	r = f

	idxExt := filepath.Ext(f.Name())
	//nolint:gocritic // strings.EqualFold should not be used here.
	if strings.ToLower(idxExt) == ".gz" {
		r, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("creating .ifo gzip reader: %w", err)
		}
	}

	return New(r, options)
}

// Search performs a query of the index and returns matching words.
func (idx *Idx) Search(query string) ([]*Word, error) {
	foldedQuery, _, err := transform.String(idx.foldTransformer, query)
	if err != nil {
		return nil, fmt.Errorf("folding query %q: %w", query, err)
	}

	start := 0
	end := len(idx.words) - 1
	for start <= end {
		pivot := (start + end) / 2
		switch {
		case idx.words[pivot].folded < foldedQuery:
			start = pivot + 1
		case idx.words[pivot].folded > foldedQuery:
			end = pivot - 1
		default:
			// We have found a matching word.
			// Multiple word entries may have the same value we must find the
			// first and iterate over the index until we have found all matches.
			i := pivot
			for i > 0 && idx.words[i-1].folded == foldedQuery {
				i--
			}
			j := pivot
			for j+1 < len(idx.words) && idx.words[j+1].folded == foldedQuery {
				j++
			}

			var result []*Word
			for ; i < j+1; i++ {
				result = append(result, idx.words[i].word)
			}

			return result, nil
		}
	}

	return nil, nil
}
