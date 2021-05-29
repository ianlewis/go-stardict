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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ianlewis/go-stardict"
)

func renderWord(w *stardict.Word) string {
	for _, d := range w.Data() {
		switch d.Type() {
		case stardict.UTFTextType, stardict.HTMLType:
			return string(d.Data())
		}
	}
	return ""
}

func indent(text, indent string) string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			line = indent + line
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func queryCommand() *cobra.Command {
	var full bool

	c := &cobra.Command{
		Use:   "query [DIR] [QUERY]",
		Short: "Query dictionaries",
		Long:  `Query all dictionaries in a directory.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			query := args[1]

			dicts, errs := stardict.OpenAll(path)
			for _, err := range errs {
				fmt.Fprintln(os.Stderr, err)
			}

			for _, d := range dicts {
				fmt.Println(d.Bookname())
				fmt.Println()
				idx, err := d.Index()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				dict, err := d.Dict()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				for _, e := range idx.FullTextSearch(query) {
					fmt.Println("  " + e.Word)
					if full {
						a, err := dict.Word(e)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
							continue
						}
						fmt.Printf("%v\n", indent(renderWord(a), "    "))
						fmt.Println()
					}
				}
				fmt.Println()
			}

			if len(errs) > 0 {
				os.Exit(1)
			}
		},
	}
	c.Flags().BoolVarP(&full, "full", "f", false, "output full entries")

	return c
}
