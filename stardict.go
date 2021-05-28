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
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const ifoMagic = "StarDict's dict ifo file"

// Stardict is a stardict dictionary.
type Stardict struct {
	ifo  *Ifo
	idx  *Idx
	dict *Dict

	version          string
	bookname         string
	wordcount        int64
	synwordcount     int64
	idxfilesize      int64
	idxoffsetbits    int64
	author           string
	email            string
	website          string
	description      string
	sametypesequence []WordType
}

// OpenAll opens all dictionaries under a directory. This function will return
// all successfully opened dictionaries along with any errors that occurred.
func OpenAll(path string) ([]*Stardict, []error) {
	var dicts []*Stardict
	var errs []error
	if err := filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		// Walking the file path will ignore errors.
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if !info.IsDir() && (filepath.Ext(info.Name()) == ".ifo" || filepath.Ext(info.Name()) == ".IFO") {
			dict, err := Open(path)
			if err != nil {
				errs = append(errs, err)
				return nil
			}
			dicts = append(dicts, dict)
		}
		return nil
	}); err != nil {
		errs = append(errs, err)
		return nil, errs
	}
	return dicts, errs
}

// Open opens a Stardict dictionary from the given .ifo file path.
func Open(path string) (*Stardict, error) {
	ifoExt := filepath.Ext(path)
	baseName := strings.TrimSuffix(path, ifoExt)
	if ifoExt != ".ifo" && ifoExt != ".IFO" {
		return nil, fmt.Errorf("bad extension: %v", ifoExt)
	}

	ifoFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %w", path, err)
	}
	defer ifoFile.Close()

	ifo, err := NewIfo(ifoFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", path, err)
	}

	if ifo.Magic() != ifoMagic {
		return nil, fmt.Errorf("%q bad magic data", path)
	}

	s := &Stardict{
		ifo:           ifo,
		idxoffsetbits: 32,
	}

	// Validate the version
	s.version = ifo.Value("version")
	switch s.version {
	case "2.4.2":
	case "3.0.0":
	default:
		return nil, fmt.Errorf("invalid version: %v", s.version)
	}

	s.bookname = ifo.Value("bookname")
	if s.bookname == "" {
		return nil, fmt.Errorf("missing bookname")
	}

	s.wordcount, err = strconv.ParseInt(ifo.Value("wordcount"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad wordcount: %w", err)
	}

	s.idxfilesize, err = strconv.ParseInt(ifo.Value("idxfilesize"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad idxfilesize: %w", err)
	}

	idxoffsetbits := ifo.Value("idxoffsetbits")
	if idxoffsetbits != "" && s.version == "3.0.0" {
		s.idxoffsetbits, err = strconv.ParseInt(idxoffsetbits, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid idxoffsetbits: %w", err)
		}
	}

	synwordcount := ifo.Value("synwordcount")
	if synwordcount != "" {
		s.synwordcount, err = strconv.ParseInt(synwordcount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad synwordcount: %w", err)
		}
	}

	sametypesequence := ifo.Value("sametypesequence")
	if sametypesequence != "" {
		for _, r := range sametypesequence {
			s.sametypesequence = append(s.sametypesequence, WordType(r))
		}
	}

	s.author = ifo.Value("author")
	s.email = ifo.Value("email")
	s.description = ifo.Value("description")
	s.website = ifo.Value("website")

	// TODO: .syn file.

	// Open the index file.
	s.idx, err = openIdx(baseName, s.idxoffsetbits)
	if err != nil {
		return nil, err
	}

	// Open the dict file.
	s.dict, err = openDict(baseName, s.sametypesequence)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Bookname returns the dictionary name.
func (s *Stardict) Bookname() string {
	return s.bookname
}

// Description returns the dictionary description.
func (s *Stardict) Description() string {
	return s.description
}

// Author returns the dictionary author.
func (s *Stardict) Author() string {
	return s.author
}

// Email returns the dictionary contact email.
func (s *Stardict) Email() string {
	return s.email
}

// Website returns the dictionary website url.
func (s *Stardict) Website() string {
	return s.website
}

// WordCount returns the dictionary word count.
func (s *Stardict) WordCount() int64 {
	return s.wordcount
}

// Version returns the dictionary format version.
func (s *Stardict) Version() string {
	return s.version
}

// Index returns the dictionary's index.
func (s *Stardict) Index() *Idx {
	return s.idx
}

// Index returns the dictionary's index.
func (s *Stardict) Article(e *Entry) (Article, error) {
	return s.dict.Article(e)
}

// Close closes all dictionary resources.
func (s *Stardict) Close() error {
	return s.idx.Close()
}

func openIdx(baseName string, idxoffsetbits int64) (*Idx, error) {
	// TODO: .syn file.
	idxExts := []string{".idx.gz", ".idx", ".IDX", ".IDX.gz", ".IDX.GZ"}
	var idxPath string
	for _, ext := range idxExts {
		idxPath = baseName + ext
		if _, err := os.Stat(idxPath); err == nil {
			break
		}
	}
	if idxPath == "" {
		return nil, fmt.Errorf("no index found")
	}

	var r io.ReadCloser
	var err error
	r, err = os.Open(idxPath)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %w", idxPath, err)
	}

	idxExt := filepath.Ext(idxPath)
	if idxExt == ".gz" || idxExt == ".GZ" {
		r, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("error opening %q: %w", idxPath, err)
		}
	}

	idx, err := NewIdx(r, idxoffsetbits)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", idxPath, err)
	}
	return idx, nil
}

func openDict(baseName string, sametypesequence []WordType) (*Dict, error) {
	dictExts := []string{".dict.dz", ".dict", ".DICT", ".DICT.dz", ".DICT.DZ"}
	var dictPath string
	for _, ext := range dictExts {
		dictPath = baseName + ext
		if _, err := os.Stat(dictPath); err == nil {
			break
		}
	}
	if dictPath == "" {
		return nil, fmt.Errorf("no dict found")
	}

	var r io.ReadSeekCloser
	var err error
	r, err = os.Open(dictPath)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %w", dictPath, err)
	}

	// TODO: Support .dz
	// dictExt := filepath.Ext(dictPath)
	// if dictExt == ".dz" || dictExt == ".DZ" {
	// 	r, err = gzip.NewReader(r)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error opening %q: %w", dictPath, err)
	// 	}
	// }

	dict, err := NewDict(r, sametypesequence)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", dictPath, err)
	}
	return dict, nil

}
