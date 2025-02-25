// Copyright 2025 Ian Lewis
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

//go:build !windows

package main

import (
	"os"
	"path/filepath"
)

func dictLocations() []string {
	loc := []string{
		"/usr/share/stardict/dic",
	}

	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		loc = append(loc, filepath.Join(xdgDataHome, "stardict/dic"))
	}

	if stardictDataDir := os.Getenv("STARDICT_DATA_DIR"); stardictDataDir != "" {
		loc = append(loc, filepath.Join(stardictDataDir, "dic"))
	}

	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		loc = append(loc, filepath.Join(homeDir, ".stardict/dic"))
	}

	return loc
}
