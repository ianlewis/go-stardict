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

var queryCmd = &cobra.Command{
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
					fmt.Println("    " + e.Word)
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
