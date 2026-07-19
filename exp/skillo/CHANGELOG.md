# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 0996a2a8 - Update CHANGELOGs
- b10cc15a - Update CHANGELOGs
- 51f343c6 - Bump go deps

## [v0.4.0] - 2026-07-05

### Added
- d6ca9388 - add vim-style keys and Ctrl-p/n for menu navigation
- 4949dd07 - add interactive skill selection menu for add and remove

### Changed
- dfe50743 - Run `go fix`
- a71db481 - flatten package to exp/skillo root, drop cmd/ subdir

### Fixed
- 2d7ab0df - prevent extracted skills from becoming orphaned on partial error
- aed6f603 - ensure skillo writes JSON files with trailing newline
- 3c40256d - prevent skillo from showing project skills as stale after install

## [v0.3.0] - 2026-06-29

### Added
- 96fe7a3e - add tree view for list command (-t/--tree)
- 4c772023 - add project-mode skill directories with two-scope architecture

### Changed
- 581727ff - Run `go fix`
- 79771f93 - Update all changelogs

### Fixed
- a3987139 - improve list output with headers and reliable version lookup
- dc2ef76f - skip invalid skills during extraction instead of aborting
- f1665efa - use go list -m -e -json with module cache fallback
- 4443f76a - make list respect --user and --project flags
- d8456369 - fall back to module cache when go list -m fails for module dir
- 11e02765 - add go mod download before go list -m in add command

## [v0.2.0] - 2026-06-29

### Added
- a9335c65 - enhance CLI with user/project dirs and rich list output
- 6c4543b8 - add skilldirs package and version tracking to manifest

### Changed
- fafc522e - add BDD feature tests with godog
- ae73b3c2 - Update all the changelogs
- c48959a8 - Update all changelogs
- 6731ec0a - Update all changelogs
- 0638b54a - Update changelogs
- 63db3549 - Update changelogs
- 7cb48abd - Add changelogs for all modules via `sv changelog -aw`
- 49cce786 - Update deps
- a8cd3eec - Update go deps

## [v0.1.3] - 2026-06-18

### Changed
- 7cae5fa3 - Update go deps
- 54dd9f3e - Update go deps

### Fixed
- 6ed0c4bf - accept string or array for allowed-tools in SKILL.md

## [v0.1.2] - 2026-05-24

### Changed
- 8ccb22df - Run `go fix work`

## [v0.1.1] - 2026-05-24

### Changed
- 51d22a3a - Bump go deps

## [v0.1.0] - 2026-04-06

### Added
- ef89d9a6 - Add x/skillo

### Changed
- b9925414 - Fix broken URL
- 844a2f69 - Rename x/ directory to exp/

### Fixed
- 877027a4 - Remove outdated warning expectation in get test
