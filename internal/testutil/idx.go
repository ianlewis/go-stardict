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
	"fmt"
	"math"

	"github.com/ianlewis/go-stardict/idx"
)

// MakeIndex make a test index given a list of words.
func MakeIndex(words []*idx.Word, idxoffsetbits int64) []byte {
	b := []byte{}
	for _, w := range words {
		b = append(b, []byte(w.Word)...)
		b = append(b, 0) // Add the zero byte terminator.
		var b2 []byte
		switch idxoffsetbits {
		case 32:
			b2 = make([]byte, 8)
			if w.Offset > math.MaxUint32 {
				panic(fmt.Sprintf("word offset too large %d > %d", w.Offset, idxoffsetbits))
			}
			//nolint:gosec // test code, offset size determined by idxoffsetbits
			binary.BigEndian.PutUint32(b2[:4], uint32(w.Offset))
			binary.BigEndian.PutUint32(b2[4:8], w.Size)
		case 64:
			b2 = make([]byte, 12)
			binary.BigEndian.PutUint64(b2[:8], w.Offset)
			binary.BigEndian.PutUint32(b2[8:12], w.Size)
		default:
			panic(fmt.Sprintf("unsupported offset bits: %d", idxoffsetbits))
		}
		b = append(b, b2...)
	}
	return b
}
