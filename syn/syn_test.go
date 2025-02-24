// Copyright 2025 Ian Lewis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syn_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/go-stardict/internal/testutil"
	"github.com/ianlewis/go-stardict/syn"
)

// TestSyn_Search tests Syn.Search.
func TestSyn_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		query    string
		synWords []*syn.Word

		expected []*syn.Word
	}{
		{
			name:     "empty index",
			query:    "foo",
			synWords: []*syn.Word{},

			expected: nil,
		},
		{
			name:  "no match",
			query: "hoge",
			synWords: []*syn.Word{
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

			expected: nil,
		},
		{
			name:  "single match first",
			query: "bar",
			synWords: []*syn.Word{
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

			expected: []*syn.Word{
				{
					Word: "bar",
				},
			},
		},
		{
			name:  "single match last",
			query: "foo",
			synWords: []*syn.Word{
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

			expected: []*syn.Word{
				{
					Word: "foo",
				},
			},
		},
		{
			name:  "single match middle",
			query: "hoge",
			synWords: []*syn.Word{
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

			expected: []*syn.Word{
				{
					Word: "hoge",
				},
			},
		},
		{
			name:  "multi-match",
			query: "hoge",
			synWords: []*syn.Word{
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
					Word:              "hoge",
					OriginalWordIndex: 123,
				},
				{
					Word:              "hoge",
					OriginalWordIndex: 234,
				},
				{
					Word:              "hoge",
					OriginalWordIndex: 345,
				},
				{
					Word: "pico",
				},
			},

			expected: []*syn.Word{
				{
					Word:              "hoge",
					OriginalWordIndex: 123,
				},
				{
					Word:              "hoge",
					OriginalWordIndex: 234,
				},
				{
					Word:              "hoge",
					OriginalWordIndex: 345,
				},
			},
		},
		{
			name:  "folding",
			query: "hoge",
			synWords: []*syn.Word{
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

			expected: []*syn.Word{
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
			synWords: []*syn.Word{
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

			expected: []*syn.Word{
				{
					// NOTE: The returned index word is the value in the index
					//       and not the folded value.
					Word: "grüßen",
				},
			},
		},
		{
			name:  "folding whitespace",
			query: "\u3000 こんにちは \t 世界 \u3000 ",
			synWords: []*syn.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "こんにちは\u3000世界",
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

			expected: []*syn.Word{
				{
					// NOTE: The returned index word is the value in the index
					//       and not the folded value.
					Word: "こんにちは\u3000世界",
				},
			},
		},
		{
			name:  "folding punctuation",
			query: "foo. bar?",
			synWords: []*syn.Word{
				{
					Word: "bar",
				},
				{
					Word: "baz",
				},
				{
					Word: "foo bar",
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

			expected: []*syn.Word{
				{
					// NOTE: The returned index word is the value in the index
					//       and not the folded value.
					Word: "foo bar",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			b := testutil.MakeSyn(t, test.synWords)

			index, err := syn.New(io.NopCloser(bytes.NewReader(b)))
			if err != nil {
				t.Fatalf("syn.New: %v", err)
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
