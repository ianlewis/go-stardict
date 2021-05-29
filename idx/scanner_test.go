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
	"bytes"
	"encoding/binary"
	"testing"
)

// Makes index bytes given a list of words.
func makeIndex(words []*Word, idxoffsetbits int64) []byte {
	b := []byte{}
	for _, w := range words {
		b = append(b, []byte(w.Word)...)
		b = append(b, 0) // Add the zero byte terminator.
		var b2 []byte
		if idxoffsetbits == 64 {
			b2 = make([]byte, 12)
			binary.BigEndian.PutUint64(b2[:8], w.Offset)
			binary.BigEndian.PutUint32(b2[8:12], w.Size)
		} else {
			b2 = make([]byte, 8)
			binary.BigEndian.PutUint32(b2[:4], uint32(w.Offset))
			binary.BigEndian.PutUint32(b2[4:8], w.Size)
		}
		b = append(b, b2...)
	}
	return b
}

// expectWordsEqual compares two word lists
func expectWordsEqual(t *testing.T, expected, words []*Word) {
	if want, got := len(expected), len(words); want != got {
		t.Fatalf("unexpected # of words; want: %d, got: %d", want, got)
		return
	}
	for i := range expected {
		if want, got := *expected[i], *words[i]; want != got {
			t.Errorf("unexpected word; want: %#v, got: %#v", want, got)
		}
	}
}

// TestIdxScanner tests IdxScanner
func TestIdxScanner(t *testing.T) {
	tests := []struct {
		name          string
		expected      []*Word
		idxoffsetbits int64
	}{
		{
			name: "multi 64 bit",
			expected: []*Word{
				{
					Word:   "hoge",
					Offset: 123,
					Size:   456,
				},
				{
					Word:   "fuga pico",
					Offset: 12,
					Size:   45,
				},
			},
			idxoffsetbits: 64,
		},
		{
			name: "multi 32 bit",
			expected: []*Word{
				{
					Word:   "hoge",
					Offset: 123,
					Size:   456,
				},
				{
					Word:   "fuga pico",
					Offset: 12,
					Size:   45,
				},
			},
			idxoffsetbits: 32,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			b := makeIndex(test.expected, test.idxoffsetbits)

			var words []*Word
			s, err := NewIdxScanner(bytes.NewReader(b), test.idxoffsetbits)
			if err != nil {
				t.Fatal(err)
			}
			for s.Scan() {
				words = append(words, s.Word())
			}
			if err := s.Err(); err != nil {
				t.Fatal(err)
			}
			expectWordsEqual(t, test.expected, words)
		})
	}
}
