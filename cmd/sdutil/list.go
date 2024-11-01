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

	"github.com/ianlewis/go-stardict"
)

type listCommand struct{}

func (*listCommand) Name() string {
	return "list"
}

func (*listCommand) Synopsis() string {
	return "List dictionaries"
}

func (*listCommand) Usage() string {
	return `list [DIR]
List all dictionaries in a directory.`
}

func (*listCommand) SetFlags(_ *flag.FlagSet) {}

func (*listCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "unexpected number of arguments")
		return subcommands.ExitUsageError
	}

	dicts, errs := stardict.OpenAll(args[0])
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	defer func() {
		for _, d := range dicts {
			d.Close()
		}
	}()

	for _, dict := range dicts {
		fmt.Printf("Name         %s\n", dict.Bookname())
		fmt.Printf("Author:      %s\n", dict.Author())
		fmt.Printf("Email:       %s\n", dict.Email())
		fmt.Printf("Word Count:  %d\n", dict.WordCount())
		fmt.Println()
	}

	if len(errs) > 0 {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
