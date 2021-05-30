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

package dict

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ianlewis/go-stardict/idx"
)

func makeDict(words []*Word, sametypesequence []DataType) (*os.File, error) {
	b := []byte{}
	for _, w := range words {
		data := w.Data()
		for i, d := range data {
			// TODO: Support file type data.
			if len(sametypesequence) == 0 {
				b = append(b, byte(d.Type()))
				b = append(b, d.Data()...)
				b = append(b, 0) // Append a zero byte terminator.
			} else {
				b = append(b, d.Data()...)
				// Null terminator is not present on the last data item.
				if i == len(data)-1 {
					b = append(b, 0) // Append a zero byte terminator.
				}
			}
		}
	}

	tmpfile, err := ioutil.TempFile("", "dict")
	if err != nil {
		return nil, err
	}
	if _, err := tmpfile.Write(b); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, err
	}

	if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, err
	}
	return tmpfile, nil
}

func expectWordsEqual(t *testing.T, expected *Word, word *Word) {
	expectedData := expected.Data()
	wordData := word.Data()
	if want, got := len(expectedData), len(wordData); want != got {
		t.Fatalf("unexpected # of data; want: %d, got: %d", want, got)
	}
	for i := range expectedData {
		if want, got := expectedData[i].Type(), wordData[i].Type(); want != got {
			t.Errorf("unexpected type; want: %v, got: %v", want, got)
		}
		if want, got := expectedData[i].Data(), wordData[i].Data(); bytes.Compare(want, got) != 0 {
			t.Errorf("unexpected data; want: %#v, got: %#v", want, got)
		}
	}
}

// TestDict tests Dict.
func TestDict(t *testing.T) {
	tests := []struct {
		name             string
		dict             []*Word
		index            *idx.Word
		expected         *Word
		sametypesequence []DataType
	}{
		{
			name: "utf",
			dict: []*Word{
				{
					data: []*Data{
						{
							t:    UTFTextType,
							data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(6),
			},
			expected: &Word{
				data: []*Data{
					{
						t:    UTFTextType,
						data: []byte{'h', 'o', 'g', 'e'},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := makeDict(test.dict, test.sametypesequence)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())

			// TODO: support isDictZip
			d, err := New(f, test.sametypesequence, false)
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
