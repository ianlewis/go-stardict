// Copyright 2021 Google LLC
// Copyright 2025 Ian Lewis
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

package stardict

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/internal/testutil"
	"github.com/ianlewis/go-stardict/syn"
)

type testDict struct {
	ifo  string
	dict []*dict.Word
	idx  []*idx.Word
	syn  []*syn.Word
}

// writeDict writes out a test dictionary set of files.
func writeDict(t *testing.T, d *testDict) string {
	t.Helper()

	path, err := os.MkdirTemp("", "stardict")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filepath.Join(path, "dictionary.ifo"), []byte(d.ifo), 0o600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(path, "dictionary.idx"), testutil.MakeIndex(d.idx, 32), 0o600); err != nil {
		panic(err)
	}
	// NOTE: syn is optional.
	if len(d.syn) > 0 {
		if err := os.WriteFile(filepath.Join(path, "dictionary.syn"), testutil.MakeSyn(t, d.syn), 0o600); err != nil {
			panic(err)
		}
	}
	if err := os.WriteFile(filepath.Join(path, "dictionary.dict"), testutil.MakeDict(t, d.dict, nil), 0o600); err != nil {
		panic(err)
	}

	return path
}

// TestOpen tests Open.
func TestOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		dicts []*testDict

		err       error
		bookname  string
		wordcount int64
	}{
		{
			name: "basic open",
			dicts: []*testDict{
				{
					ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=123
idxfilesize=6`,
				},
			},

			bookname:  "hoge",
			wordcount: 123,
		},
		{
			name: "invalid idxoffsetbits",
			dicts: []*testDict{
				{
					ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=6
idxoffsetbits=123`,
				},
			},
			err: idx.ErrInvalidIdxOffset,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			for _, td := range test.dicts {
				path := writeDict(t, td)
				defer os.RemoveAll(path)

				s, err := Open(filepath.Join(path, "dictionary.ifo"), nil)
				if diff := cmp.Diff(test.err, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("Open: (-want, +got):\n%s", diff)
				}
				if test.err != nil {
					continue
				}

				if diff := cmp.Diff(s.bookname, s.Bookname()); diff != "" {
					t.Errorf("Bookname: (-want, +got):\n%s", diff)
				}

				if diff := cmp.Diff(s.wordcount, s.WordCount()); diff != "" {
					t.Errorf("WordCount: (-want, +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		dict  *testDict
		query string

		expected []*Entry
		err      error
	}{
		{
			name: "empty dictionary",
			dict: &testDict{
				ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=0
idxfilesize=0`,
			},
			query: "foo",

			expected: nil,
			err:      nil,
		},
		{
			name: "index search",
			dict: &testDict{
				ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=0`,
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
				idx: []*idx.Word{
					{
						Word:   "hoge",
						Offset: 0,
						Size:   6,
					},
				},
			},
			query: "hoge",

			expected: []*Entry{
				{
					word: "hoge",
					data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte{'h', 'o', 'g', 'e'},
						},
					},
				},
			},
			err: nil,
		},

		{
			name: "syn search",
			dict: &testDict{
				ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=0`,
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
				idx: []*idx.Word{
					{
						Word:   "hoge",
						Offset: 0,
						Size:   6,
					},
				},
				syn: []*syn.Word{
					{
						Word:              "foo",
						OriginalWordIndex: 0,
					},
				},
			},
			query: "foo",

			expected: []*Entry{
				{
					word: "hoge",
					data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte("hoge"),
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "combined idx syn search",
			dict: &testDict{
				ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=0`,
				dict: []*dict.Word{
					{
						Data: []*dict.Data{
							{
								Type: dict.UTFTextType,
								Data: []byte("hoge"),
							},
							{
								Type: dict.UTFTextType,
								Data: []byte("foo"),
							},
						},
					},
				},
				idx: []*idx.Word{
					{
						Word:   "hoge",
						Offset: 0,
						Size:   6,
					},
					{
						Word:   "foo",
						Offset: 6,
						Size:   5,
					},
				},
				syn: []*syn.Word{
					{
						Word:              "foo",
						OriginalWordIndex: 0,
					},
				},
			},
			query: "foo",

			expected: []*Entry{
				{
					word: "foo",
					data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte("foo"),
						},
					},
				},
				{
					word: "hoge",
					data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte("hoge"),
						},
					},
				},
			},
			err: nil,
		},
		{
			name: "folding idx syn search",
			dict: &testDict{
				ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=0`,
				dict: []*dict.Word{
					{
						Data: []*dict.Data{
							{
								Type: dict.UTFTextType,
								Data: []byte("hoge"),
							},
							{
								Type: dict.UTFTextType,
								Data: []byte("foo"),
							},
							{
								Type: dict.UTFTextType,
								Data: []byte("grussen"),
							},
						},
					},
				},
				idx: []*idx.Word{
					{
						Word:   "hoge",
						Offset: 0,
						Size:   6,
					},
					{
						Word:   "foo",
						Offset: 6,
						Size:   5,
					},
					{
						Word:   "grüßen",
						Offset: 11,
						Size:   9,
					},
				},
				syn: []*syn.Word{
					{
						Word:              "grussen",
						OriginalWordIndex: 0,
					},
				},
			},
			query: "grussen",

			expected: []*Entry{
				{
					word: "grüßen",
					data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte("grussen"),
						},
					},
				},
				{
					word: "hoge",
					data: []*dict.Data{
						{
							Type: dict.UTFTextType,
							Data: []byte("hoge"),
						},
					},
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := writeDict(t, test.dict)
			defer os.RemoveAll(path)

			d, err := Open(filepath.Join(path, "dictionary.ifo"), nil)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			results, err := d.Search(test.query)
			if diff := cmp.Diff(test.err, err); diff != "" {
				t.Errorf("Search (-want, +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.expected, results, cmp.AllowUnexported(Entry{})); diff != "" {
				t.Errorf("Search (-want, +got):\n%s", diff)
			}
		})
	}
}

// TODO(#1): Restore concurrency test
// TestConcurrency tests that Stardict can be used concurrently.
// func TestConcurrency(t *testing.T) {
//	td := testDict{
//		ifo: `StarDict's dict ifo file
// version=3.0.0
// bookname=hoge
// wordcount=1
// idxfilesize=6`,
//		idx: []*idx.Word{
//			{
//				Word:   "hoge",
//				Offset: 0,
//				Size:   6,
//			},
//		},
//		dict: []*dict.Word{
//			{
//				Data: []*dict.Data{
//					{
//						Type: dict.UTFTextType,
//						Data: []byte{'h', 'o', 'g', 'e'},
//					},
//				},
//			},
//		},
//	}

//	path, err := writeDict(td)
//	if err != nil {
//		t.Fatalf("writeDict: %v", err)
//	}
//	defer os.RemoveAll(path)

//	s, err := Open(filepath.Join(path, "dictionary.ifo"))
//	if err != nil {
//		t.Fatalf("Open: %v", err)
//	}

//	var entries []*Entry
//	var mu sync.Mutex
//	var wg sync.WaitGroup
//	for i := 0; i < 1000; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			e, err := s.Search("hoge")
//			if err != nil {
//				return
//			}
//			mu.Lock()
//			defer mu.Unlock()
//			entries = append(entries, e...)
//		}()
//	}

//	wg.Wait()

//	if want, got := 1000, len(entries); want != got {
//		t.Fatalf("Unexpected size: want %v, got: %v", want, got)
//	}

// }
