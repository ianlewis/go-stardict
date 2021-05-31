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

func queryCommand() *cobra.Command {
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
				for _, d := range dicts {
					d.Close()
				}
			}()

			for i, d := range dicts {
				entries, err := d.Search(query)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				if len(entries) > 0 {
					if i > 0 {
						fmt.Println()
					}
					fmt.Println(d.Bookname())
					fmt.Println()

					for _, e := range entries {
						// Trim off any trailing whitespace.
						fmt.Println(strings.TrimSpace(e.String()))
					}
				}
			}

			if len(errs) > 0 {
				os.Exit(1)
			}
		},
	}

	return c
}
