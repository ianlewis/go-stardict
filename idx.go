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
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Entry struct {
	Word   string
	Offset uint64
	Size   uint32
}

// Idx is a stardict dictionary index. A stardict index is a simple list of //
// word entries and their offset and size in the dictionary file. Idx provides
// functionality to read the index into entries and basic means for word lookup.
// Dictionary applications that wish to provide search functionality should
// provide their own search capabilities (e.g. fuzzy search etc.)
type Idx struct {
	r             io.ReadCloser
	idxoffsetbits int
	err           error
	s             *bufio.Scanner
}

func NewIdx(r io.ReadCloser, idxoffsetbits int64) (*Idx, error) {
	if idxoffsetbits != 32 && idxoffsetbits != 64 {
		return nil, fmt.Errorf("invalid idxoffsetbits: %v", idxoffsetbits)
	}

	s := bufio.NewScanner(bufio.NewReader(r))
	idx := &Idx{
		r:             r,
		idxoffsetbits: int(idxoffsetbits),
		s:             s,
	}
	s.Split(idx.splitIndex)
	return idx, nil
}

// Scan advances the index to the next index entry. It returns false if the
// scan stops either by reaching the end of the index or an error.
func (idx *Idx) Scan() bool {
	return idx.s.Scan()
}

// Err returns the first error encountered.
func (idx *Idx) Err() error {
	return idx.s.Err()
}

// Entry gets the next entry in the index.
func (idx *Idx) Entry() *Entry {
	var e Entry
	b := idx.s.Bytes()
	if i := bytes.IndexByte(b, 0); i >= 0 {
		e.Word = string(b[0:i])
		if idx.idxoffsetbits == 64 {
			e.Offset = binary.BigEndian.Uint64(b[i+1:])
		} else {
			e.Offset = uint64(binary.BigEndian.Uint32(b[i+1:]))
		}
		e.Size = binary.BigEndian.Uint32(b[i+1+idx.idxoffsetbits/8:])
	}

	return &e
}

// Close closes the index file.
func (idx *Idx) Close() error {
	return idx.r.Close()
}

// splitIndex splits an index entry in the index file.
func (idx *Idx) splitIndex(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		// Found zero byte.
		tokenSize := i + 1 + idx.idxoffsetbits/8 + 4
		if len(data) >= tokenSize {
			return tokenSize, data[:tokenSize], nil
		}
	}
	// Request more data.
	return 0, nil, nil
}
