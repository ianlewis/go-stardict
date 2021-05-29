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
	"github.com/ianlewis/go-stardict/dict"
)

// Entry is a dictionary entry.
type Entry struct {
	word string
	data []*dict.Data
}

// Title return the entry's title.
func (e *Entry) Title() string {
	return e.word
}

// Data returns the entry's data entries.
func (e *Entry) Data() []*dict.Data {
	return e.data
}

// String returns a string representation of the Entry.
func (e *Entry) String() string {
	str := e.word + "\n"
	for _, d := range e.Data() {
		switch d.Type() {
		case dict.PhoneticType, dict.UTFTextType, dict.YinBiaoOrKataType, dict.HTMLType:
			str += string(d.Data()) + "\n"
		}
	}
	return str
}
