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

package idx

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Scanner scans an index from start to end.
type Scanner struct {
	r             io.ReadCloser
	s             *bufio.Scanner
	idxoffsetbits int
}

// NewScanner return a new index scanner that scans the index from start to end.
func NewScanner(r io.ReadCloser, idxoffsetbits int64) (*Scanner, error) {
	if idxoffsetbits != 32 && idxoffsetbits != 64 {
		return nil, fmt.Errorf("invalid idxoffsetbits: %v", idxoffsetbits)
	}
	s := &Scanner{
		r:             r,
		s:             bufio.NewScanner(bufio.NewReader(r)),
		idxoffsetbits: int(idxoffsetbits),
	}
	s.s.Split(s.splitIndex)
	return s, nil
}

// Scan advances the index to the next index entry. It returns false if the
// scan stops either by reaching the end of the index or an error.
func (s *Scanner) Scan() bool {
	return s.s.Scan()
}

// Err returns the first error encountered.
func (s *Scanner) Err() error {
	return s.s.Err()
}

// Close closes the underlying reader.
func (s *Scanner) Close() error {
	return s.r.Close()
}

// Word gets the next entry in the index.
func (s *Scanner) Word() *Word {
	var e Word
	b := s.s.Bytes()
	if i := bytes.IndexByte(b, 0); i >= 0 {
		e.Word = string(b[0:i])
		if s.idxoffsetbits == 64 {
			e.Offset = binary.BigEndian.Uint64(b[i+1:])
		} else {
			e.Offset = uint64(binary.BigEndian.Uint32(b[i+1:]))
		}
		e.Size = binary.BigEndian.Uint32(b[i+1+s.idxoffsetbits/8:])
	}

	return &e
}

// splitIndex splits an index entry in the index file.
func (s *Scanner) splitIndex(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		// Found zero byte.
		tokenSize := i + 1 + s.idxoffsetbits/8 + 4
		if len(data) >= tokenSize {
			return tokenSize, data[:tokenSize], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
