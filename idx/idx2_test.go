// Copyright 2024 Google LLC
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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/text/transform"
)

func Test_whitespaceFolder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		src   []byte
		dst   []byte
		atEOF bool

		expected []byte
		nDst     int
		nSrc     int
		err      error
	}{
		{
			name:  "leading whitespace",
			src:   []byte(" \t\u3000foo"),
			dst:   make([]byte, 5),
			atEOF: false,

			expected: []byte{'f', 'o', 'o', 0, 0},
			nDst:     3,
			nSrc:     8,
			err:      nil,
		},
		{
			name:  "trailing whitespace",
			src:   []byte("foo \t\u3000"),
			dst:   make([]byte, 5),
			atEOF: false,

			expected: []byte{'f', 'o', 'o', 0, 0},
			nDst:     3,
			nSrc:     8,
			err:      nil,
		},
		{
			name:  "whitespace spans",
			src:   []byte("foo \t\u3000 bar \t\u3000 baz"),
			dst:   make([]byte, 12),
			atEOF: false,

			expected: []byte{'f', 'o', 'o', ' ', 'b', 'a', 'r', ' ', 'b', 'a', 'z', 0},
			nDst:     11,
			nSrc:     21,
			err:      nil,
		},
		{
			name:  "all whitespace",
			src:   []byte(" \t\u3000 foo \t\u3000 bar \t\u3000 baz \t\u3000"),
			dst:   make([]byte, 12),
			atEOF: false,

			expected: []byte{'f', 'o', 'o', ' ', 'b', 'a', 'r', ' ', 'b', 'a', 'z', 0},
			nDst:     11,
			nSrc:     32,
			err:      nil,
		},
		{
			name:  "fill dst",
			src:   []byte(" \t\u3000 foo \t\u3000 bar \t\u3000 baz \t\u3000"),
			dst:   make([]byte, 11),
			atEOF: false,

			expected: []byte{'f', 'o', 'o', ' ', 'b', 'a', 'r', ' ', 'b', 'a', 'z'},
			nDst:     11,
			nSrc:     32,
			err:      nil,
		},
		{
			name:  "short dst",
			src:   []byte(" \t\u3000 foo \t\u3000 bar \t\u3000 baz \t\u3000"),
			dst:   make([]byte, 3),
			atEOF: false,

			expected: []byte{'f', 'o', 'o'},
			nDst:     3,
			nSrc:     15,
			err:      transform.ErrShortDst,
		},
		{
			name: "short src incomplete unicode",
			// NOTE: the last character is only partially included.
			src:   []byte(" \t\u3000 foo \t\u3000")[:12],
			dst:   make([]byte, 10),
			atEOF: false,

			expected: []byte{'f', 'o', 'o', 0, 0, 0, 0, 0, 0, 0},
			nDst:     3,
			nSrc:     11,
			err:      transform.ErrShortSrc,
		},
		{
			name: "incomplete unicode at EOF",
			// NOTE: the last character is only partially included.
			src:   []byte(" \t\u3000 foo \t\u3000")[:12],
			dst:   make([]byte, 10),
			atEOF: true,

			// NOTE: []byte{0xef, 0xbf, 0xbd} is utf8.RuneError.
			expected: []byte{'f', 'o', 'o', ' ', 0xef, 0xbf, 0xbd, 0, 0, 0},
			nDst:     7,
			nSrc:     12,
			err:      nil,
		},
		{
			name: "invalid unicode",
			// NOTE: the last character is only partially included.
			src:   []byte{' ', 'f', 'o', 'o', ' ', 0xe3, ' ', ' ', 'b', 'a', 'r'},
			dst:   make([]byte, 12),
			atEOF: true,

			// NOTE: []byte{0xef, 0xbf, 0xbd} is utf8.RuneError.
			// Invalid unicode characters are treated as non-space characters.
			expected: []byte{'f', 'o', 'o', ' ', 0xef, 0xbf, 0xbd, ' ', 'b', 'a', 'r', 0},
			nDst:     11,
			nSrc:     11,
			err:      nil,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ws := whitespaceFolder{}
			nDst, nSrc, err := ws.Transform(test.dst, test.src, test.atEOF)
			if diff := cmp.Diff(test.nDst, nDst); diff != "" {
				t.Fatalf("nDst (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.nSrc, nSrc); diff != "" {
				t.Fatalf("nSrc (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("err (-want, +got):\n%s", diff)
			}
		})
	}
}
