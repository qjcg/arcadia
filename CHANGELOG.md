# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 08199a49 - update changelogs for released versions
- 9a048264 - Remove unused .gitattributes
- 4c193c05 - run changelog generation and version tagging as GitHub Actions
- 5819007b - Update CHANGELOGs
- e6f5cba4 - Bump sv version
- 183dc16b - Bump sv to latest release
- 89949d7a - Bump deps
- 7fb64af2 - Exclude skills dir
- 77c222ac - Various tweaks
- ebcbdc1d - make Taskfile the authoritative source for tasks
- be077bae - Update deps

## [v0.47.1] - 2026-08-18

### Changed
- 08ab9327 - Bump omarchy version
- 435d8b55 - Remove .skillo
- a51e966b - Remove exp/skillo
- 1766ce66 - Scan complete workspace by default
- 53b49e27 - Update deps

### Fixed
- e74bee91 - Bump disk size to allow `omarchy update`

## [v0.47.0] - 2026-08-06

### Added
- 52d9b671 - add awesome-lint CLI tool for linting awesome lists

### Changed
- 0d83ccf5 - Bump action versions in sv-release
- 7c5bb1e1 - Add lint:go-crap task
- 950f7964 - Update go.work.sum
- 5b0ca43a - Update CHANGELOGs
- d4208456 - Bump deps

## [v0.46.0] - 2026-07-19

### Added
- 08b2308d - add fuzzy search TUI with bubbletea v2; fix exit to cleanly propagate

### Changed
- 6417380f - Update go.sum
- e3946493 - Update CHANGELOGs
- 0996a2a8 - Update CHANGELOGs
- b10cc15a - Update CHANGELOGs
- ea8cde3a - Pin trivy version

## [v0.45.0] - 2026-07-18

### Added
- 6ed9977b - Add terebra
- d6ca9388 - add vim-style keys and Ctrl-p/n for menu navigation
- 4949dd07 - add interactive skill selection menu for add and remove

### Changed
- 51f343c6 - Bump go deps
- 317458fd - Add formatters, unify conventional-commit check
- 5139f958 - Tweak go.work
- 0ce2ade3 - Add typstyle lint task
- 6038f013 - Run `typstyle -i .`

### Fixed
- d9733a23 - Use tabs in Earthfiles
- 2d7ab0df - prevent extracted skills from becoming orphaned on partial error
- aed6f603 - ensure skillo writes JSON files with trailing newline
- 3c40256d - prevent skillo from showing project skills as stale after install

## [v0.44.0] - 2026-06-30

### Added
- ff07089d - Add mnemosyne scaffolding (idea phase)

### Changed
- ee6368c0 - Update go.work.sum

## [v0.43.0] - 2026-06-30

### Added
- daebfcce - Add specs skill
- 96fe7a3e - add tree view for list command (-t/--tree)
- 4c772023 - add project-mode skill directories with two-scope architecture

### Changed
- 18798705 - Update deps
- 4f9815c1 - Bump trivy
- b396bc19 - Remove agent-browser skill
- 6d6b8415 - Add .agents/skills to .gitignore (favor skillo sync)
- b8042e1a - Update skillo selections
- dfe50743 - Run `go fix`
- a71db481 - flatten package to exp/skillo root, drop cmd/ subdir
- f937f3c6 - Add .skillo
- 7a372144 - Update go.work.sum
- 581727ff - Run `go fix`
- dafb1370 - Update deps
- 79771f93 - Update all changelogs

### Fixed
- dd5462bd - Quote description to avoid YAML loading error
- a3987139 - improve list output with headers and reliable version lookup
- dc2ef76f - skip invalid skills during extraction instead of aborting
- f1665efa - use go list -m -e -json with module cache fallback
- 4443f76a - make list respect --user and --project flags
- d8456369 - fall back to module cache when go list -m fails for module dir
- 11e02765 - add go mod download before go list -m in add command

## [v0.42.0] - 2026-06-29

