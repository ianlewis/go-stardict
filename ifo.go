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
	"os"
	"regexp"
	"strings"
)

const magic = "StarDict's dict ifo file"

var keyRegex = regexp.MustCompile("[a-zA-Z0-9-_]+")

type Ifo struct {
	metadata map[string]string
}

// NewIfo returns a new dictionary info object from the path.
func NewIfo(path string) (*Ifo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	metadata := map[string]string{}
	s := bufio.NewScanner(bufio.NewReader(f))
	if s.Scan() {
		if s.Text() != magic {
			return nil, fmt.Errorf("bad magic data")
		}
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

		metadata[key] = value
		i++
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return &Ifo{
		metadata: metadata,
	}, nil
}

func (i *Ifo) Value(key string) string {
	return i.metadata[key]
}

func readKV(line string) (string, string, error) {
	v := strings.SplitN(line, "=", 2)
	key := strings.TrimRight(v[0], " ")
	value := strings.TrimLeft(v[1], " ")
	return key, value, nil
}
