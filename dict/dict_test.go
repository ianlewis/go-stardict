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

package dict_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/internal/testutil"
)

func expectWordsEqual(t *testing.T, expected *dict.Word, word *dict.Word) {
	t.Helper()

	if want, got := len(expected.Data), len(word.Data); want != got {
		t.Fatalf("unexpected # of data; want: %d, got: %d", want, got)
	}
	for i := range expected.Data {
		if want, got := expected.Data[i].Type, word.Data[i].Type; want != got {
			t.Errorf("unexpected type; want: %v, got: %v", want, got)
		}
		if want, got := expected.Data[i].Data, word.Data[i].Data; !bytes.Equal(want, got) {
			t.Errorf("unexpected data; want: %#v, got: %#v", want, got)
		}
	}
}

// TestDict tests Dict.
func TestDict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dict             []*dict.Word
		index            *idx.Word
		expected         *dict.Word
		sametypesequence []dict.DataType
	}{
		{
			name: "utf",
			dict: []*dict.Word{
				{
					Data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(6),
			},
			expected: &dict.Word{
				Data: []*dict.Data{
					{
						Type: dict.UTFTextType,
						Data: []byte{'h', 'o', 'g', 'e'},
					},
				},
			},
		},
		{
			name: "utf sametype",
			sametypesequence: []dict.DataType{
				dict.UTFTextType,
			},
			dict: []*dict.Word{
				{
					Data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(4),
			},
			expected: &dict.Word{
				Data: []*dict.Data{
					{
						Type: dict.UTFTextType,
						Data: []byte{'h', 'o', 'g', 'e'},
					},
				},
			},
		},
		{
			name: "file type",
			dict: []*dict.Word{
				{
					Data: []*dict.Data{
						{
							Type: dict.WavType,
							Data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(9), // 1 (type) + 4 (file size) + 4 data
			},
			expected: &dict.Word{
				Data: []*dict.Data{
					{
						Type: dict.WavType,
						Data: []byte{'h', 'o', 'g', 'e'},
					},
				},
			},
		},
		{
			name: "file sametype",
			sametypesequence: []dict.DataType{
				dict.WavType,
			},
			dict: []*dict.Word{
				{
					Data: []*dict.Data{
						{
							Type: dict.WavType,
							Data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(8), // 4 (file size) + 4 data
			},
			expected: &dict.Word{
				Data: []*dict.Data{
					{
						Type: dict.WavType,
						Data: []byte{'h', 'o', 'g', 'e'},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.CreateTemp("", "stardict")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())

			if f.Write(testutil.MakeDict(test.dict, test.sametypesequence)); err != nil {
				t.Fatal(err)
			}
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				t.Fatal(err)
			}

			d, err := dict.New(f, test.sametypesequence)
			if err != nil {
				t.Fatal(err)
			}

			w, err := d.Word(test.index)
			if err != nil {
				t.Fatal(err)
			}

			expectWordsEqual(t, test.expected, w)
		})
	}
}