### Added
- 8d9efdd2 - add go-gherkin-testing skill
- a9335c65 - enhance CLI with user/project dirs and rich list output
- 6c4543b8 - add skilldirs package and version tracking to manifest

### Changed
- 5baab1c4 - Update go.sum
- fafc522e - add BDD feature tests with godog
- ae73b3c2 - Update all the changelogs

## [v0.41.0] - 2026-06-28

### Changed
- b15c0483 - Bump version to v1.0.1
- 1153dfcc - Bump version to v1.0.0
- c48959a8 - Update all changelogs
- 150c17c7 - Update go tool to use new cmd/sv path
- 9e483f97 - Graduate out of exp(erimental) dir to cmd

## [v0.40.0] - 2026-06-27

### Added
- 75d6dbf3 - add shell completion for -t flag and normalize project_name

### Changed
- 16baed62 - Update deps
- 29681f6c - reimagine pavona as a cookiecutter-inspired template engine
- 6731ec0a - Update all changelogs

## [v0.39.0] - 2026-06-26

### Added
- 7a1c7faa - implement all six Pavona project type scaffolds

### Changed
- 6b44bd2f - migrate from modernc.org/sqlite to ncruces/go-sqlite3
- 1770a026 - Various tweaks
- 5f0ea237 - Bump omarchy version
- 0638b54a - Update changelogs

## [v0.38.0] - 2026-06-26

### Added
- 4da59a85 - add roguelike game set in a French industrial town
- 3963b900 - Add `--write` flag
- 3c8598e0 - add validate-cc subcommand for conventional commit validation

### Changed
- 3db75bd5 - Update deps
- 63db3549 - Update changelogs
- be0fb451 - Add `update` task
- 30ce367e - Bump sv tool version
- 7cb48abd - Add changelogs for all modules via `sv changelog -aw`
- 49cce786 - Update deps
- 34a64f69 - Bump go tool sv to v0.9.0
- 4ff23fe2 - Use sv validate-cc for commit-msg git hook checks
- f91b0d45 - Remove unneeded scripts
- 7b93b891 - Run `go fix work`
- 67b7d44f - Bump sv go tool
- 0cbeca29 - Bump sv go tool
- e48df127 - use go tool sv instead of hardcoded version in workflow
- e36add3a - Bump sv version to v0.8.2

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

## [v0.37.0] - 2026-06-19

### Added
- dfd978e9 - replace --since with --from/--to with date-aware tag generation
- 80391959 - add changelog command for keepachangelog output
- 8741730e - add changelog feature design document
- cb2838cb - add --dry-run flag to `sv next`
- eb6f23dc - add annotated git tag support to `sv next`
- 2ee093a1 - skip retracted Go module versions when calculating next version

### Changed
- 35e69b52 - Run `go fix`
- 84507050 - Remove unused `tag:*` tasks
- efa25bba - Remove chglog tool (unused)
- ba761ae3 - Update go.work.sum
- 91d532aa - Add goimports to linters and bump golangci-lint
- a8cd3eec - Update go deps
- b2cdf599 - document --tag, --tag-format, and --dry-run flags in README
- c1ddc273 - Update sv version and use new `--tag` flag
- 3f071074 - rename --debug flag to --verbose
- 61a57e29 - Add Earthfile for typst templates
- f60057e6 - Bump sv version to v0.6.0
- f4fc5314 - Update go.work.sum
- 4fd1efa9 - Retract v1.0.1 and v1.0.2
- a05e5030 - Tweak makefile to simplify and avoid unnecessary work
- 486dcb8e - Update typst Taskfile and add *.pdf to .gitignore
- f8438f1d - Retract v1.0.1

### Fixed
- 6ed0c4bf - accept string or array for allowed-tools in SKILL.md

## [v0.36.1] - 2026-06-11

### Fixed
- d96f997f - Retract v1.0.0 in go.mod

## [v0.36.0] - 2026-06-11

### Added
- 75e63633 - Add letter template

### Changed
- 101a6474 - Update go.work.sum
- 7cae5fa3 - Update go deps
- 54dd9f3e - Update go deps

## [v0.35.0] - 2026-06-04

