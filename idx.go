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

package stardict

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Entry struct {
	Word   string
	Offset uint64
	Size   uint32
}

// IdxScanner scans an index from start to end.
type IdxScanner struct {
	r             io.Reader
	s             *bufio.Scanner
	idxoffsetbits int
}

// NewIdxScanner return a new index scanner that scans the index from start to end.
func NewIdxScanner(r io.Reader, idxoffsetbits int64) (*IdxScanner, error) {
	if idxoffsetbits != 32 && idxoffsetbits != 64 {
		return nil, fmt.Errorf("invalid idxoffsetbits: %v", idxoffsetbits)
	}
	s := &IdxScanner{
		r:             r,
		s:             bufio.NewScanner(bufio.NewReader(r)),
		idxoffsetbits: int(idxoffsetbits),
	}
	s.s.Split(s.splitIndex)
	return s, nil
}

// Scan advances the index to the next index entry. It returns false if the
// scan stops either by reaching the end of the index or an error.
func (s *IdxScanner) Scan() bool {
	return s.s.Scan()
}

// Err returns the first error encountered.
func (s *IdxScanner) Err() error {
	return s.s.Err()
}

// Entry gets the next entry in the index.
func (s *IdxScanner) Entry() *Entry {
	var e Entry
	b := s.s.Bytes()
	if i := bytes.IndexByte(b, 0); i >= 0 {
		e.Word = string(b[0:i])
		if s.idxoffsetbits == 64 {
			e.Offset = binary.BigEndian.Uint64(b[i+1:])
		} else {
			e.Offset = uint64(binary.BigEndian.Uint32(b[i+1:]))
		}
		e.Size = binary.BigEndian.Uint32(b[i+1+s.idxoffsetbits/8:])
	}

	return &e
}

// splitIndex splits an index entry in the index file.
func (s *IdxScanner) splitIndex(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		// Found zero byte.
		tokenSize := i + 1 + s.idxoffsetbits/8 + 4
		if len(data) >= tokenSize {
			return tokenSize, data[:tokenSize], nil
		}
	}
	// Request more data.
	return 0, nil, nil
}

// Idx is a very basic implementation of an in memory index.
type Idx struct {
	idx     map[string][]int
	entries []*Entry
}

// NewIdx returns a new in-memory index.
func NewIdx(r io.Reader, idxoffsetbits int64) (*Idx, error) {
	idx := &Idx{
		idx: map[string][]int{},
	}

	i := 0
	s, err := NewIdxScanner(r, idxoffsetbits)
	if err != nil {
		return nil, err
	}
	for s.Scan() {
		e := s.Entry()
		for _, t := range tokenize(e.Word) {
			idx.idx[t] = append(idx.idx[t], i)
		}
		idx.entries = append(idx.entries, s.Entry())
		i++
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return idx, nil
}

// Prefix search searches entries by prefix.
func (idx *Idx) PrefixSearch(str string) []*Entry {
	// TODO: implement binary search over idx.entries
	var result []*Entry
	for _, e := range idx.entries {
		if strings.HasPrefix(strings.ToLower(e.Word), strings.ToLower(str)) {
			result = append(result, e)
		}
	}
	return result
}

// FullTextSearch searches full text of index entries.
func (idx *Idx) FullTextSearch(str string) []*Entry {
	var result []*Entry
	for _, w := range tokenize(str) {
		for _, id := range idx.idx[w] {
			result = append(result, idx.entries[id])
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
