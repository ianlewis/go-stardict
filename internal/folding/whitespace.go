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

package folding

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/transform"
)

// WhitespaceFolder will perform whitespace folding on the input. It removes
// spaces from the beginning and end of the input and replaces all internal
// whitespace spans with a single ASCII space rune.
type WhitespaceFolder struct {
	// notStart is true after encounting the first non-whitespace rune.
	notStart bool

	// wsSpan is true if the transformer is currently handling a whitespace span.
	wsSpan bool
}

// Transform implements [transform.Transformer.Transform].
func (w *WhitespaceFolder) Transform(dst, src []byte, atEOF bool) (int, int, error) {
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
func (w *WhitespaceFolder) Reset() {
	*w = WhitespaceFolder{}
}