### Added
- 0cbc50e5 - Add new Go module exp/pangrams with pangram printer

## [v0.34.1] - 2026-05-25

### Added
- b1e1b863 - add --default-patch flag to next subcommand
- b0a103fc - Remove -v/--verbose flags from subcommands
- 6a33fcef - support repeated and comma-separated --path flag

### Changed
- e9e67498 - Bump sv version
- 35baa6af - Run `task lint:fix`
- 512e7454 - extract CLI subcommands into internal/cli

### Fixed
- 1b19d549 - prevent sv-release failure when no version bumps needed

## [v0.34.0] - 2026-05-25

### Added
- a6763c15 - Add project-level agent-browser skill
- e732c8f8 - Create fc financial calculator with tax pic subcommand
- d7aae705 - Add tax penalties and interest calculator

### Changed
- 7cffc888 - Update go deps
- 20b2e9ca - Update fc references to fin across module
- ac2fe6d8 - Tweaks
- 03ee0abf - Rename pic internal symbols to pi
- 6b6a4075 - Split monolithic internal package into subdirectory packages
- 0b0613a4 - Rename pic subcommand to penalties-and-interest with pi alias
- 2798bdeb - Remove exp/tpi directory for rebrand to fc
- bd2c6dc2 - Move rates package under internal
- 0506b178 - Add testscript CLI integration tests
- 4d3d5a42 - Move CLI commands to internal/cli package
- 97bd4cb2 - Use decimal.Decimal for Money type
- c7bc845e - use boa declarative CLI framework

## [v0.33.9] - 2026-05-24

### Changed
- 8d8e3505 - Bump go.work.sum

## [v0.33.8] - 2026-05-24

### Changed
- 991531a9 - Bump go deps & run `go fix`
- fc94dca9 - Various updates
- b79367a9 - Tweak Taskfile
- 8ccb22df - Run `go fix work`
- ea813144 - Remove .beads
- 5c7dca44 - Add lint:cuefmt
- fdef2e59 - Switch to `test:go` prefixes

## [v0.33.7] - 2026-05-24

### Changed
- b8a1ccb8 - Add `test:build` task
- 51d22a3a - Bump go deps

## [v0.33.6] - 2026-05-24

### Changed
- 612498bb - Remove "root = true"
- 0956ecc8 - Fix mistaken task name collision
- 7f40825c - Bump lefthook
- 8fa03b86 - Add lint tasks
- 307643df - Add `lint` and `test:go-unit` tasks
- 34504e6c - Bump go deps
- 529ab5bc - Bump go deps
- 929dd651 - Add lint:govulncheck task
- 182d149e - Bump golangci-lint & dynamically handle GOTOOLCHAIN

### Fixed
- 741e7cc0 - Bump golang.org/x/net to fix govulncheck vulns

## [v0.33.5] - 2026-05-19

### Changed
- 85d23d61 - Run `gofumpt -w .`

## [v0.33.4] - 2026-05-19

### Changed
- dd162920 - Bump sv version
- 5b9213e1 - Tweak clean task
- 1db5f54c - use boa CLI framework

## [v0.33.3] - 2026-05-19

### Changed
- 77484709 - Pin sv version

## [v0.33.2] - 2026-05-19

### Fixed
- df6f81f8 - correct nested module tag extraction in sv-release

## [v0.33.1] - 2026-05-19

### Fixed
- 9160c48c - set GH_TOKEN in sv-release workflow

## [v0.33.0] - 2026-05-19

### Added
- 4d4a0546 - Add cli-design skill

### Changed
- 6a76c59e - Update go.work.sum
- 7498c2a8 - Update go deps
- f64b5684 - Update go deps
- 2b996f8a - Tweak table formatting
- 8a1160c7 - Update go deps
- 4d600c13 - remove git-tag-semver from lefthook

### Fixed
- 24dedb43 - fix sv-release workflow empty heredoc and tag matching
- d9d8d0ec - show modules with chore-only commits in sv next -a
- 726d4bcd - Remove bento examples

## [v0.32.3] - 2026-04-24

