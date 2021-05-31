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

	"github.com/pebbe/dictzip"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/ifo"
)

const ifoMagic = "StarDict's dict ifo file"

// Stardict is a stardict dictionary.
type Stardict struct {
	ifo  *ifo.Ifo
	idx  *idx.Idx
	dict *dict.Dict

	dictFile *os.File

	ifoPath string

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
	sametypesequence []dict.DataType
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
	s := &Stardict{
		ifoPath:       path,
		idxoffsetbits: 32,
	}

	ifoExt := filepath.Ext(s.ifoPath)
	if ifoExt != ".ifo" && ifoExt != ".IFO" {
		return nil, fmt.Errorf("bad extension: %v", ifoExt)
	}

	ifoFile, err := os.Open(s.ifoPath)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %w", s.ifoPath, err)
	}
	defer ifoFile.Close()

	s.ifo, err = ifo.New(ifoFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", s.ifoPath, err)
	}

	if s.ifo.Magic() != ifoMagic {
		return nil, fmt.Errorf("%q bad magic data", s.ifoPath)
	}

	// Validate the version
	s.version = s.ifo.Value("version")
	switch s.version {
	case "2.4.2":
	case "3.0.0":
	default:
		return nil, fmt.Errorf("invalid version: %v", s.version)
	}

	s.bookname = s.ifo.Value("bookname")
	if s.bookname == "" {
		return nil, fmt.Errorf("missing bookname")
	}

	s.wordcount, err = strconv.ParseInt(s.ifo.Value("wordcount"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad wordcount: %w", err)
	}

	s.idxfilesize, err = strconv.ParseInt(s.ifo.Value("idxfilesize"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad idxfilesize: %w", err)
	}

	idxoffsetbits := s.ifo.Value("idxoffsetbits")
	if idxoffsetbits != "" && s.version == "3.0.0" {
		s.idxoffsetbits, err = strconv.ParseInt(idxoffsetbits, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid idxoffsetbits: %w", err)
		}
	}

	synwordcount := s.ifo.Value("synwordcount")
	if synwordcount != "" {
		s.synwordcount, err = strconv.ParseInt(synwordcount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad synwordcount: %w", err)
		}
	}

	sametypesequence := s.ifo.Value("sametypesequence")
	if sametypesequence != "" {
		for _, r := range sametypesequence {
			s.sametypesequence = append(s.sametypesequence, dict.DataType(r))
		}
	}

	s.author = s.ifo.Value("author")
	s.email = s.ifo.Value("email")
	s.description = s.ifo.Value("description")
	s.website = s.ifo.Value("website")

	// TODO: support the .syn file.
	// TODO: support the .tdx file.
	// TODO: support resource storage

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

// Search performs a simple full text search of the dictionary for the
// given query and returns dictionary entries.
func (s *Stardict) Search(query string) ([]*Entry, error) {
	idx, err := s.Index()
	if err != nil {
		return nil, err
	}
	dict, err := s.Dict()
	if err != nil {
		return nil, err
	}

	var entries []*Entry
	for _, idxWord := range idx.FullTextSearch(query) {
		dictWord, err := dict.Word(idxWord)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &Entry{
			word: idxWord.Word,
			data: dictWord.Data,
		})
	}
	return entries, nil
}

// Index returns a simple in-memory version of the dictionary's index.
func (s *Stardict) Index() (*idx.Idx, error) {
	if s.idx != nil {
		return s.idx, nil
	}
	if err := s.openIdx(); err != nil {
		return nil, err
	}
	return s.idx, nil
}

// Dict returns the dictionary's dict.
func (s *Stardict) Dict() (*dict.Dict, error) {
	if s.dict != nil {
		return s.dict, nil
	}
	// Open the dict file.
	if err := s.openDict(); err != nil {
		return nil, err
	}
	return s.dict, nil
}

// Close closes the dict and any underlying readers.
func (s *Stardict) Close() error {
	if s.dictFile != nil {
		return s.dictFile.Close()
	}
	return nil
}

func (s *Stardict) openIdx() error {
	ifoExt := filepath.Ext(s.ifoPath)
	baseName := strings.TrimSuffix(s.ifoPath, ifoExt)

	idxExts := []string{".idx.gz", ".idx", ".IDX", ".IDX.gz", ".IDX.GZ"}
	var idxPath string
	for _, ext := range idxExts {
		idxPath = baseName + ext
		if _, err := os.Stat(idxPath); err == nil {
			break
		}
	}
	if idxPath == "" {
		return fmt.Errorf("no index found")
	}

	var r io.ReadCloser
	var err error
	r, err = os.Open(idxPath)
	if err != nil {
		return fmt.Errorf("error opening %q: %w", idxPath, err)
	}
	defer r.Close()

	idxExt := filepath.Ext(idxPath)
	if strings.ToLower(idxExt) == ".gz" {
		r, err = gzip.NewReader(r)
		if err != nil {
			return fmt.Errorf("error opening %q: %w", idxPath, err)
		}
		defer r.Close()
	}

	idx, err := idx.New(r, s.idxoffsetbits)
	if err != nil {
		return fmt.Errorf("error reading %q: %w", idxPath, err)
	}

	s.idx = idx

	return nil
}

func (s *Stardict) openDict() error {
	ifoExt := filepath.Ext(s.ifoPath)
	baseName := strings.TrimSuffix(s.ifoPath, ifoExt)

	dictExts := []string{".dict.dz", ".dict", ".DICT", ".DICT.dz", ".DICT.DZ"}
	var dictPath string
	for _, ext := range dictExts {
		dictPath = baseName + ext
		if _, err := os.Stat(dictPath); err == nil {
			break
		}
	}
	if dictPath == "" {
		return fmt.Errorf("no dict found")
	}

	var r io.ReaderAt
	f, err := os.Open(dictPath)
	if err != nil {
		return fmt.Errorf("error opening %q: %w", dictPath, err)
	}
	r = f

	var dz *dictzip.Reader
	dictExt := filepath.Ext(dictPath)
	if strings.ToLower(dictExt) == ".dz" {
		dz, err = dictzip.NewReader(f)
		if err != nil {
			return err
		}
		r = dz
	}

	dict, err := dict.New(r, s.sametypesequence)
	if err != nil {
		return fmt.Errorf("error reading %q: %w", dictPath, err)
	}

	s.dict = dict
	s.dictFile = f

	return nil
}
