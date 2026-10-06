# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

## [v0.3.0] - 2026-10-05

### Added
- b6960607 - establish command promotion lifecycle
- 823f0ab1 - make language formatting conventions modular
- 8202ca1d - lint multiple files in one invocation
- 112e669c - add punctuation fix to --fix
- cd588d80 - add --fix flag for auto-correcting fixable issues
- 52d9b671 - add awesome-lint CLI tool for linting awesome lists

### Changed
- 3e477efd - Bump deps
- 348b5218 - keep task update from breaking the build
- 878b20c2 - Update deps
- be077bae - Update deps
- b8eceb75 - Add Taskfile
- 53b49e27 - Update deps
- 862a6421 - Update deps
- 05e26cbe - cover Fix funnels to pass the CRAP gate
- 7c5bb1e1 - Add lint:go-crap task
- ac3f5326 - reduce CRAP scores for validateListItem and setupCLI
- e5011487 - add unit tests to bring linter CRAP scores under threshold

### Fixed
- 71db28d5 - use Chinese full stop in Chinese READMEs
- 8a3a2dc3 - only spell-check prose so --fix stops corrupting links
- 621f4f09 - handle case-insensitive filenames and fix false positives
- 3d8d5989 - match npx output on badge, list-item, double-link rules
