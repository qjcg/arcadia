# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 51f343c6 - Bump go deps
- 79771f93 - Update all changelogs
- fab8e0ff - Add newline at end-of-file
- ae73b3c2 - Update all the changelogs

## [v1.0.1] - 2026-06-28

### Changed
- c48959a8 - Update all changelogs

### Fixed
- b1127611 - scope breaking change detection to source code per module
- 8a52bfa9 - include tags with no scoped commits in changelog

## [v1.0.0] - 2026-06-28

### Changed
- 150c17c7 - Update go tool to use new cmd/sv path

## [v0.1.1] - 2026-06-28

### Fixed
- f93827df - Use appropriate dir names in relative_paths.txtar

## [v0.1.0] - 2026-06-28

### Added
- 3963b900 - Add `--write` flag
- 3c8598e0 - add validate-cc subcommand for conventional commit validation
- dfd978e9 - replace --since with --from/--to with date-aware tag generation
- 80391959 - add changelog command for keepachangelog output
- 8741730e - add changelog feature design document
- cb2838cb - add --dry-run flag to `sv next`
- eb6f23dc - add annotated git tag support to `sv next`
- 2ee093a1 - skip retracted Go module versions when calculating next version
- b1e1b863 - add --default-patch flag to next subcommand
- b0a103fc - Remove -v/--verbose flags from subcommands
- 6a33fcef - support repeated and comma-separated --path flag
- 937e5936 - add --version flag using debug.ReadBuildInfo
- 011d7d20 - default to v0.1.0 for untagged modules
- d887b140 - implement semantic versioning tool with monorepo support

### Changed
- 9e483f97 - Graduate out of exp(erimental) dir to cmd
- 6731ec0a - Update all changelogs
- 0638b54a - Update changelogs
- 63db3549 - Update changelogs
- 7cb48abd - Add changelogs for all modules via `sv changelog -aw`
- 49cce786 - Update deps
- 7b93b891 - Run `go fix work`
- 35e69b52 - Run `go fix`
- a8cd3eec - Update go deps
- b2cdf599 - document --tag, --tag-format, and --dry-run flags in README
- 3f071074 - rename --debug flag to --verbose
- 7cae5fa3 - Update go deps
- 54dd9f3e - Update go deps
- 35baa6af - Run `task lint:fix`
- 512e7454 - extract CLI subcommands into internal/cli
- 51d22a3a - Bump go deps
- 5b9213e1 - Tweak clean task
- 1db5f54c - use boa CLI framework
- f64b5684 - Update go deps
- 3c976e6b - improve test coverage and fix error handling
- a7d247e8 - Improve sv test coverage and fix code quality
- 844a2f69 - Rename x/ directory to exp/
- efeafa48 - Simplify output format and add verbose flag
- 0301281f - Tweak Taskfile
- 6c0c0481 - Move main.go to top-level
- 46748969 - align slidesdeck and sv testing with AGENTS.md guidance
- dd411f1a - Update Taskfile
- 16b67c59 - add module definitions and git attributes
- c585192c - add product vision and automation documentation

### Fixed
- 8a828b66 - include pre-rename commits in submodule changelogs
- 8060e250 - remove redundant VALID/INVALID output from validate-cc
- f1cf5ccb - sort changelog entries by semver, not lexicographically
- 93257ba6 - prevent cross-module commit leaks from bogus rename detection
- f4f57a20 - prevent retracted versions from duplicating commits in changelog
- 348b27e8 - avoid duplicate # Changelog headings in multi-module output
- 7d51a924 - exclude submodule commits from root module changelog
- 0da1deca - include historical commits in changelog after module renames
- f84dd7ff - search recursively for changelog overview files
- 63502d70 - suppress changelog stdout with --dir and clean stale entry files
- d9d8d0ec - show modules with chore-only commits in sv next -a
- 0a787d56 - use commit hash instead of message in file filtering
- 56610dcc - filter commits by module path
