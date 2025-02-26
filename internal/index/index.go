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

package index

import (
	"fmt"
	"slices"
	"sort"
)

// Index is a generic sorted array index.
type Index[V fmt.Stringer] struct {
	// by the original file index.
	index []V

	cmp func(string, string) int
}

// NewIndex creates an index from the given slice and comparison function.
// cmp(a, b) should return a negative number when a < b, a positive number when
// a > b and zero when a == b or a and b are incomparable in the sense of a
// strict weak ordering.
func NewIndex[V fmt.Stringer](index []V, cmp func(string, string) int) *Index[V] {
	sorted := make([]V, len(index))
	copy(sorted, index)
	slices.SortFunc(sorted, func(a, b V) int {
		return cmp(a.String(), b.String())
	})

	return &Index[V]{
		index: sorted,
		cmp:   cmp,
	}
}

// Search performs a binary search over the index and returns matching words.
func (idx *Index[V]) Search(query string) []V {
	i, found := sort.Find(len(idx.index), func(i int) int {
		return idx.cmp(query, idx.index[i].String())
	})

	if !found {
		return nil
	}

	j := i
	//nolint:revive // This block increments j.1
	for ; j < len(idx.index) && idx.cmp(query, idx.index[j].String()) == 0; j++ {
	}
	return idx.index[i:j]
}
