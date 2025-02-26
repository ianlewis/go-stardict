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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/ianlewis/go-stardict/idx"
	"github.com/ianlewis/go-stardict/ifo"
	"github.com/ianlewis/go-stardict/internal/folding"
	"github.com/ianlewis/go-stardict/syn"
)

const ifoMagic = "StarDict's dict ifo file"

// Stardict is a stardict dictionary.
type Stardict struct {
	ifo  *ifo.Ifo
	idx  *idx.Idx
	syn  *syn.Syn
	dict *dict.Dict

	dictFile *os.File

	ifoPath string

	version          string
	bookname         string
	wordcount        int64
	synwordcount     int64
	idxfilesize      int64
	idxoffsetbits    int
	author           string
	email            string
	website          string
	description      string
	sametypesequence []dict.DataType

	folder transform.Transformer
}

// Options are options for the Stardict dictionary.
type Options struct {
	// Folder performs folding (e.g. case folding, whitespace folding, etc.) on dictionary entries.
	Folder transform.Transformer
}

var (
	errNoBookname     = errors.New("missing bookname")
	errInvalidVersion = errors.New("invalid version")
	errInvalidMagic   = errors.New("invalid magic data")
	errIfoExtension   = errors.New("invalid .ifo file extension")
)

// OpenAll opens all dictionaries under a directory. This function will return
// all successfully opened dictionaries along with any errors that occurred.
func OpenAll(path string, options *Options) ([]*Stardict, []error) {
	var dicts []*Stardict
	var errs []error
	if err := filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		// Walking the file path will ignore errors.
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if !info.IsDir() && (filepath.Ext(info.Name()) == ".ifo" || filepath.Ext(info.Name()) == ".IFO") {
			dict, err := Open(path, options)
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
func Open(path string, options *Options) (*Stardict, error) {
	if options == nil {
		options = &Options{
			Folder: transform.Chain(
				// Unicode Normalization Form D (Canonical Decomposition.
				norm.NFD,
				// Perform case folding.
				cases.Fold(),
				// Perform whitespace folding.
				&folding.WhitespaceFolder{},
				// Remove Non-spacing marks ([, ] {, }, etc.).
				runes.Remove(runes.In(unicode.Mn)),
				// Remove punctuation.
				runes.Remove(runes.In(unicode.P)),
				// Unicode Normalization Form C (Canonical Decomposition, followed by Canonical Composition)
				// NOTE: Case folding does not normalize the input and may not
				// preserve a normal form. Canonical Decomposition is thus necessary
				// to be performed a second time.
				norm.NFC,
			),
		}
	}

	s := &Stardict{
		ifoPath:       path,
		idxoffsetbits: 32,
	}

	s.folder = transform.Nop
	if options.Folder != nil {
		s.folder = options.Folder
	}

	ifoExt := filepath.Ext(s.ifoPath)
	if ifoExt != ".ifo" && ifoExt != ".IFO" {
		return nil, fmt.Errorf("%w: %v", errIfoExtension, ifoExt)
	}

	ifoFile, err := os.Open(s.ifoPath)
	if err != nil {
		return nil, fmt.Errorf("opening %q: %w", s.ifoPath, err)
	}
	defer ifoFile.Close()

	s.ifo, err = ifo.New(ifoFile)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", s.ifoPath, err)
	}

	if s.ifo.Magic() != ifoMagic {
		return nil, fmt.Errorf("%w: %q", errInvalidMagic, s.ifoPath)
	}

	// Validate the version
	s.version = s.ifo.Value("version")
	switch s.version {
	case "2.4.2":
	case "3.0.0":
	default:
		return nil, fmt.Errorf("%w: %v", errInvalidVersion, s.version)
	}

	s.bookname = s.ifo.Value("bookname")
	if s.bookname == "" {
		return nil, errNoBookname
	}

	s.wordcount, err = strconv.ParseInt(s.ifo.Value("wordcount"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid workdcount: %w", err)
	}

	s.idxfilesize, err = strconv.ParseInt(s.ifo.Value("idxfilesize"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid idxfilesize: %w", err)
	}

	idxoffsetbits := s.ifo.Value("idxoffsetbits")
	// TODO(#28): version check should be > 3.0.0
	if idxoffsetbits != "" && s.version == "3.0.0" {
		var idxoffsetbits64 int64
		idxoffsetbits64, err = strconv.ParseInt(idxoffsetbits, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid idxoffsetbits: %w", err)
		}
		s.idxoffsetbits = int(idxoffsetbits64)
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
	s.description = strings.ReplaceAll(s.ifo.Value("description"), "<br>", "\n")
	s.website = s.ifo.Value("website")

	// TODO(#3): support the .tdx file.
	// TODO(#4): support resource storage
	// TODO(#7): support collation files.
	// TODO(#8): support offset cache files.

	return s, nil
}

// Bookname returns the dictionary name.
func (s *Stardict) Bookname() string {
	return s.bookname
}

// Description returns the dictionary description. This field is optional for
// dictionaries. <br> has been replaced with '\n'.
func (s *Stardict) Description() string {
	return s.description
}

// Author returns the dictionary author. This field is optional for
// dictionaries.
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

// SynWordCount returns the wordcount in the synonym index.
func (s *Stardict) SynWordCount() int64 {
	return s.synwordcount
}

// Version returns the dictionary format version.
func (s *Stardict) Version() string {
	return s.version
}

// Search performs a simple full text search of the dictionary for the
// given query and returns dictionary entries.
func (s *Stardict) Search(query string) ([]*Entry, error) {
	var entries []*Entry

	// Read entries from the index.
	index, err := s.Index()
	if err != nil {
		return nil, err
	}
	idxResults, err := index.Search(query)
	if err != nil {
		return nil, fmt.Errorf("searching index: %w", err)
	}

	// Read the entries from the dict.
	d, err := s.Dict()
	if err != nil {
		return nil, err
	}
	for _, idxWord := range idxResults {
		dictWord, err := d.Word(idxWord)
		if err != nil {
			return nil, fmt.Errorf("reading word: %w", err)
		}
		entries = append(entries, &Entry{
			word: idxWord.Word,
			data: dictWord.Data,
		})
	}
	return entries, nil
}

// IndexScanner returns a new index scanner. The caller assumes ownership of
// the underlying reader so Close should be called on the scanner when
// finished.
func (s *Stardict) IndexScanner() (*idx.Scanner, error) {
	sc, err := idx.NewScannerFromIfoPath(s.ifoPath, &idx.ScannerOptions{
		OffsetBits: s.idxoffsetbits,
	})
	if err != nil {
		return nil, fmt.Errorf("creating index scanner: %w", err)
	}
	return sc, nil
}

// Index returns a simple in-memory version of the dictionary's index.
func (s *Stardict) Index() (*idx.Idx, error) {
	if s.idx != nil {
		return s.idx, nil
	}

	// Open the .idx file.
	index, err := idx.NewFromIfoPath(s.ifoPath, &idx.Options{
		Folder: s.folder,
		ScannerOptions: &idx.ScannerOptions{
			OffsetBits: s.idxoffsetbits,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("opening index: %w", err)
	}
	s.idx = index

	return s.idx, nil
}

// Syn returns a simple in-memory version of the dictionary's synonym index.
func (s *Stardict) Syn() (*syn.Syn, error) {
	if s.syn != nil {
		return s.syn, nil
	}

	// Open the .syn file.
	synIndex, err := syn.NewFromIfoPath(s.ifoPath, &syn.Options{
		Folder: s.folder,
	})
	if err != nil {
		return nil, fmt.Errorf("opening synonym index: %w", err)
	}
	s.syn = synIndex

	return s.syn, nil
}

// Dict returns the dictionary's dict.
func (s *Stardict) Dict() (*dict.Dict, error) {
	if s.dict != nil {
		return s.dict, nil
	}

	// Open the .dict file.
	d, err := dict.NewFromIfoPath(s.ifoPath, &dict.Options{
		SameTypeSequence: s.sametypesequence,
	})
	if err != nil {
		return nil, fmt.Errorf("opening dict: %w", err)
	}
	s.dict = d

	return s.dict, nil
}

// Close closes the dict and any underlying readers.
func (s *Stardict) Close() error {
	if s.dictFile != nil {
		if err := s.dictFile.Close(); err != nil {
			return fmt.Errorf("closing dict file: %w", err)
		}
	}
	return nil
}
