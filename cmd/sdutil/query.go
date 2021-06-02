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
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/subcommands"

	"github.com/ianlewis/go-stardict"
)

type queryCommand struct{}

func (*queryCommand) Name() string {
	return "query"
}

func (*queryCommand) Synopsis() string {
	return "Query dictionaries"
}

func (*queryCommand) Usage() string {
	return `query [DIR] [QUERY]
Query all dictionaries in a directory.`
}

func (*queryCommand) SetFlags(f *flag.FlagSet) {}

func (c *queryCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	path := args[0]
	query := args[1]

	dicts, errs := stardict.OpenAll(path)
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		for _, d := range dicts {
			d.Close()
		}
	}()

	dictResults := 0
	for _, d := range dicts {
		entries, err := d.Search(query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if len(entries) > 0 {
			if dictResults > 0 {
				fmt.Println()
			}
			dictResults++

			fmt.Println(d.Bookname())
			fmt.Println()

			for _, e := range entries {
				// Trim off any trailing whitespace.
				fmt.Println(strings.TrimSpace(e.String()))
				fmt.Println()
			}
		}
	}

	if len(errs) > 0 {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
