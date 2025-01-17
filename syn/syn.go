// Copyright 2025 Ian Lewis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syn

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/ianlewis/go-stardict/internal/folding"
)

// Word is a .syn file entry.
type Word struct {
	// Word is the synonym word.
	Word string

	// OriginalWordIndex is the index into the .idx index.
	OriginalWordIndex uint32
}

type foldedWord struct {
	folded string
	word   *Word
}

// Syn is is the synonym index. It is largely a map of synonym words to related
// index entries.
type Syn struct {
	folded          []*foldedWord
	foldTransformer transform.Transformer
}

// New returns a new Syn by reading the data from r.
func New(r io.ReadCloser) (*Syn, error) {
	syn := Syn{
		foldTransformer: transform.Chain(
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

	i := 0
	s, err := NewScanner(r)
	if err != nil {
		return nil, fmt.Errorf("creating synonym index scanner: %w", err)
	}
	for s.Scan() {
		word := s.Word()
		folded, _, err := transform.String(syn.foldTransformer, word.Word)
		if err != nil {
			return nil, fmt.Errorf("folding word %q: %w", word.Word, err)
		}

		syn.folded = append(syn.folded, &foldedWord{
			folded: folded,
			word:   word,
		})
		i++
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning synonym index %w", err)
	}

	// We need to re-sort based on the folded word.
	sort.Slice(syn.folded, func(i, j int) bool {
		return syn.folded[i].folded < syn.folded[j].folded
	})

	return &syn, nil
}

// NewFromIfoPath returns a new in-memory index.
func NewFromIfoPath(ifoPath string) (*Syn, error) {
	f, err := openSynFile(ifoPath)
	if err != nil {
		return nil, err
	}
	return New(f)
}

func openSynFile(ifoPath string) (*os.File, error) {
	baseName := strings.TrimSuffix(ifoPath, filepath.Ext(ifoPath))

	synExts := []string{
		".syn",
		".SYN",
	}
	var f *os.File
	var err error
	for _, ext := range synExts {
		f, err = os.Open(baseName + ext)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("opening .syn file: %w", err)
		}
	}

	// Catch the case when no .syn file was found.
	if err != nil {
		return nil, fmt.Errorf("opening .syn file: %w", err)
	}

	return f, nil
}

// Search performs a query of the index and returns matching words.
func (syn *Syn) Search(query string) ([]*Word, error) {
	foldedQuery, _, err := transform.String(syn.foldTransformer, query)
	if err != nil {
		return nil, fmt.Errorf("folding query %q: %w", query, err)
	}

	start := 0
	end := len(syn.folded) - 1
	for start <= end {
		pivot := (start + end) / 2
		switch {
		case syn.folded[pivot].folded < foldedQuery:
			start = pivot + 1
		case syn.folded[pivot].folded > foldedQuery:
			end = pivot - 1
		default:
			// We have found a matching word.
			// Multiple word entries may have the same value we must find the
			// first and iterate over the index until we have found all matches.
			i := pivot
			for i > 0 && syn.folded[i-1].folded == foldedQuery {
				i--
			}
			j := pivot
			for j+1 < len(syn.folded) && syn.folded[j+1].folded == foldedQuery {
				j++
			}

			var result []*Word
			for ; i < j+1; i++ {
				result = append(result, syn.folded[i].word)
			}

			return result, nil
		}
	}

	return nil, nil
}
