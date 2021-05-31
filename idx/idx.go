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
	"strings"
	"unicode"
)

// Word is an .idx file entry.
type Word struct {
	Word   string
	Offset uint64
	Size   uint32
}

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// IdxScanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	idx   map[string][]int
	words []*Word
}

// New returns a new in-memory index.
func New(r io.Reader, idxoffsetbits int64) (*Idx, error) {
	idx := &Idx{
		idx: map[string][]int{},
	}

	i := 0
	s, err := NewIdxScanner(r, idxoffsetbits)
	if err != nil {
		return nil, err
	}
	for s.Scan() {
		e := s.Word()
		for _, t := range tokenize(e.Word) {
			idx.idx[t] = append(idx.idx[t], i)
		}
		idx.words = append(idx.words, s.Word())
		i++
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return idx, nil
}

// FullTextSearch searches full text of index entries.
func (idx *Idx) FullTextSearch(str string) []*Word {
	var result []*Word
	for _, w := range tokenize(str) {
		for _, id := range idx.idx[w] {
			result = append(result, idx.words[id])
		}
	}
	return result
}

// tokenize tokenizes English text in a very basic way.
func tokenize(str string) []string {
	words := strings.FieldsFunc(str, func(r rune) bool {
		// Split on any character that is not a letter or a number.
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	var stopwords = map[string]struct{}{
		"a": {}, "and": {}, "be": {}, "i": {},
		"in": {}, "of": {}, "that": {}, "the": {},
		"this": {}, "to": {},
	}

	var tokens []string
	for _, w := range words {
		t := strings.ToLower(w)
		if _, ok := stopwords[t]; !ok {
			tokens = append(tokens, t)
		}
	}

	return tokens
}
