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
	"errors"
	"fmt"
	"io"
)

// ErrInvalidIdxOffset indicates that the OffsetBits is an invalid value.
var ErrInvalidIdxOffset = errors.New("invalid idxoffsetbits")

// Scanner scans an index from start to end.
type Scanner struct {
	r             io.ReadCloser
	s             *bufio.Scanner
	idxoffsetbits int
}

// ScannerOptions are options for scanning an .idx file.
type ScannerOptions struct {
	// OffsetBits are the number of bits in the offset fields. Valid values for
	// OffsetBits are either 32 or 64.
	OffsetBits int
}

// DefaultScannerOptions is the default options for a Scanner.
var DefaultScannerOptions = &ScannerOptions{
	OffsetBits: 32,
}

// NewScanner return a new index scanner that scans the index from start to
// end. The Scanner assumes ownership of the reader and should be closed with the
// Close method.
func NewScanner(r io.ReadCloser, options *ScannerOptions) (*Scanner, error) {
	if options == nil {
		options = DefaultScannerOptions
	}

	if options.OffsetBits != 32 && options.OffsetBits != 64 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidIdxOffset, options.OffsetBits)
	}
	s := &Scanner{
		r:             r,
		s:             bufio.NewScanner(bufio.NewReader(r)),
		idxoffsetbits: options.OffsetBits,
	}
	s.s.Split(s.splitIndex)
	return s, nil
}

// NewScannerFromIfoPath returns a new in-memory index.
func NewScannerFromIfoPath(ifoPath string, options *ScannerOptions) (*Scanner, error) {
	f, err := Open(ifoPath)
	if err != nil {
		return nil, err
	}
	return NewScanner(f, options)
}

// Scan advances the index to the next index entry. It returns false if the
// scan stops either by reaching the end of the index or an error.
func (s *Scanner) Scan() bool {
	return s.s.Scan()
}

// Err returns the first error encountered.
func (s *Scanner) Err() error {
	//nolint:wrapcheck // error should not be wrapped
	return s.s.Err()
}

// Close closes the underlying reader.
func (s *Scanner) Close() error {
	err := s.r.Close()
	if err != nil {
		return fmt.Errorf("closing idx file: %w", err)
	}
	return nil
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
