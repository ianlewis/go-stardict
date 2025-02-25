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

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

var listCommand = &cli.Command{
	Name:            "list",
	Usage:           "List available dictionaries",
	HideHelp:        true,
	HideHelpCommand: true,
	Flags: []cli.Flag{
		// Special flags are shown at the end.
		&cli.BoolFlag{
			Name:               "help",
			Usage:              "print this help text and exit",
			Aliases:            []string{"h"},
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "version",
			Usage:              "print version information and exit",
			Aliases:            []string{"V"},
			DisableDefaultText: true,
		},
	},
	Action: func(c *cli.Context) error {
		if c.Bool("help") {
			check(cli.ShowCommandHelp(c, c.Command.Name))
			return nil
		}
		if c.Bool("version") {
			return printVersion(c)
		}

		dicts, errs := openStardicts(c.StringSlice("data-dir"))
		for _, err := range errs {
			// Ignore errors where data dir doesn't exist.
			if !errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "WARNING: %v\n", err)
			}
		}
		defer func() {
			for _, d := range dicts {
				d.Close()
			}
		}()

		tbl := table.New("Name", "Version", "Author", "Email", "Word Count")

		for _, dict := range dicts {
			tbl.AddRow(
				dict.Bookname(),
				dict.Version(),
				dict.Author(),
				dict.Email(),
				dict.WordCount(),
			)
		}
		tbl.Print()

		return nil
	},
}
