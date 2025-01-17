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

package syn_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/go-stardict/internal/testutil"
	"github.com/ianlewis/go-stardict/syn"
)

// TestSynScanner tests the SynScanner type.
func TestSynScanner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected []*syn.Word
	}{
		{
			name: "multi 64 bit",
			expected: []*syn.Word{
				{
					Word:              "hoge",
					OriginalWordIndex: 5,
				},
				{
					Word:              "fuga pico",
					OriginalWordIndex: 3,
				},
			},
		},
		{
			name: "multi 32 bit",
			expected: []*syn.Word{
				{
					Word:              "hoge",
					OriginalWordIndex: 5,
				},
				{
					Word:              "fuga pico",
					OriginalWordIndex: 3,
				},
			},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			b := testutil.MakeSyn(t, test.expected)

			var words []*syn.Word
			s, err := syn.NewScanner(io.NopCloser(bytes.NewReader(b)))
			if err != nil {
				t.Fatal(err)
			}
			for s.Scan() {
				words = append(words, s.Word())
			}
			if err := s.Err(); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expected, words); diff != "" {
				t.Fatalf("words (-want, +got):\n%s", diff)
			}
		})
	}
}
