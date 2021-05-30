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

// package dict implements reading .dict files.
package dict

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pebbe/dictzip"

	"github.com/ianlewis/go-stardict/idx"
)

// Dict represents a Stardict dictionary's dictionary data.
type Dict struct {
	r                io.ReadSeekCloser
	dz               *dictzip.Reader
	sametypesequence []DataType
}

// Word is a full dictionary entry.
type Word struct {
	data []*Data
}

// Data returns all data for this word in order.
func (a Word) Data() []*Data {
	return a.data
}

// DataType is a type of dictionary.
type DataType byte

const (
	UTFTextType          = DataType('m')
	LocaleTextType       = DataType('l')
	PangoTextType        = DataType('g')
	PhoneticType         = DataType('t')
	XdxfType             = DataType('x')
	YinBiaoOrKataType    = DataType('y')
	PowerDataType        = DataType('p')
	MediaWikiType        = DataType('w')
	HTMLType             = DataType('h')
	WordNetType          = DataType('n')
	ResourceFileListType = DataType('r')
	WavType              = DataType('W')
	PictureType          = DataType('P')
	ExperimentalType     = DataType('X')
)

type Data struct {
	t    DataType
	data []byte
}

func (w Data) Type() DataType {
	return w.t
}

func (w Data) Data() []byte {
	return w.data
}

// New returns a new Dict from the given reader. Dict takes ownership of the
// reader. The reader can be closed via the Dict's Close method.
func New(r io.ReadSeekCloser, sametypesequence []DataType, isDictZip bool) (*Dict, error) {
	// verify sametypesequence
	for _, s := range sametypesequence {
		switch s {
		case UTFTextType,
			LocaleTextType,
			PangoTextType,
			PhoneticType,
			XdxfType,
			YinBiaoOrKataType,
			PowerDataType,
			MediaWikiType,
			HTMLType,
			WordNetType,
			ResourceFileListType,
			WavType,
			PictureType,
			ExperimentalType:
		default:
			return nil, fmt.Errorf("invalid type: %v", s)
		}
	}

	var dzReader *dictzip.Reader
	var err error
	if isDictZip {
		dzReader, err = dictzip.NewReader(r)
		if err != nil {
			return nil, err
		}
	}

	return &Dict{
		r:                r,
		dz:               dzReader,
		sametypesequence: sametypesequence,
	}, nil
}

// Word retrieves the word for the given index entry from the
// dictionary.
func (d *Dict) Word(e *idx.Word) (*Word, error) {
	b, err := d.getDictBytes(e)
	if err != nil {
		return nil, err
	}

	var wordData []*Data
	if len(d.sametypesequence) > 0 {
		// When sametypesequence is specified, that determines the type of the
		// word's data.
		for _, t := range d.sametypesequence {
			var data []byte
			if 'a' <= t && t <= 'z' {
				// Data is a string like sequence.
				i := bytes.IndexByte(b, 0)
				if i >= 0 {
					i += 1
				} else {
					// Use the full length of the buffer if no null terminator
					// is found. The final data won't have one.
					i = len(b)
				}
				data = b[:i]
				b = b[i:]
			} else {
				// Data is a file like sequence.
				size := binary.BigEndian.Uint32(b)
				data = b[4 : 4+size]
				b = b[4+size:]
			}
			wordData = append(wordData, &Data{
				t:    t,
				data: data,
			})
		}
	} else {
		for len(b) > 0 {
			t := DataType(b[0])
			b = b[1:]

			var data []byte
			if 'a' <= t && t <= 'z' {
				// Data is a string like sequence.
				i := bytes.IndexByte(b, 0)
				if i < 0 {
					i = len(b)
				}
				data = b[:i]
				b = b[i+1:] // Skip the null terminator
			} else {
				// Data is a file like sequence.
				size := binary.BigEndian.Uint32(b)
				data = b[4 : 4+size]
				b = b[4+size:]
			}
			wordData = append(wordData, &Data{
				t:    t,
				data: data,
			})
		}
	}

	return &Word{
		data: wordData,
	}, nil
}

// Close closes the dict file.
func (d *Dict) Close() error {
	return d.r.Close()
}

// getDictBytes reads bytes from the underlying readers.
func (d *Dict) getDictBytes(e *idx.Word) ([]byte, error) {
	if d.dz != nil {
		return d.dz.Get(int64(e.Offset), int64(e.Size))
	}

	d.r.Seek(int64(e.Offset), io.SeekStart)
	b := make([]byte, e.Size)
	_, err := io.ReadFull(d.r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
