# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 0996a2a8 - Update CHANGELOGs
- b10cc15a - Update CHANGELOGs

## [v0.1.6] - 2026-07-18

### Changed
- 79771f93 - Update all changelogs
- ae73b3c2 - Update all the changelogs
- c48959a8 - Update all changelogs
- 6731ec0a - Update all changelogs
- 0638b54a - Update changelogs
- 63db3549 - Update changelogs
- 7cb48abd - Add changelogs for all modules via `sv changelog -aw`

### Fixed
- d9733a23 - Use tabs in Earthfiles

## [v0.1.5] - 2026-05-24

### Changed
- 991531a9 - Bump go deps & run `go fix`
- 8ccb22df - Run `go fix work`

## [v0.1.4] - 2026-05-24

### Changed
- 51d22a3a - Bump go deps

## [v0.1.3] - 2026-05-24

### Changed
- 529ab5bc - Bump go deps

## [v0.1.2] - 2026-05-19

### Changed
- 85d23d61 - Run `gofumpt -w .`

## [v0.1.1] - 2026-05-19

### Changed
- f64b5684 - Update go deps

## [v0.1.0] - 2026-04-06

### Added
- 8adbf437 - Add joliv-spark example tests
- 6bef1285 - Add examples/go/thirdparty/rod
- 96b0006c - first commit

### Changed
- 066a5fb5 - Bump go version to 1.26.0
- d46a1052 - Bump go deps
- 3eb56a4b - Bump go deps
- e5ea5a85 - Update deps
- 13fa7a72 - Remove cruft
- d2bf7b16 - Update go modules
- f73cb4e2 - Run gofumpt
- ff3011b8 - Replace single-word msg t.Logf with t.Attr
- 8e3160f8 - Add Taskfile
- 08cdcd26 - Migrate to go.work with submodules
- 38ac174f - Tweak thirdparty/rod test
- 3cc59747 - Replace wg.Add/Done with wg.Go
- 25cd08f4 - Run gofumpt

### Fixed
- 3e8341fd - Bump antchfx/xpath (govulncheck)
- 1d9fd13a - restore stdout/stderr after test capture in cli-simple
- 8cf13fc7 - Testcontainers nats test
- 54beacd2 - Formatter errors (cue fmt)
- 7466edb7 - Add brokenOnArch tag to playwright test
- 0184efa3 - Run modernize
- c0b93bda - Trivy vulnerabilities via go mod updates
- a9bc4785 - All integration test errors
- e98dda8a - Use valid example name and bump image version
- c16b7e46 - Use explicit `main` alias for main_test packages (for goimports)
- 3e81667a - Use integration tag for rscio-script ping test
- 378300f5 - Move ping test to integration tests folder
