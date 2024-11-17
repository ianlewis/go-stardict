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
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/internal/testutil"
)

// TestData_String tests Data.String.
func TestData_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     *dict.Data
		expected string
	}{
		{
			name: "UTFTextType",
			data: &dict.Data{
				Type: dict.UTFTextType,
				Data: []byte("ユニコード"),
			},
			expected: "ユニコード",
		},
		{
			name: "PhoneticType",
			data: &dict.Data{
				Type: dict.PhoneticType,
				Data: []byte("ゆにこーど"),
			},
			expected: "ゆにこーど",
		},
		{
			name: "HTMLType",
			data: &dict.Data{
				Type: dict.HTMLType,
				Data: []byte("<html><head><title>Title</title></head><body>Body</body></html>"),
			},
			expected: "Body",
		},
		{
			name: "XDXFType",
			data: &dict.Data{
				Type: dict.XDXFType,
				Data: []byte("Some XDXF Format"),
			},
			// TODO(#22): Support other formats.
			expected: "",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(test.expected, test.data.String()); diff != "" {
				t.Fatalf("Data.String (-want, +got):\n%s", diff)
			}
		})
	}
}

// TestDict_Word tests Dict.Word.
func TestDict_Word(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		dict             []*dict.Word
		index            *idx.Word
		expected         *dict.Word
		sameTypeSequence []dict.DataType
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
			sameTypeSequence: []dict.DataType{
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
			sameTypeSequence: []dict.DataType{
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
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := testutil.MakeTempDict(t, test.dict, &testutil.MakeDictOptions{
				SameTypeSequence: test.sameTypeSequence,
			})
			defer f.Close()
			defer os.Remove(f.Name())

			d, err := dict.New(f, &dict.Options{
				SameTypeSequence: test.sameTypeSequence,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()

			w, err := d.Word(test.index)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expected, w); diff != "" {
				t.Fatalf("Dict.Word (-want, +got):\n%s", diff)
			}
		})
	}
}

// TestDict_NewFromIfoPath tests NewFromIfoPath.
func TestDict_NewFromIfoPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  *testutil.MakeDictOptions
		dict     []*dict.Word
		index    *idx.Word
		expected *dict.Word
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
			options: &testutil.MakeDictOptions{
				Ext: ".DICT",
				SameTypeSequence: []dict.DataType{
					dict.UTFTextType,
				},
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
			options: &testutil.MakeDictOptions{
				SameTypeSequence: []dict.DataType{
					dict.WavType,
				},
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
		{
			name: "dictzip",
			options: &testutil.MakeDictOptions{
				DictZip: true,
			},
			dict: []*dict.Word{
				{
					Data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte("hoge"),
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
						Data: []byte("hoge"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := testutil.MakeTempDict(t, test.dict, test.options)
			defer f.Close()
			defer os.Remove(f.Name())
			ifoPath := strings.TrimSuffix(f.Name(), test.options.GetExt()) + ".ifo"

			d, err := dict.NewFromIfoPath(ifoPath, &dict.Options{
				SameTypeSequence: test.options.GetSameTypeSequence(),
			})
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()

			w, err := d.Word(test.index)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expected, w); diff != "" {
				t.Fatalf("Dict.Word (-want, +got):\n%s", diff)
			}
		})
	}
}
