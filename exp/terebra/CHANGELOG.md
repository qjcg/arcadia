# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- d4208456 - Bump deps

## [v0.5.0] - 2026-07-19

### Added
- 08b2308d - add fuzzy search TUI with bubbletea v2; fix exit to cleanly propagate

### Fixed
- 444decd8 - make Ctrl+r immediately enter the TUI; fix setupStdinPipe syntax
- 71a1eb92 - make Ctrl+r immediately enter the TUI instead of waiting for Enter

## [v0.4.4] - 2026-07-19

### Fixed
- d3d8a459 - terebra build and multi-line script execution

## [v0.4.3] - 2026-07-19

### Fixed
- 3e13ca70 - also handle ^Z in fg builtin, not just in foreground commands
- f6e4f81e - make ^Z (SIGTSTP) stop foreground jobs instead of doing nothing

## [v0.4.2] - 2026-07-19

### Fixed
- 1d1168d4 - handle Ctrl+C in fg via signal forwarding, no terminal manipulation

## [v0.4.1] - 2026-07-19

### Changed
- e3946493 - Update CHANGELOGs

### Fixed
- 23efe68a - enable Ctrl+C and Ctrl+Z in fg'd background jobs

## [v0.4.0] - 2026-07-19

### Added
- a140ff69 - implement temporary env vars for one command via `FOO=bar cmd`
- 36257d73 - implement here-string via `<<<`
- 447fc11b - implement &>, &>>, and |&
- 6c9fef5e - add per-builtin help via `help <builtin>`
- 445ba72f - allow array access without curly braces
- 2016f826 - implement cd - (toggle to previous directory)

### Changed
- 45ce93e1 - remove cue.txtar — tests an external binary, not terebra behavior
- 18cfedff - add comprehensive testscript coverage for shell features
- 11085af9 - fix brittle assertions in basic and builtin tests
- a38d54b4 - fix brittle testscript assertions
- f5504b73 - move cmd/terebra/* to project root, remove cmd/terebra subdir
- 5d1bdb95 - use debug.ReadBuildInfo() for version instead of hardcoded constant
- 2436ab2b - add globstar ** recursive glob tests
- 7d81dc01 - remove declare builtin (redundant with auto-promotion)
- 89b51c95 - update README and help to reflect current implementation
- b0b0f615 - update testscript plan with 15 test categories
- d36bd607 - add 15 testscript test suites covering all shell features
- 4d5b3098 - add testscript test harness and basic test suite

### Fixed
- 762b8dc4 - correct cursor position in prompts with ANSI escape codes
- eab8d074 - allow encoder-only auger pipelines and fix parser early return
- ea86167e - filename completion with ./ prefix no longer duplicates the prefix
- ac507f3e - heredoc interactive mode - trailing newline and empty content detection
- b29fcae4 - heredocs now work in interactive REPL sessions
- 572091a4 - remove declare from help text (missed in earlier removal)

## [v0.3.0] - 2026-07-19

### Added
- 4a2b3554 - auto-promote arrays to associative on non-numeric key
- 9e131a99 - add associative array (dictionary) support

## [v0.2.0] - 2026-07-19

### Added
- 19b6f9cc - add debug mode, PS1 expansion, $(cmd), fuzzy search, rc file
- 7bd51f8e - add alias, cue, history, readonly; remove ls; plugin extensions
- fa5a512b - add globbing, command chaining, and heredoc support

### Changed
- fa150990 - update Taskfile and add implementation plan docs
- 0996a2a8 - Update CHANGELOGs
- f5915ee4 - Bump deps

## [v0.1.0] - 2026-07-18

### Added
- 6ed9977b - Add terebra

### Changed
- a49480db - Add missing final newline
