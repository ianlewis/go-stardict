# Changelog

All notable changes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0-rc.0] - 2025-03-06

### Added in v0.2.0-rc.0

- The index now supports whitespace and punctuation folding ([#25](https://github.com/ianlewis/go-stardict/issues/25)).
- The synonym index (.syn) file is now supported ([#2](https://github.com/ianlewis/go-stardict/issues/2)).
- `Stardict.Search` and `Idx.Search` now support queries in glob format ([#21](https://github.com/ianlewis/go-stardict/issues/21)).

### Changed in v0.2.0-rc.0

- The minimum supported Go version is now 1.23.
- `stardict.Open` and `stardict.OpenAll` now take an `options` argument which allows for specifying options for opening dictionaries ([#87](https://github.com/ianlewis/go-stardict/issues/87)).
- `stardict.idx.Options.Folder` is now a constructor `func() transform.Transformer` rather than a static `golang.org/x/text/transform.Transformer` value ([#87](https://github.com/ianlewis/go-stardict/issues/87)).

## [0.1.0] - 2024-11-04

- Initial release
- Basic dict entry, index, search support.

[0.1.0]: https://github.com/ianlewis/go-stardict/releases/tag/v0.1.0
[0.2.0-rc.0]: https://github.com/ianlewis/go-stardict/releases/tag/v0.2.0-rc.0