### Changed
- 0cedbafb - remove rsc.io/markdown replace directive
- 6b5561fd - use debug.ReadBuildInfo for version
- 9540e9fd - bump exp/sv to v0.2.0

### Fixed
- 4122a23c - fix sv install path in release workflow

## [v0.32.2] - 2026-04-06

### Added
- 937e5936 - add --version flag using debug.ReadBuildInfo

### Changed
- 814fcb2f - Use sv instead of svu in Taskfile

### Fixed
- 0a787d56 - use commit hash instead of message in file filtering

## [v0.32.1] - 2026-04-06

### Changed
- 3c976e6b - improve test coverage and fix error handling

### Fixed
- 56610dcc - filter commits by module path

## [v0.32.0] - 2026-04-06

### Added
- 011d7d20 - default to v0.1.0 for untagged modules

### Changed
- 35ef18a6 - Bump to latest release
- 444cbe20 - Add .envrc with LEFTHOOK_BIN set to use via `go tool`

## [v0.31.4] - 2026-04-05

### Fixed
- d8d1583b - prefer go tool lefthook for consistent version

## [v0.31.3] - 2026-04-05

### Changed
- 910af1b1 - Update lefthook

### Fixed
- 27422fa9 - remove duplicate lefthook call in pre-push hook

## [v0.31.2] - 2026-04-05

### Changed
- 55fc6cdb - Update beads
- e6433590 - Bump golangci-lint version

### Fixed
- c64fe0ab - Errors from editorconfig-checker

## [v0.31.1] - 2026-04-05

### Fixed
- fedd806e - Only create semver tags on main branch
- 877027a4 - Remove outdated warning expectation in get test

## [v0.31.0] - 2026-04-04

### Added
- cb1baaac - Add testing skill
- a8ba51a5 - Add pyramid skill (initial draft)
- ef89d9a6 - Add x/skillo
- d9748250 - Add some typst templates
- 4f273760 - automate sv versioning/tags/releases via scripts
- d887b140 - implement semantic versioning tool with monorepo support
- ab55e46b - set gradient-dark as the default theme
- 058bf58e - implement distinct h2 gradients for Keynote themes
- 36e2aaac - expand gradient themes for high-fidelity Keynote look
- c7a8d08a - harmonize h1/h2 gradients for Keynote themes
- a7b6cb6d - refine gradient theme palettes and directions
- 768be494 - implement Keynote-style gradient themes
- ef8e1aa4 - center current theme in palette on open
- 8eb5605a - apply thematic accents and dynamic backgrounds
- b8e4766e - enhance pause mode timer and UI visibility
- 60ffbe80 - overhaul pause mode and command palettes
- b69decaf - refactor font themes to use CSS as source of truth
- 60d51759 - implement orderless-style search with comprehensive tests
- 3cc7548d - Implement comprehensive search palette fixes and enhancements
- 073f0989 - add feature overview example
- 1e899c1b - final refinements and formatting fixes
- 129d2db3 - finalize implementation with formatting fixes
- 83f7434b - implement CLI and core features
- 74f3f950 - initialize project with documentation and tools

