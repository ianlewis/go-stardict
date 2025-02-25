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

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ianlewis/go-stardict"
)

const (
	// ExitCodeSuccess is successful error code.
	ExitCodeSuccess int = iota

	// ExitCodeFlagParseError is the exit code for a flag parsing error.
	ExitCodeFlagParseError

	// ExitCodeUnknownError is the exit code for an unknown error.
	ExitCodeUnknownError
)

// ErrSdutil is a parent error for all command errors.
var ErrSdutil = errors.New("sdutil")

// ErrFlagParse is a flag parsing error.
var ErrFlagParse = fmt.Errorf("%w: parsing flags", ErrSdutil)

// ErrUnsupported indicates a feature is unsupported.
var ErrUnsupported = fmt.Errorf("%w: unsupported", ErrSdutil)

var copyrightNames = []string{
	"2021 Google LLC",
	"2024 Ian Lewis",
}

//nolint:gochecknoinits // init needed needed for global variable.
func init() {
	// Set the HelpFlag to a random name so that it isn't used. `cli` handles
	// the flag with the root command such that it takes a command name argument
	// but we don't use commands.
	//
	// This is done because `dictzip --help foo` will display a
	// "command foo not found" error instead of the help.
	//
	// This flag is hidden by the help output.
	// See: github.com/urfave/cli/issues/1809
	cli.HelpFlag = &cli.BoolFlag{
		// NOTE: Use a random name no one would guess.
		Name:               "d41d8cd98f00b204e980",
		DisableDefaultText: true,
	}
}

// check checks the error and panics if not nil.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

func openStardicts(dirs []string) ([]*stardict.Stardict, []error) {
	var dicts []*stardict.Stardict
	var errs []error

	for _, path := range dirs {
		openDicts, openErrs := stardict.OpenAll(path)

		dicts = append(dicts, openDicts...)
		errs = append(errs, openErrs...)
	}

	return dicts, errs
}

func newStardictApp() *cli.App {
	return &cli.App{
		Name:  filepath.Base(os.Args[0]),
		Usage: "Search Stardict dictionaries.",
		Description: strings.Join([]string{
			"Stardict utility written in Go.",
			"http://github.com/ianlewis/go-stardict",
		}, "\n"),
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "data-dir",
				Usage:   "include dictionaries in `DIR`",
				Aliases: []string{"d"},
				Value:   cli.NewStringSlice(dictLocations()...),
			},

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
		Copyright:       strings.Join(copyrightNames, "\n"),
		HideHelp:        true,
		HideHelpCommand: true,
		Action: func(c *cli.Context) error {
			if c.Bool("version") {
				return printVersion(c)
			}

			check(cli.ShowAppHelp(c))
			return nil
		},
		Commands: []*cli.Command{
			listCommand,
			queryCommand,
		},
	}
}
