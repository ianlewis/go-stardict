# go-stardict

[![Go Reference](https://pkg.go.dev/badge/github.com/ianlewis/go-stardict.svg)](https://pkg.go.dev/github.com/ianlewis/go-stardict)
[![codecov](https://codecov.io/gh/ianlewis/go-stardict/graph/badge.svg?token=2HTI2KXI93)](https://codecov.io/gh/ianlewis/go-stardict)
[![Tests](https://github.com/ianlewis/go-stardict/actions/workflows/pre-submit.units.yml/badge.svg)](https://github.com/ianlewis/go-stardict/actions/workflows/pre-submit.units.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ianlewis/go-stardict)](https://goreportcard.com/report/github.com/ianlewis/go-stardict)

A stardict dictionary library for Go.

## Status

The API is currently _unstable_ and will change. This package will use [module
version numbering](https://golang.org/doc/modules/version-numbers) to manage
versions and compatibility.

## Features

- \[x] Reading dictionary metadata.
- \[x] Reading & searching the dictionary index.
- \[x] Reading full dictionary articles.
- \[x] Efficient access for large files.
- \[x] Dictzip support.
- \[x] Support for concurrent access.
- \[ ] Synonym support (.syn file) (#2).
- \[ ] Support for tree dictionaries (.tdx file) (#3).
- \[ ] Support for Resource Storage (res/ directory) (#4).
- \[ ] Glob/Wildcard search support (#21).
- \[ ] Capitalization, diacritic, punctuation, and whitespace folding (#19, #25).

## Installation

To install this package run

`go get github.com/ianlewis/go-stardict`

## Examples

You can search a stardict dictionary directly and list the entries.

```golang
// Open dictonaries in a directory
dictionaries, _ := stardict.OpenAll(".")

// Search the dictionaries.
for _, d := range dictionaries {
  entries, _ := d.Search("banana")
  for _, e := range entries {
    // Print out matching index entries.
    fmt.Println(e)
  }
}
```

## Related projects

- [ilius/go-stardict](github.com/ilius/go-stardict)
- [dyatlov/gostardict](https://github.com/dyatlov/gostardict)

## References

- [Format for StarDict dictionary files](https://github.com/huzheng001/stardict-3/blob/master/dict/doc/StarDictFileFormat)