### Changed
- b9925414 - Fix broken URL
- db1cdcf9 - Go module updates
- ba35000a - Tweaks
- d111bcf4 - Bump go deps
- 9871c72f - Bump cue mod version
- 4f75421b - Add exp/README.md
- d175d000 - Use top-level variable for golangci-lint version
- 614e7a8e - Rename selfhosting/ directory to infra/
- a7d247e8 - Improve sv test coverage and fix code quality
- 844a2f69 - Rename x/ directory to exp/
- 9b5d7909 - Remove cruft
- 414dd73f - Update go.work.sum
- efeafa48 - Simplify output format and add verbose flag
- 0301281f - Tweak Taskfile
- 6c0c0481 - Move main.go to top-level
- 066a5fb5 - Bump go version to 1.26.0
- d922008a - Move testing guidance to tester skill
- 4822da77 - Bump golangci-lint version
- d34d7180 - add tailwindcss to .gitignore
- f5db047d - align test suite with AGENTS.md and unit testing principles
- 46748969 - align slidesdeck and sv testing with AGENTS.md guidance
- dd411f1a - Update Taskfile
- 16b67c59 - add module definitions and git attributes
- 7bc3f15a - initialize implementation plan for sv
- c585192c - add product vision and automation documentation
- b0985922 - initialize sv module and workspace integration
- b5247265 - update issue tracking notes and bd conventions
- 80f8cd20 - expand overview examples and add keyboard styling
- 56a7ff26 - Tweak build task
- 5c75d8be - Move css-split-themes to its own step
- 34b9c1a1 - improve help screen formatting
- b6d48b33 - Add install task
- 6ec6bafc - update go.work.sum with slidesdeck dependencies
- 0932b4a0 - update build config and dependencies
- cef3cdff - add comprehensive unit and integration tests
- c4d0da8c - remove tailwindcss binary from repo
- 33bb5e0f - Update go deps

### Fixed
- 3e3839c7 - Set GOTOOLCHAIN to fix test failures
- cd63aa5c - Add required front matter
- 59f3cbee - update integration tests for first slide title logic
- 226968e3 - remove unknown @property rule from vendor CSS
- e91610f7 - resolve heading layout issues in Markdown
- b2a29d8c - update gradient-dark h2 color palette
- fe98e396 - correct heading level mapping for Org-mode
- 86120500 - intensify h2 color transitions in gradient themes
- 3d83d747 - ensure full gradient gamut is visible on headings
- 6126f00c - restore heading sizes for gradient themes
- 899ef5a5 - resolve black screen bug and empty leading slides
- 0ee1c3a5 - make slide splitting block-aware and fix org-mode warnings

## [v0.30.0] - 2026-02-03

### Added
- 8c9a0b90 - Add scraper example

### Fixed
- a74e978c - resolve test failures and improve type checking

## [v0.29.2] - 2026-02-01

### Fixed
- 8a1ffdfa - pass Counter by pointer in concurrency test
- 8803649e - improve testing interop and function return logic

## [v0.29.1] - 2026-02-01

### Changed
- 5ab531bf - consolidate examples and tests into standard _test.elb format

### Fixed
- 6c8a7610 - improve code generation for polymorphic binary operators and expressions

## [v0.29.0] - 2026-02-01

### Added
- 86d442b9 - enhance testing ecosystem and concurrency interop
- c22d86a6 - implement support for Go testing ecosystem

### Changed
- 5556f1f0 - Remove cruft

## [v0.28.0] - 2026-02-01

### Added
- d1f20756 - add version subcommand

## [v0.27.0] - 2026-02-01

### Added
- 7097c3dd - implement v0.2 core features and comprehensive testing suite

## [v0.26.1] - 2026-02-01

### Fixed
- e8dfd2c6 - normalize outpaint mode values in shader

## [v0.26.0] - 2026-02-01

### Added
- 8281d9aa - allow inpaint and outpaint modes to be active simultaneously

### Fixed
- af1e97e8 - prevent Ctrl+num from triggering fractal switches in fractalis
- 38e652c9 - correct paint mode conditions in shader for fire inpaint

## [v0.25.0] - 2026-02-01

### Added
- c859c062 - add realistic fire inpainting and outpainting modes
- 1393d064 - add inpainting and outpainting effects to fractalis ebiten engine

### Fixed
- 34107653 - correct paint mode conditions in shader for fire inpaint

## [v0.24.0] - 2026-02-01

### Added
- a7d3545b - create jmacs - Emacs text editor in Elbereth
- e1316b11 - add fullscreen CLI flag to fractalis ebiten engine

### Fixed
- 17db0728 - Correct 2D fractal rendering path detection
- a1da4d2d - Center 2D fractals correctly for non-square viewports

## [v0.23.0] - 2026-02-01

### Added
- 5b82edb7 - enhance Elbereth REPL with evaluation and better CLI
- 2590455c - add interactive REPL for Elbereth language

### Changed
- aad179df - Remove binary

