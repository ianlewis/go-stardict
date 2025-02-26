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

package index

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type String string

func (s String) String() string {
	return string(s)
}

func TestIndex_string(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		index    []String
		query    string
		expected []String
	}{
		{
			name:     "single results",
			index:    []String{"foo", "bar", "baz", "bar"},
			query:    "foo",
			expected: []String{"foo"},
		},
		{
			name:     "multiple results",
			index:    []String{"foo", "bar", "baz", "bar"},
			query:    "bar",
			expected: []String{"bar", "bar"},
		},
		{
			name:     "no results",
			index:    []String{"foo", "bar", "baz", "bar"},
			query:    "none",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			index := NewIndex(test.index, strings.Compare)

			if diff := cmp.Diff(test.expected, index.Search(test.query)); diff != "" {
				t.Fatalf("Search (-want, +got):\n%s", diff)
			}
		})
	}
}
