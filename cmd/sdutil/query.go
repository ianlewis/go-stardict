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

func renderArticle(a stardict.Article) string {
	for _, w := range a {
		switch w.Type() {
		case stardict.UTFTextType, stardict.HTMLType:
			return string(w.Data())
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
			defer func() {
				for _, dict := range dicts {
					dict.Close()
				}
			}()

			for _, dict := range dicts {
				fmt.Println(dict.Bookname())
				fmt.Println()
				idx := dict.Index()
				for idx.Scan() {
					e := idx.Entry()
					if strings.Contains(e.Word, query) {
						fmt.Println("  " + e.Word)
						if full {
							a, err := dict.Article(e)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								continue
							}
							fmt.Printf("%v\n", indent(renderArticle(a), "    "))
							fmt.Println()
						}
					}
				}
				if err := idx.Err(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				fmt.Println()
			}

			if len(errs) > 0 {
				os.Exit(1)
			}
		},
	}
	c.Flags().BoolVarP(&full, "full", "f", false, "output full articles")

	return c
}
