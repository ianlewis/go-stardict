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

// Package dict implements reading .dict files.
package dict

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/ianlewis/go-stardict/idx"
)

var (
	errInvalidType        = errors.New("invalid type")
	errWordOffsetTooLarge = errors.New("word offset too large")
)

// Dict represents a Stardict dictionary's dictionary data.
type Dict struct {
	r                io.ReaderAt
	sametypesequence []DataType
}

// Word is a full dictionary entry.
type Word struct {
	Data []*Data
}

// DataType is a type of data in a word. Data types are specified by a single
// byte at the beginning of a word. Lower case characters represent string-like
// data that is terminated by a null terminator ('\0'). Upper case characters
// represent file-like data that starts with a 32-bit size followed by file
// data.
type DataType byte

const (
	// UTFTextType is utf-8 text.
	UTFTextType = DataType('m')

	// LocaleTextType is text in a locale encoding.
	LocaleTextType = DataType('l')

	// PangoTextType is utf-8 text in the Pango text format.
	PangoTextType = DataType('g')

	// PhoneticType is utf-8 text representing an English phonetic string.
	PhoneticType = DataType('t')

	// XDXFType is utf-8 encoded xml in XDXF format.
	XDXFType = DataType('x')

	// YinBiaoOrKataType is utf-8 encoded Yin Biao or Kana phonetic string.
	YinBiaoOrKataType = DataType('y')

	// PowerWordType is a utf-8 encoded KingSoft PowerWord XML format.
	PowerWordType = DataType('p')

	// MediaWikiType is utf-8 encoded text in MediaWiki format.
	MediaWikiType = DataType('w')

	// HTMLType is utf-8 encoded HTML text.
	HTMLType = DataType('h')

	// WordNetType is WordNet data.
	WordNetType = DataType('n')

	// ResourceFileListType is a list of files in resource storage.
	ResourceFileListType = DataType('r')

	// WavType is .wav sound file data.
	WavType = DataType('W')

	// PictureType is image file data. This was used by the
	// stardict-advertisement-plugin. Images are better stored in a resource
	// file list.
	PictureType = DataType('P')

	// ExperimentalType is reserved for experimental features.
	ExperimentalType = DataType('X')
)

// Data is a data entry in a Word.
type Data struct {
	Type DataType
	Data []byte
}

// New returns a new Dict from the given reader. Dict takes ownership of the
// reader. The reader can be closed via the Dict's Close method.
func New(r io.ReaderAt, sametypesequence []DataType) (*Dict, error) {
	// verify sametypesequence
	for _, s := range sametypesequence {
		switch s {
		case UTFTextType,
			LocaleTextType,
			PangoTextType,
			PhoneticType,
			XDXFType,
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
			return nil, fmt.Errorf("%w: %v", errInvalidType, s)
		}
	}

	return &Dict{
		r:                r,
		sametypesequence: sametypesequence,
	}, nil
}

// Word retrieves the word for the given index entry from the
// dictionary.
func (d *Dict) Word(e *idx.Word) (*Word, error) {
	b := make([]byte, e.Size)
	// TODO(#9): Support dictionary word offsets math.MaxInt64 > x < math.MaxUint64
	if e.Offset > math.MaxInt64 {
		return nil, fmt.Errorf("%w: %d", errWordOffsetTooLarge, e.Offset)
	}
	// NOTE: if ReadAt does not read e.Size bytes then an error should be
	// returned.
	//nolint:gosec // offset size is bounds checked above.
	_, err := d.r.ReadAt(b, int64(e.Offset))
	if err != nil {
		return nil, fmt.Errorf("reading dictionary: %w", err)
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
					i++
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
				Type: t,
				Data: data,
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
				Type: t,
				Data: data,
			})
		}
	}

	return &Word{
		Data: wordData,
	}, nil
}
