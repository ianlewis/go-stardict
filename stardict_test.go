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

package stardict

import (
	"sync"

	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
)

type testDict struct {
	ifo  string
	dict []*dict.Word
	idx  []*idx.Word
}

// writeDict writes out a test dictionary set of files.
func writeDict(d testDict) (string, error) {
	path, err := ioutil.TempDir("", "stardict")
	if err != nil {
		return "", nil
	}

	if err := ioutil.WriteFile(filepath.Join(path, "dictionary.ifo"), []byte(d.ifo), 0600); err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filepath.Join(path, "dictionary.idx"), idx.MakeIndex(d.idx, 32), 0600); err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filepath.Join(path, "dictionary.dict"), dict.MakeDict(d.dict, nil), 0600); err != nil {
		return "", err
	}

	return path, nil
}

// TestOpen tests Open.
func TestOpen(t *testing.T) {
	tests := []struct {
		name  string
		dicts []testDict
	}{
		{
			name: "basic open",
			dicts: []testDict{
				{
					ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=6`,
					idx: []*idx.Word{
						{
							Word:   "hoge",
							Offset: 0,
							Size:   6,
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
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, td := range test.dicts {
				path, err := writeDict(td)
				if err != nil {
					t.Fatalf("writeDict: %v", err)
				}
				defer os.RemoveAll(path)

				s, err := Open(filepath.Join(path, "dictionary.ifo"))
				if err != nil {
					t.Fatalf("Open: %v", err)
				}

				// Open the .idx file.
				if _, err := s.Index(); err != nil {
					t.Fatalf("Index: %v", err)
				}

				// Open the .dict file.
				if _, err := s.Dict(); err != nil {
					t.Fatalf("Dict: %v", err)
				}
			}
		})
	}
}

// TestConcurrency tests that Stardict can be used concurrently.
func TestConcurrency(t *testing.T) {
	td := testDict{
		ifo: `StarDict's dict ifo file
version=3.0.0
bookname=hoge
wordcount=1
idxfilesize=6`,
		idx: []*idx.Word{
			{
				Word:   "hoge",
				Offset: 0,
				Size:   6,
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
	}

	path, err := writeDict(td)
	if err != nil {
		t.Fatalf("writeDict: %v", err)
	}
	defer os.RemoveAll(path)

	s, err := Open(filepath.Join(path, "dictionary.ifo"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	var entries []*Entry
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e, err := s.Search("hoge")
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			entries = append(entries, e...)
		}()
	}

	wg.Wait()

	if want, got := 1000, len(entries); want != got {
		t.Fatalf("Unexpected size: want %v, got: %v", want, got)
	}

}
