# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [exp/sv/unreleased]

### Fixed
- 8a828b66 - include pre-rename commits in submodule changelogs

## [exp/sv/v0.10.0] - 2026-06-25

### Added
- 3963b900 - Add `--write` flag

### Changed
- 49cce786 - Update deps

## [exp/sv/v0.9.0] - 2026-06-19

### Added
- 3c8598e0 - add validate-cc subcommand for conventional commit validation

### Changed
- 7b93b891 - Run `go fix work`

### Fixed
- 8060e250 - remove redundant VALID/INVALID output from validate-cc

## [exp/sv/v0.8.6] - 2026-06-19

### Fixed
- f1cf5ccb - sort changelog entries by semver, not lexicographically

## [exp/sv/v0.8.5] - 2026-06-19

### Fixed
- 93257ba6 - prevent cross-module commit leaks from bogus rename detection

## [exp/sv/v0.8.4] - 2026-06-19

### Fixed
- f4f57a20 - prevent retracted versions from duplicating commits in changelog

## [exp/sv/v0.8.3] - 2026-06-19

### Fixed
- 348b27e8 - avoid duplicate # Changelog headings in multi-module output

## [exp/sv/v0.8.2] - 2026-06-19

### Fixed
- 7d51a924 - exclude submodule commits from root module changelog

## [exp/sv/v0.8.1] - 2026-06-19

### Fixed
- 0da1deca - include historical commits in changelog after module renames
- f84dd7ff - search recursively for changelog overview files
- 63502d70 - suppress changelog stdout with --dir and clean stale entry files

## [exp/sv/v0.8.0] - 2026-06-19

### Added
- dfd978e9 - replace --since with --from/--to with date-aware tag generation
- 80391959 - add changelog command for keepachangelog output
- 8741730e - add changelog feature design document

### Changed
- 35e69b52 - Run `go fix`
- a8cd3eec - Update go deps
- b2cdf599 - document --tag, --tag-format, and --dry-run flags in README

## [exp/sv/v0.7.0] - 2026-06-18

### Added
- cb2838cb - add --dry-run flag to `sv next`
- eb6f23dc - add annotated git tag support to `sv next`

### Changed
- 3f071074 - rename --debug flag to --verbose

## [exp/sv/v0.6.0] - 2026-06-15

### Added
- 2ee093a1 - skip retracted Go module versions when calculating next version

### Changed
- 7cae5fa3 - Update go deps
- 54dd9f3e - Update go deps

## [exp/sv/v0.5.0] - 2026-05-25

### Added
- b1e1b863 - add --default-patch flag to next subcommand

## [exp/sv/v0.4.0] - 2026-05-25

### Added
- b0a103fc - Remove -v/--verbose flags from subcommands

## [exp/sv/v0.3.0] - 2026-05-25

### Added
- 6a33fcef - support repeated and comma-separated --path flag

### Changed
- 35baa6af - Run `task lint:fix`
- 512e7454 - extract CLI subcommands into internal/cli

## [exp/sv/v0.2.4] - 2026-05-24

### Changed
- 51d22a3a - Bump go deps

## [exp/sv/v0.2.3] - 2026-05-19

### Changed
- 5b9213e1 - Tweak clean task

## [exp/sv/v0.2.2] - 2026-05-19

### Changed
- 1db5f54c - use boa CLI framework

## [exp/sv/v0.2.1] - 2026-05-19

### Changed
- f64b5684 - Update go deps

### Fixed
- d9d8d0ec - show modules with chore-only commits in sv next -a

## [exp/sv/v0.2.0] - 2026-04-06

### Added
- 937e5936 - add --version flag using debug.ReadBuildInfo

### Fixed
- 0a787d56 - use commit hash instead of message in file filtering

## [exp/sv/v0.1.0] - 2026-04-06

### Added
- 011d7d20 - default to v0.1.0 for untagged modules
- d887b140 - implement semantic versioning tool with monorepo support

### Changed
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
- 56610dcc - filter commits by module path
