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
)

// Dict represents a Stardict dictionary's dictionary data.
type Dict struct {
	r                io.ReadSeekCloser
	sametypesequence []WordType
}

type Article []Word

type WordType byte

const (
	UTFTextType          = WordType('m')
	LocaleTextType       = WordType('l')
	PangoTextType        = WordType('g')
	PhoneticType         = WordType('t')
	XdxfType             = WordType('x')
	YinBiaoOrKataType    = WordType('y')
	PowerWordType        = WordType('p')
	MediaWikiType        = WordType('w')
	HTMLType             = WordType('h')
	WordNetType          = WordType('n')
	ResourceFileListType = WordType('r')
	WavType              = WordType('W')
	PictureType          = WordType('P')
	ExperimentalType     = WordType('X')
)

type Word struct {
	t    WordType
	data []byte
}

func (w Word) Type() WordType {
	return w.t
}

func (w Word) Data() []byte {
	return w.data
}

// NewDict returns a new Dict.
func NewDict(r io.ReadSeekCloser, sametypesequence []WordType) (*Dict, error) {
	// verify sametypesequence
	for _, s := range sametypesequence {
		switch s {
		case UTFTextType,
			LocaleTextType,
			PangoTextType,
			PhoneticType,
			XdxfType,
			YinBiaoOrKataType,
			PowerWordType,
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

	return &Dict{
		r:                r,
		sametypesequence: sametypesequence,
	}, nil
}

// Article retrieves the article for the given index entry from the
// dictionary.
func (d *Dict) Article(e *Entry) (Article, error) {
	d.r.Seek(int64(e.Offset), io.SeekStart)
	b := make([]byte, e.Size)
	_, err := io.ReadFull(d.r, b)
	if err != nil {
		return nil, err
	}

	var a []Word
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Split(d.splitWord)
	for i := 0; s.Scan(); i++ {
		token := s.Bytes()
		var t WordType
		var data []byte
		if len(d.sametypesequence) > 0 {
			if i >= len(d.sametypesequence) {
				return nil, fmt.Errorf("invalid article data")
			}
			t = d.sametypesequence[i]
			data = token
		} else {
			t = WordType(token[0])
			data = token[1:]
		}

		a = append(a, Word{
			t:    t,
			data: data,
		})
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	return Article(a), nil
}

// Close closes the dict file.
func (d *Dict) Close() error {
	return d.r.Close()
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
