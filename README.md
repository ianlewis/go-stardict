# go-stardict

[![Go Reference](https://pkg.go.dev/badge/github.com/ianlewis/go-stardict.svg)](https://pkg.go.dev/github.com/ianlewis/go-stardict)

A stardict dictionary library for Go.

## Status

The API is currently *unstable* and will change. This package will use [module
version numbering](https://golang.org/doc/modules/version-numbers) to manage
versions and compatibility.

## Installation

To install this package run

`go get github.com/ianlewis/go-stardict`

## Examples

You can search a stardict dictionary via it's index.

```golang
// Open dictonaries in a directory
dictionaries, _ := stardict.OpenAll(".")
for _, d := range dictionaries {

  // Search the index.
  idx, _ := d.Index()
  dict, _ := d.Dict()
  for _, e := range idx.FullTextSearch("banana") {

    // Print out matching index entries.
    fmt.Println(e.Word)
    fmt.Println()

    // Print out a full word.
    w, _ := dict.Word(e)
    for _, d := range w.Data() {
      fmt.Println(string(d.Data()))
    }
  }
}
```

## References

- [Format for StarDict dictionary files](https://github.com/huzheng001/stardict-3/blob/master/dict/doc/StarDictFileFormat)
