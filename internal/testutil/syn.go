// Copyright 2024 Google LLC
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

package testutil

import (
	"encoding/binary"
	"testing"

	"github.com/ianlewis/go-stardict/syn"
)

// MakeSyn make a test index given a list of words.
func MakeSyn(t *testing.T, words []*syn.Word) []byte {
	t.Helper()

	b := []byte{}
	for _, w := range words {
		// Allow for zero byte terminator and 4 extra bytes for the word index.
		bWord := make([]byte, len(w.Word)+5)
		copy(bWord, w.Word)
		i := len(w.Word)
		// NOTE: byte array entries are already initialized to zero but we set
		// it anyway for clarity.
		bWord[i] = 0
		i++

		binary.BigEndian.PutUint32(bWord[i:], w.OriginalWordIndex)
		b = append(b, bWord...)
	}
	return b
}
