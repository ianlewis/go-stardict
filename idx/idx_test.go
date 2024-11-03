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

package idx_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/internal/testutil"
)

// TestIdx_Search tests Idx.Search.
func TestIdx_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		query         string
		idxWords      []*idx.Word
		idxoffsetbits int

		expected []*idx.Word
	}{
		{
			name:          "empty index",
			query:         "foo",
			idxWords:      []*idx.Word{},
			idxoffsetbits: 32,

			expected: nil,
		},
		{
			name:  "no match",
			query: "hoge",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
			},
			idxoffsetbits: 32,

			expected: nil,
		},
		{
			name:  "single match first",
			query: "bar",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
			},
			idxoffsetbits: 32,

			expected: []*idx.Word{
				{
					Word: "bar",
				},
			},
		},
		{
			name:  "single match last",
			query: "foo",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
			},
			idxoffsetbits: 32,

			expected: []*idx.Word{
				{
					Word: "foo",
				},
			},
		},
		{
			name:  "single match middle",
			query: "hoge",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
				{
					Word: "fuga",
				},
				{
					Word: "hoge",
				},
				{
					Word: "pico",
				},
			},
			idxoffsetbits: 32,

			expected: []*idx.Word{
				{
					Word: "hoge",
				},
			},
		},
		{
			name:  "multi-match",
			query: "hoge",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
				{
					Word: "fuga",
				},
				{
					Word:   "hoge",
					Offset: 123,
					Size:   456,
				},
				{
					Word:   "hoge",
					Offset: 234,
					Size:   567,
				},
				{
					Word:   "hoge",
					Offset: 345,
					Size:   678,
				},
				{
					Word: "pico",
				},
			},
			idxoffsetbits: 32,

			expected: []*idx.Word{
				{
					Word:   "hoge",
					Offset: 123,
					Size:   456,
				},
				{
					Word:   "hoge",
					Offset: 234,
					Size:   567,
				},
				{
					Word:   "hoge",
					Offset: 345,
					Size:   678,
				},
			},
		},
		{
			name:  "folding",
			query: "hoge",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
				{
					Word: "fuga",
				},
				{
					Word: "Hoge",
				},
				{
					Word: "pico",
				},
			},
			idxoffsetbits: 32,

			expected: []*idx.Word{
				{
					// NOTE: The returned index word is the value in the index
					//       and not the folded value.
					Word: "Hoge",
				},
			},
		},
		{
			name:  "folding german",
			query: "grussen",
			idxWords: []*idx.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo",
				},
				{
					Word: "fuga",
				},
				{
					Word: "grüßen",
				},
				{
					Word: "Hoge",
				},
				{
					Word: "pico",
				},
			},
			idxoffsetbits: 32,

			expected: []*idx.Word{
				{
					// NOTE: The returned index word is the value in the index
					//       and not the folded value.
					Word: "grüßen",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			b := testutil.MakeIndex(test.idxWords, test.idxoffsetbits)

			index, err := idx.New(io.NopCloser(bytes.NewReader(b)), &idx.Options{
				OffsetBits: test.idxoffsetbits,
			})
			if err != nil {
				t.Fatalf("idx.New: %v", err)
			}

			result, err := index.Search(test.query)
			if diff := cmp.Diff(nil, err); diff != "" {
				t.Fatalf("b.Search (-want, +got):\n%s", diff)
			}

			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Fatalf("b.Search (-want, +got):\n%s", diff)
			}
		})
	}
}
