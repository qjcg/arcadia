# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 6731ec0a - Update all changelogs
- 0638b54a - Update changelogs
- 63db3549 - Update changelogs
- 7cb48abd - Add changelogs for all modules via `sv changelog -aw`
- 49cce786 - Update deps

## [v0.2.3] - 2026-06-19

### Fixed
- 93257ba6 - prevent cross-module commit leaks from bogus rename detection

## [v0.2.2] - 2026-06-19

### Fixed
- f4f57a20 - prevent retracted versions from duplicating commits in changelog

## [v0.2.1] - 2026-06-19

### Fixed
- 348b27e8 - avoid duplicate # Changelog headings in multi-module output

## [v0.2.0] - 2026-06-19

### Changed
- a8cd3eec - Update go deps
- 7cae5fa3 - Update go deps
- 54dd9f3e - Update go deps

## [v0.1.1] - 2026-05-25

### Fixed
- 19697b5d - Correct CRA/RQ penalty and interest calculations

## [v0.1.0] - 2026-05-25

### Added
- ed329188 - Add fin README
- a6763c15 - Add project-level agent-browser skill
- e732c8f8 - Create fc financial calculator with tax pic subcommand

### Changed
- 7cffc888 - Update go deps
- 20b2e9ca - Update fc references to fin across module
- 47ba6ef7 - Rename exp/fc to exp/fin
- 03ee0abf - Rename pic internal symbols to pi
- 6b6a4075 - Split monolithic internal package into subdirectory packages
- 0b0613a4 - Rename pic subcommand to penalties-and-interest with pi alias
