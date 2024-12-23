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
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Word is an .idx file entry.
type Word struct {
	// Word is the word as it appears in the index.
	Word string

	// Offset is the offset in the .dict file that the corresponding entry appears.
	Offset uint64

	// Size is the total size of the corresponding .dict file entry.
	Size uint32
}

type foldedWord struct {
	folded string
	word   *Word
}

// whitespaceFolder will perform whitespace folding on the input. It removes
// spaces from the beginning and end of the input and replaces all internal
// whitespace spans with a single ASCII space rune.
type whitespaceFolder struct {
	// notStart is true after encounting the first non-whitespace rune.
	notStart bool

	// wsSpan is true if the transformer is currently handling a whitespace span.
	wsSpan bool
}

// Transform implements [transform.Transformer.Transform].
func (w *whitespaceFolder) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	var nSrc, nDst int
	for nSrc < len(src) {
		c, size := utf8.DecodeRune(src[nSrc:])
		if c == utf8.RuneError && !atEOF {
			return nDst, nSrc, transform.ErrShortSrc
		}

		isSpace := unicode.IsSpace(c)
		if isSpace {
			nSrc += size
			if !w.notStart {
				// Ignore leading whitespace.
				continue
			}
			// We are in an internal whitespace span.
			w.wsSpan = true
			continue
		}

		if w.wsSpan {
			// Emit a single space if we are coming out of a whitespace span.
			// NOTE: trailing whitespace is never emitted by design.
			spc := ' '
			if nDst+utf8.RuneLen(spc) > len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			nDst += utf8.EncodeRune(dst[nDst:], spc)
			// We are no longer in an internal whitespace span.
			w.wsSpan = false
		}
		w.notStart = true
		nSrc += size

		// Emit the character.
		// NOTE: we cannot use size here because c could be utf8.RuneError in
		// which case size would be 1 but the length of utf8.RuneError is 3.
		if nDst+utf8.RuneLen(c) > len(dst) {
			return nDst, nSrc, transform.ErrShortDst
		}
		nDst += utf8.EncodeRune(dst[nDst:], c)
	}

	return nDst, nSrc, nil
}

// Reset implements [transform.Transformer.Reset].
func (w *whitespaceFolder) Reset() {
	*w = whitespaceFolder{}
}

// Options are options for the idx data.
type Options struct {
	// OffsetBits are the number of bits in the offset fields.
	OffsetBits int
}

// DefaultOptions is the default options for an Idx.
var DefaultOptions = &Options{
	OffsetBits: 32,
}

// Idx is a very basic implementation of an in memory search index.
// Implementers of dictionaries apps or tools may wish to consider using
// Scanner to read the .idx file and generate their own more robust search
// index.
type Idx struct {
	words           []*foldedWord
	foldTransformer transform.Transformer
}

// New returns a new in-memory index.
func New(r io.ReadCloser, options *Options) (*Idx, error) {
	if options == nil {
		options = DefaultOptions
	}

	idx := &Idx{
		foldTransformer: transform.Chain(
			// Unicode Normalization Form D (Canonical Decomposition.
			norm.NFD,
			// Perform case folding.
			cases.Fold(),
			// Perform whitespace folding.
			&whitespaceFolder{},
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
	s, err := NewScanner(r, options)
	if err != nil {
		return nil, fmt.Errorf("creating index scanner: %w", err)
	}
	for s.Scan() {
		word := s.Word()
		folded, _, err := transform.String(idx.foldTransformer, word.Word)
		if err != nil {
			return nil, fmt.Errorf("folding word %q: %w", word.Word, err)
		}

		idx.words = append(idx.words, &foldedWord{
			folded: folded,
			word:   word,
		})
		i++
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning index %w", err)
	}

	// We need to re-sort based on the folded word.
	sort.Slice(idx.words, func(i, j int) bool {
		return idx.words[i].folded < idx.words[j].folded
	})

	return idx, nil
}

func openIdxFile(ifoPath string) (*os.File, error) {
	baseName := strings.TrimSuffix(ifoPath, filepath.Ext(ifoPath))

	idxExts := []string{
		".idx",
		".idx.gz",
		".idx.GZ",
		".idx.dz",
		".idx.DZ",
		".IDX",
		".IDX.gz",
		".IDX.GZ",
		".IDX.dz",
		".IDX.DZ",
	}
	var f *os.File
	var err error
	for _, ext := range idxExts {
		f, err = os.Open(baseName + ext)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("opening .idx file: %w", err)
		}
	}

	// Catch the case when no .idx file was found.
	if err != nil {
		return nil, fmt.Errorf("opening .idx file: %w", err)
	}

	return f, nil
}

// NewFromIfoPath returns a new in-memory index.
func NewFromIfoPath(ifoPath string, options *Options) (*Idx, error) {
	var r io.ReadCloser
	f, err := openIdxFile(ifoPath)
	if err != nil {
		return nil, err
	}
	r = f

	idxExt := strings.ToLower(filepath.Ext(f.Name()))
	if idxExt == ".gz" || idxExt == ".dz" {
		r, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("creating .ifo gzip reader: %w", err)
		}
	}

	return New(r, options)
}

// Search performs a query of the index and returns matching words.
func (idx *Idx) Search(query string) ([]*Word, error) {
	foldedQuery, _, err := transform.String(idx.foldTransformer, query)
	if err != nil {
		return nil, fmt.Errorf("folding query %q: %w", query, err)
	}

	start := 0
	end := len(idx.words) - 1
	for start <= end {
		pivot := (start + end) / 2
		switch {
		case idx.words[pivot].folded < foldedQuery:
			start = pivot + 1
		case idx.words[pivot].folded > foldedQuery:
			end = pivot - 1
		default:
			// We have found a matching word.
			// Multiple word entries may have the same value we must find the
			// first and iterate over the index until we have found all matches.
			i := pivot
			for i > 0 && idx.words[i-1].folded == foldedQuery {
				i--
			}
			j := pivot
			for j+1 < len(idx.words) && idx.words[j+1].folded == foldedQuery {
				j++
			}

			var result []*Word
			for ; i < j+1; i++ {
				result = append(result, idx.words[i].word)
			}

			return result, nil
		}
	}

	return nil, nil
}
