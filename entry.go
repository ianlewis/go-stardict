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
	"strings"

	"github.com/ianlewis/go-stardict/dict"
)

// DataList is entry's data.
type DataList []*dict.Data

// String returns a string representation of the data list.
func (l DataList) String() string {
	var b strings.Builder
	for _, d := range l {
		_, _ = b.WriteString(d.String())
		_, _ = b.WriteRune('\n')
	}
	return b.String()
}

// Entry is a dictionary entry.
type Entry struct {
	word string
	data DataList
}

// Title return the entry's title.
func (e *Entry) Title() string {
	return e.word
}

// Data returns the entry's data entries.
func (e *Entry) Data() DataList {
	return e.data
}

// String returns a string representation of the Entry.
func (e *Entry) String() string {
	var b strings.Builder
	_, _ = b.WriteString(e.word)
	_, _ = b.WriteRune('\n')
	_, _ = b.WriteString(e.data.String())
	_, _ = b.WriteRune('\n')
	return b.String()
}
