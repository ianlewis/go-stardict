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
	"fmt"
	"io"

	"github.com/pebbe/dictzip"
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

// NewDict returns a new Dict.
func NewDict(r io.ReadSeekCloser, sametypesequence []DataType, isDictZip bool) (*Dict, error) {
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
func (d *Dict) Word(e *Entry) (*Word, error) {
	b, err := d.getDictBytes(e)
	if err != nil {
		return nil, err
	}

	var wordData []*Data
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Split(d.splitWord)
	for i := 0; s.Scan(); i++ {
		token := s.Bytes()
		var t DataType
		var data []byte
		if len(d.sametypesequence) > 0 {
			if i >= len(d.sametypesequence) {
				return nil, fmt.Errorf("invalid word data")
			}
			t = d.sametypesequence[i]
			data = token
		} else {
			t = DataType(token[0])
			data = token[1:]
		}

		wordData = append(wordData, &Data{
			t:    t,
			data: data,
		})
	}

	if err := s.Err(); err != nil {
		return nil, err
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
func (d *Dict) getDictBytes(e *Entry) ([]byte, error) {
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

// splitWord splits an article by word.
func (d *Dict) splitWord(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	wordData := data
	if len(d.sametypesequence) == 0 {
		wordData = data[1:]
	}
	if i := bytes.IndexByte(wordData, 0); i >= 0 {
		// Found zero byte.
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
