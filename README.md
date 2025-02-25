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

- \[x] Reading dictionary metadata (.ifo files).
- \[x] Reading & searching the dictionary index (.idx file).
- \[x] Reading full dictionary articles.
- \[x] Efficient access for large files.
- \[x] Dictzip support.
- \[x] Capitalization, diacritic, punctuation, and whitespace folding ([#19](https://github.com/ianlewis/go-stardict/issues/19), [#25](https://github.com/ianlewis/go-stardict/issues/25)).
- \[x] Synonym support (.syn file) ([#2](https://github.com/ianlewis/go-stardict/issues/2)).
- \[ ] Support for tree dictionaries (.tdx file) ([#3](https://github.com/ianlewis/go-stardict/issues/3)).
- \[ ] Support for Resource Storage (res/ directory) ([#4](https://github.com/ianlewis/go-stardict/issues/4)).
- \[ ] Support for collation files (.idx.clt, .syn.clt) ([#7](https://github.com/ianlewis/go-stardict/issues/7))
- \[ ] Support for offset cache files (.idx.oft, .syn.oft) ([#8](https://github.com/ianlewis/go-stardict/issues/8))
- \[ ] Glob/Wildcard search support ([#21](https://github.com/ianlewis/go-stardict/issues/21)).

## Installation

To install this package run

`go get github.com/ianlewis/go-stardict`

## Examples

### Searching a Dictionary

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

- [ilius/go-stardict](https://github.com/ilius/go-stardict)
- [dyatlov/gostardict](https://github.com/dyatlov/gostardict)

## References

- [Format for StarDict dictionary files](https://github.com/huzheng001/stardict-3/blob/master/dict/doc/StarDictFileFormat)
