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
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/text/cases"

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
		options       *idx.Options

		expected []*idx.Word
		err      error
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
			options: &idx.Options{
				Folder: cases.Fold(),
			},

			expected: []*idx.Word{
				{
					// NOTE: The returned index word is the value in the index
					//       and not the folded value.
					Word: "Hoge",
				},
			},
		},
		{
			name:  "glob search folded",
			query: "Fu[G]A*",
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
				{
					Word: "fUga hoge",
				},
			},
			idxoffsetbits: 32,
			options: &idx.Options{
				Folder: cases.Fold(),
			},

			// NOTE: The returned index word is the value in the index
			//       and not the folded value.
			expected: []*idx.Word{
				{
					Word: "fuga",
				},
				{
					Word: "fUga hoge",
				},
			},
		},
		{
			name:  "glob no prefix",
			query: "*uga",
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
			options: &idx.Options{
				Folder: cases.Fold(),
			},

			expected: nil,
			err:      idx.ErrPrefix,
		},
		{
			name:  "glob err",
			query: "[fuga",
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
			options: &idx.Options{
				Folder: cases.Fold(),
			},

			expected: nil,
			err:      idx.ErrGlob,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			b := testutil.MakeIndex(test.idxWords, test.idxoffsetbits)

			index, err := idx.New(io.NopCloser(bytes.NewReader(b)), test.options)
			if err != nil {
				t.Fatalf("idx.New: %v", err)
			}

			result, err := index.Search(test.query)
			if diff := cmp.Diff(test.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("b.Search (-want, +got):\n%s", diff)
			}

			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Fatalf("b.Search (-want, +got):\n%s", diff)
			}
		})
	}
}