## [v0.22.1] - 2026-01-31

### Fixed
- 87f6ca4d - reduce Ebiten movement speeds to match bubbletea behavior

## [v0.22.0] - 2026-01-31

### Added
- e204819a - enhance vantage mode with jump and 2D support in Ebiten

## [v0.21.0] - 2026-01-31

### Added
- 2c266dfa - formalize rendering engines as 'bubbletea' and 'ebiten'

### Fixed
- f9d4596e - respect initial config and 2D defaults in Ebiten engine

## [v0.20.0] - 2026-01-31

### Added
- 56bd40ba - implement emulated double precision in 2D shader
- 5392c8f8 - add embedded web server for WASM fractal viewing
- bcf04e83 - implement high-precision 2D fractal engine and autopilot in Ebiten

## [v0.19.0] - 2026-01-31

### Added
- 41bda308 - add color mode toggle and default to color in terminal mode
- 0b1743a1 - make Mandelbox default 3D view and enhance vantage points
- 1f761171 - add Mandelbox fractal and implement mouse look controls

### Changed
- 0fba2548 - overhaul CLI flags and improve help menu organization
- e955e6c8 - split mandelbox and mandelbulb shaders into separate files
- 9ce0e6e3 - rename cmd/3d to cmd/fractalis-ebiten-wasm
- 1d6816e0 - Update go deps

## [v0.18.0] - 2026-01-31

### Added
- 1fd79a2b - Implement GPU-accelerated 3D fractal viewer using Ebiten

### Changed
- 522b4813 - Modularize core logic and reorganize project structure
- ab943166 - Update dependencies and project structure
- d46a1052 - Bump go deps

## [v0.17.1] - 2026-01-30

### Changed
- f61ec6b6 - Add spaces
- c479e9ac - Update AGENTS.md
- 2a00a613 - Rename ai -> crush

## [v0.17.0] - 2026-01-30

### Added
- 19a8c9d4 - Add AGENTS.md
- 2284f008 - Add AGENTS.md (WIP)

### Changed
- 3108d264 - Tweak formatting
- 8f4556b9 - Update go.work.sum
- f93e991f - Add top-level test target
- 69648aeb - Bump golangci-lint version
- 7cb8ec5c - Add test coverage to Task test
- b9398e85 - Update deps
- fe5dd80e - Tweak Taskfile
- 421cf59d - Bump deps

## [v0.16.0] - 2026-01-29

### Added
- 40cbf79b - Rename x/fractals -> x/fractalis

### Changed
- 6600e360 - Shorten quote
- ae3b8716 - Tiny simplification
- 57e62871 - Update .gitignore
- d8c57113 - Various perf improvements

## [v0.15.0] - 2026-01-21

### Added
- 1eeb9bc1 - Add vantage mode

## [v0.14.0] - 2026-01-21

### Added
- 452939f8 - Add `-a` / `-autopilot` cli flag

### Changed
- 3c8ae0c1 - Rename SRS.md -> SPEC.md
- 7e11abb2 - Add reinstall task
- 1f9b9dee - Move url handling to persistence package
- 7bdb1591 - Add SRS
- c916d9fe - Move bookmarks and screenshots to persistence

## [v0.13.1] - 2026-01-21

### Changed
- 3eb56a4b - Bump go deps
- e786c345 - Add test_go_tools script
- b45ae01e - Update deps

### Fixed
- 2d0ffea5 - Avoid timeout on ci `check-links` task (skip catb.org)

## [v0.13.0] - 2026-01-20

### Added
- 4f7cad6e - Add breakthrough transition

### Changed
- 6ce380aa - Various updates
- 2e4076e9 - Modularize transitions
- 62d3430f - Modularize fractal logic into individual files
- 616a6bf4 - Bump go version
- c7277bda - Update deps

## [v0.12.0] - 2026-01-18

### Added
- e31a6cc0 - Add x/fractals

## [v0.11.3] - 2026-01-18

