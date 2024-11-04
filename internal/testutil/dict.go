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

package testutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/pebbe/dictzip"

	"github.com/ianlewis/go-stardict/dict"
)

type MakeDictOptions struct {
	// Ext is the .dict file's extension (e.g. .dict).
	Ext string

	// DictZip indicates that the dict file should be compressed with DictZip.
	DictZip bool

	// SameTypeSequence is the sametypesequence option.
	SameTypeSequence []dict.DataType
}

// MakeTempDict creates a temporary .dict file and returns the path to the file.
func MakeTempDict(t *testing.T, words []*dict.Word, opts *MakeDictOptions) string {
	t.Helper()

	f, err := os.CreateTemp("", "stardict.*"+opts.Ext)
	if err != nil {
		t.Fatal(err)
	}

	d := MakeDict(t, words, opts.SameTypeSequence)

	if opts.DictZip {
		// Just get the file name.
		path := f.Name()
		f.Close()

		fmt.Println(path)
		err = dictzip.Write(bytes.NewReader(d), path, 9)
		if err != nil {
			t.Fatal(err)
		}

		return path
	}

	defer f.Close()

	_, err = f.Write(d)
	if err != nil {
		t.Fatal(err)
	}

	return f.Name()
}

// MakeDict creates a test .dict file.
func MakeDict(t *testing.T, words []*dict.Word, sametypesequence []dict.DataType) []byte {
	t.Helper()

	b := []byte{}
	for _, w := range words {
		for i, d := range w.Data {
			if len(sametypesequence) == 0 {
				b = append(b, byte(d.Type))
				if 'a' <= d.Type && d.Type <= 'z' {
					// Data is a string like sequence.
					b = append(b, d.Data...)
					b = append(b, 0) // Append a zero byte terminator.
				} else {
					// Data is a file like sequence.
					sizeBytes := make([]byte, 4)
					dataLen := len(d.Data)
					if dataLen > math.MaxUint32 {
						t.Fatalf("word data too long: %d", dataLen)
					}
					binary.BigEndian.PutUint32(sizeBytes, uint32(dataLen))
					b = append(b, sizeBytes...)
					b = append(b, d.Data...)
				}
			} else {
				if 'a' <= d.Type && d.Type <= 'z' {
					// Data is a string like sequence.
					b = append(b, d.Data...)
					// Null terminator is not present on the last data item.
					if i == len(w.Data)-1 {
						b = append(b, 0)
					}
				} else {
					// Data is a file like sequence.
					sizeBytes := make([]byte, 4)
					dataLen := len(d.Data)
					if dataLen > math.MaxUint32 {
						t.Fatalf("word data too long: %d", dataLen)
					}
					binary.BigEndian.PutUint32(sizeBytes, uint32(dataLen))
					b = append(b, sizeBytes...)
					b = append(b, d.Data...)
				}
			}
		}
	}

	return b
}
