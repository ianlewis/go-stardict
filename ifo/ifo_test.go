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

package ifo

import (
	"bytes"
	"testing"
)

// TestIfo tests Ifo
func TestIfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		expect func(*testing.T, *Ifo)
		err    bool
	}{
		{
			name: "magic and version",
			data: `test magic
version=1.0.0`,
			expect: func(t *testing.T, i *Ifo) {
				t.Helper()
				if want, got := "test magic", i.Magic(); want != got {
					t.Fatalf("magic; want: %q, got: %q", want, got)
				}
				if want, got := "1.0.0", i.Value("version"); want != got {
					t.Fatalf("version; want: %q, got: %q", want, got)
				}
			},
		},
		{
			name: "missing version",
			data: `test magic`,
			err:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			i, err := New(bytes.NewReader([]byte(test.data)))
			if test.err && err == nil {
				t.Fatal("New: expected failure")
			}
			if !test.err && err != nil {
				t.Fatalf("New: %v", err)
			}
			if test.expect != nil {
				test.expect(t, i)
			}
		})
	}
}
