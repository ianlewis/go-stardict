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

	"github.com/google/subcommands"
	"github.com/k3a/html2text"
	"github.com/rodaine/table"

	"github.com/ianlewis/go-stardict"
	"github.com/ianlewis/go-stardict/dict"
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

func (*queryCommand) SetFlags(_ *flag.FlagSet) {}

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
			fmt.Println("-------------------------------------------------------------------------------")

			printEntries(entries)
			fmt.Println()
			// for _, e := range entries {
			// 	// Trim off any trailing whitespace.
			// 	fmt.Println(strings.TrimSpace(e.String()))
			// 	fmt.Println()
			// }
		}
	}

	if len(errs) > 0 {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func printEntries(entries []*stardict.Entry) {
	tbl := table.New("Title", "Data").WithHeaderFormatter(func(string, ...interface{}) string { return "" })
	for _, e := range entries {
		text := ""
		for _, d := range e.Data() {
			// string will work for PhoneticType, UTFTextType, YinBiaoOrKataType, HTMLType
			switch d.Type {
			case dict.PhoneticType, dict.UTFTextType, dict.YinBiaoOrKataType:
				text += string(d.Data) + "\n"
			case dict.HTMLType:
				text += html2text.HTML2Text(string(d.Data)) + "\n"
			default:
				// TODO(#22): Support XDXF format.
			}
		}

		tbl.AddRow(e.Title(), text)
	}

	tbl.Print()
}
