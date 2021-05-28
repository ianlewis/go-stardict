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

	"github.com/spf13/cobra"

	"github.com/ianlewis/go-stardict"
)

var listCmd = &cobra.Command{
	Use:   "list [DIR]",
	Short: "List dictionaries",
	Long:  `List all dictionaries in a directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dicts, errs := stardict.OpenAll(args[0])
		for _, err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}

		for _, dict := range dicts {
			fmt.Printf("Name         %s\n", dict.Bookname())
			fmt.Printf("Author:      %s\n", dict.Author())
			fmt.Printf("Email:       %s\n", dict.Email())
			fmt.Printf("Word Count:  %d\n", dict.WordCount())
			fmt.Println()
		}

		if len(errs) > 0 {
			os.Exit(1)
		}
	},
}
