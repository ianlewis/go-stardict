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
	"fmt"

	"bytes"
	"encoding/binary"
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
			if len(sametypesequence) == 0 {
				t := d.Type()
				b = append(b, byte(t))
				if 'a' <= t && t <= 'z' {
					// Data is a string like sequence.
					b = append(b, d.Data()...)
					b = append(b, 0) // Append a zero byte terminator.
				} else {
					// Data is a file like sequence.
					sizeBytes := make([]byte, 4)
					data := d.Data()
					binary.BigEndian.PutUint32(sizeBytes, uint32(len(data)))
					b = append(b, sizeBytes...)
					b = append(b, data...)
					fmt.Printf("makeDict: %v\n", b)
				}
			} else {
				t := d.Type()
				if 'a' <= t && t <= 'z' {
					// Data is a string like sequence.
					b = append(b, d.Data()...)
					// Null terminator is not present on the last data item.
					if i == len(data)-1 {
						b = append(b, 0)
					}
				} else {
					// Data is a file like sequence.
					sizeBytes := make([]byte, 4)
					data := d.Data()
					binary.BigEndian.PutUint32(sizeBytes, uint32(len(data)))
					b = append(b, sizeBytes...)
					b = append(b, data...)
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
		{
			name: "utf sametype",
			sametypesequence: []DataType{
				UTFTextType,
			},
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
				Size:   uint32(4),
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
		{
			name: "file type",
			dict: []*Word{
				{
					data: []*Data{
						{
							t:    WavType,
							data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(9), // 1 (type) + 4 (file size) + 4 data
			},
			expected: &Word{
				data: []*Data{
					{
						t:    WavType,
						data: []byte{'h', 'o', 'g', 'e'},
					},
				},
			},
		},
		{
			name: "file sametype",
			sametypesequence: []DataType{
				WavType,
			},
			dict: []*Word{
				{
					data: []*Data{
						{
							t:    WavType,
							data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			index: &idx.Word{
				Word:   "hoge",
				Offset: uint64(0),
				Size:   uint32(8), // 4 (file size) + 4 data
			},
			expected: &Word{
				data: []*Data{
					{
						t:    WavType,
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

			d, err := New(f, test.sametypesequence)
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
