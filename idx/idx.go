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
	"io"
)

// Word is an .idx file entry.
type Word struct {
	Word   string
	Offset uint64
	Size   uint32
}

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// Scanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	words []*Word
}

// New returns a new in-memory index.
func New(r io.ReadCloser, idxoffsetbits int64) (*Idx, error) {
	idx := &Idx{}

	i := 0
	s, err := NewScanner(r, idxoffsetbits)
	if err != nil {
		return nil, err
	}
	for s.Scan() {
		idx.words = append(idx.words, s.Word())
		i++
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return idx, nil
}

// Search performs a query of the index and returns matching words.
func (idx *Idx) Search(query string) []*Word {
	start := 0
	end := len(idx.words) - 1
	for start <= end {
		pivot := (start + end) / 2
		switch {
		case idx.words[pivot].Word < query:
			start = pivot + 1
		case idx.words[pivot].Word > query:
			end = pivot - 1
		default:
			// We have found a matching word.
			// Multiple word entries may have the same value we must find the
			// first and iterate over the index until we have found all matches.
			i := pivot
			for i > 0 && idx.words[i-1].Word == query {
				i--
			}
			j := pivot
			for j+1 < len(idx.words) && idx.words[j+1].Word == query {
				j++
			}
			return idx.words[i : j+1]
		}
	}

	return nil
}