### Changed
- e5ea5a85 - Update deps
- f7f9add8 - Remove old cruft
- 7621ecef - Bump go deps
- ddb5abb5 - Bump go version
- 00ff3b98 - Update go.work.sum
- 25ce4e7f - Tweak spacing
- a00e9b65 - Add header-length commit-msg hook
- 987d531a - Add age as go tool
- 6fb9035f - Use `integration` tag with test task
- 9b7eeffa - Update tool deps

## [v0.11.2] - 2025-11-15

### Changed
- 44a2206d - Avoid recursive lefthook run
- 9bd6d3eb - Add clean task
- f0bdcf61 - Add check-links job and fix git-tag-semver
- e201eb52 - Update go.work.sum
- ae4296cb - Add test task

### Fixed
- 27f0b333 - Re-enable and improve output for crypto block
- 8fb938c0 - Linter errors

## [v0.11.1] - 2025-11-15

### Changed
- 8266e0a1 - Add bin/ to .gitignore
- a8696ca7 - Simplify Taskfile.yml
- 5f7c67bf - Update go.mod
- d2bf7b16 - Update go modules
- 621b6814 - Add `ai` task (use crush)
- 467b0966 - Add .dir-locals.el
- 2b173142 - Various lefthook updates
- f3f1fc6d - Update go.work.sum
- 34519bce - Run cue fmt
- ce0cecc8 - Bump go version
- c86c53d4 - Add golangci-lint target
- 2ab14fb0 - Various lefthook updates
- f4499b7e - Update deps
- 42de22f4 - Bump omarchy version
- 4c9a6a69 - Add .crush to .gitignore

### Fixed
- 6c2844e8 - Add missing logo.jpg
- 8026234d - Go Linter errors (WIP)
- 5645ca42 - Run gofumpt
- 9f8d4f3a - Linter errors

## [v0.11.0] - 2025-10-25

### Added
- 11030faf - Add selfhosting dir (WIP)

### Changed
- 51c26a51 - Update deps

## [v0.10.0] - 2025-09-21

### Added
- 8adbf437 - Add joliv-spark example tests

## [v0.9.1] - 2025-09-21

### Changed
- 170d52dc - Update go.work.sum
- ad500111 - Add FIXME comment
- ea94eab0 - Move dep to new line
- 4a0279a8 - Taskfile categories: lint, security, githook
- 4ece361e - Bump go deps
- ece0745d - Add Arcadia paintings and descriptive paragraph to README

### Fixed
- 3724784b - Remove broken ed2k drive references

## [v0.9.0] - 2025-09-19

### Added
- 8aa83b47 - Add lefthook config

### Changed
- aa9c3cfa - Reorganize test tasks

## [v0.8.2] - 2025-09-18

### Changed
- ec34e745 - Bump trivy version

### Fixed
- c0b93bda - Trivy vulnerabilities via go mod updates

## [v0.8.1] - 2025-09-18

### Fixed
- 9c1dc1cf - Taskfile indentation

## [v0.8.0] - 2025-09-18

### Added
- 80a13aed - Add x/testdrive-omarchy

### Changed
- 2ed6d260 - Update .gitignore
- 2b7d084e - Add tag dep to tag:push task

## [v0.7.0] - 2025-09-15

### Added
- b52b5c02 - Move horeb from x/ to cmd/

### Changed
- 1348a335 - Update go.work.sum
- a22c2670 - Use stdlib slog
- 98e88f6b - Add doc comment
- 4d326044 - Run default tasks in parallel
- 574eb956 - Bump deps

### Fixed
- a9a97ca3 - Re-add goreleaser tool to fix broken deps

## [v0.6.1] - 2025-09-05

### Fixed
- 806711a8 - Add cluster.local to xurls exclude list

## [v0.6.0] - 2025-09-05

### Added
- 1bce2c92 - Update cue examples

### Changed
- fb1e490d - Delete convert cue to ts example (deprecated tool)
- 5961f13d - Update go.work.sum

## [v0.5.0] - 2025-09-04

### Added
- ec3eff98 - Improve generate-cue-from-go example
- ad976811 - Update generate cue from cue example

