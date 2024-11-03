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
	"fmt"
	"io"
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

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// Scanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	words           []*foldedWord
	foldTransformer transform.Transformer
}

// New returns a new in-memory index.
func New(r io.ReadCloser, idxoffsetbits int64) (*Idx, error) {
	idx := &Idx{
		foldTransformer: transform.Chain(
			norm.NFD,
			cases.Fold(),
			runes.Remove(runes.In(unicode.Mn)),
			norm.NFC,
		),
	}

	i := 0
	s, err := NewScanner(r, idxoffsetbits)
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
