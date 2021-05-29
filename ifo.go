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
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var keyRegex = regexp.MustCompile("[a-zA-Z0-9-_]+")

// Ifo holds metadata read from .ifo files.
type Ifo struct {
	magic    string
	metadata map[string]string
}

// NewIfo returns a new dictionary info object from the path.
func NewIfo(r io.Reader) (*Ifo, error) {
	ifo := &Ifo{
		metadata: map[string]string{},
	}

	s := bufio.NewScanner(bufio.NewReader(r))
	if s.Scan() {
		ifo.magic = s.Text()
	}

	i := 0
	for s.Scan() {
		line := s.Text()
		if strings.Trim(line, " ") == "" {
			continue
		}
		v := strings.SplitN(line, "=", 2)
		key := strings.TrimRight(v[0], " ")
		value := strings.TrimLeft(v[1], " ")
		if !keyRegex.Match([]byte(key)) {
			return nil, fmt.Errorf("invalid key: %v", key)
		}
		if i == 0 && key != "version" {
			return nil, fmt.Errorf("missing version")
		}

		ifo.metadata[key] = value
		i++
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return ifo, nil
}

// Magic returns the ifo file's magic string.
func (i *Ifo) Magic() string {
	return i.magic
}

// Value returns a value from the metadata file.
func (i *Ifo) Value(key string) string {
	return i.metadata[key]
}