### Changed
- 7b1c12a1 - Update deps
- 382c08a5 - Add .gitignore
- ba9a60d3 - Replace Makefile with Taskfile

## [v0.4.1] - 2025-09-02

### Changed
- 39d9e1ab - Tweak lint:check-links
- c3494709 - Bump deps
- 687c1e27 - Disable broken cmd (govulncheck)
- ff3011b8 - Replace single-word msg t.Logf with t.Attr

### Fixed
- 17c80772 - Make link-check pass
- 7046a72c - Broken links

## [v0.4.0] - 2025-08-27

### Added
- 592e56bd - Add more linting tasks

### Changed
- 08cdcd26 - Migrate to go.work with submodules
- 7ad914ce - Move benthos to bento
- 6a38c520 - Run go mod tidy
- c9558d66 - Update go deps

### Fixed
- 4eb186a2 - Use bento for benthos plugin example

## [v0.3.0] - 2025-08-22

### Added
- 6bef1285 - Add examples/go/thirdparty/rod

## [v0.2.1] - 2025-08-22

### Fixed
- 6d0aa925 - Allow `tag:push` to be called along with `tag`

## [v0.2.0] - 2025-08-22

### Added
- 492cc704 - Add ox-typst example
- 04ce8d00 - Add link checking

### Changed
- beb120a2 - Add lint:check-links task
- 496e2ce6 - Replace Makefile with Taskfile
- 7b488f40 - Update go.mod
- b6035868 - Add helm and k3d go tools
- 75fc1f8a - Add grafana helm chart dir to .gitignore
- 0fca6fcf - Add helm `charts` to editorconfig-checker exclusions
- 565fd128 - Bump deps
- 67a563c1 - Bump go.mod deps

### Fixed
- a119e3d6 - Re-import grafana helm chart values into cue (fixes ec-lint)

## [v0.1.3] - 2025-08-14

### Changed
- be1eef83 - Add tag task status test

### Fixed
- a9bc4785 - All integration test errors

## [v0.1.2] - 2025-08-14

### Changed
- e6747a64 - Tweak tag:dryrun
- cbd0e733 - Update go.mod deps
- 06ddbda7 - Bump go version to 1.25.0

## [v0.1.1] - 2025-08-11

### Changed
- d477191a - Update go.mod
- c6a94af7 - Use markdown for all READMEs
- a47935f6 - Remove go module vendoring

### Fixed
- c16b7e46 - Use explicit `main` alias for main_test packages (for goimports)

## [v0.1.0] - 2025-08-11

### Added
- 6cb63872 - Add go-native wifi package
- 8abfaf41 - Add sliders skeleton
- be7d2d6c - Add horeb
- 6f514add - Add benthos examples
- 20d49aa2 - Add scroll examples
- dedd81d7 - Add talks
- 98a9ba2d - Vendor go modules
- 96b0006c - first commit

### Changed
- 8ec6d314 - Add go mod and tag-related tasks
- 23fad060 - Use mdlayher/wifi instead of scraping iw output
- e9a9867a - Tweak Taskfile
- 6f5f142a - Add tag task
- 3bf34007 - Update deps and vendor dir
- 91b2a2b7 - Add Go Reference badge
- 96637739 - Update vendored modules
- a1bc8ccb - Tweak README
- 20027c03 - Move horeb main package to top-level
- 25cd08f4 - Run gofumpt
- 33d649bd - Remove unused cruft
- b3996bd9 - Use preferred formats
- 4c4890f6 - Add links to top-level README
- 42fae386 - Add test:gotestsum task
- e8dd2a40 - Move benthos examples up one level
- 0c431c20 - Update vendored modules
- a95e0598 - Update go modules
- 4fef937f - Run cue fmt
- cf9b8abc - Add lint:editorconfig to default task

### Fixed
- 089015f8 - Editorconfig errors
- b7144c9d - Format link in org
- 1d6b99d7 - Scroll cruft
- 465bea0e - Use correct path
- a308a8f3 - Remove unused submodule
- 566b563f - Get horeb main branch
- 8cd14e98 - Editorconfig errors
