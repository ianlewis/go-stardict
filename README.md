# go-stardict

A stardict library for Go

## Installation

To install this package run

`go get github.com/ianlewis/go-stardict`

## Examples

You can search a stardict dictionary via it's index.

```golang
// Open dictonaries in a directory
dicts, _ := stardict.OpenAll(".")
for _, dict := range dicts {

  // Search the index.
  idx, _ := dict.Index()
  for _, e := range idx.FullTextSearch("banana") {

    // Print out matching index entries.
    fmt.Println(e.Word)
    fmt.Println()

    // Print out a full article.
    a, _ := dict.Article(e)
    for _, w := range a {
      fmt.Println(string(w.Data()))
    }
  }
}
```
