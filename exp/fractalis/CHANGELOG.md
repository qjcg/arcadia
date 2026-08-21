# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 878b20c2 - Update deps
- 04bb597d - Update deps
- 53b49e27 - Update deps
- d4208456 - Bump deps
- 51f343c6 - Bump go deps
- 453a8cfc - Update go deps
- dafb1370 - Update deps
- 79771f93 - Update all changelogs
- ae73b3c2 - Update all the changelogs
- c48959a8 - Update all changelogs
- 6731ec0a - Update all changelogs
- 6f2d4ba3 - Update wasm
- 7cb48abd - Add changelogs for all modules via `sv changelog -aw`
- 49cce786 - Update deps
- a8cd3eec - Update go deps
- 7cae5fa3 - Update go deps
- 54dd9f3e - Update go deps

## [v0.1.4] - 2026-05-24

### Changed
- 991531a9 - Bump go deps & run `go fix`
- 8ccb22df - Run `go fix work`

## [v0.1.3] - 2026-05-24

### Changed
- 51d22a3a - Bump go deps

## [v0.1.2] - 2026-05-19

### Changed
- 85d23d61 - Run `gofumpt -w .`

## [v0.1.1] - 2026-05-19

### Changed
- 78aadba0 - Update fractalis wasm

## [v0.1.0] - 2026-04-06

### Added
- 8281d9aa - allow inpaint and outpaint modes to be active simultaneously
- c859c062 - add realistic fire inpainting and outpainting modes
- 1393d064 - add inpainting and outpainting effects to fractalis ebiten engine
- e1316b11 - add fullscreen CLI flag to fractalis ebiten engine
- e204819a - enhance vantage mode with jump and 2D support in Ebiten
- 2c266dfa - formalize rendering engines as 'bubbletea' and 'ebiten'
- 56bd40ba - implement emulated double precision in 2D shader
- 5392c8f8 - add embedded web server for WASM fractal viewing
- bcf04e83 - implement high-precision 2D fractal engine and autopilot in Ebiten
- 41bda308 - add color mode toggle and default to color in terminal mode
- 0b1743a1 - make Mandelbox default 3D view and enhance vantage points
- 1f761171 - add Mandelbox fractal and implement mouse look controls
- 1fd79a2b - Implement GPU-accelerated 3D fractal viewer using Ebiten
- 40cbf79b - Rename x/fractals -> x/fractalis
- 1eeb9bc1 - Add vantage mode
- 452939f8 - Add `-a` / `-autopilot` cli flag
- 4f7cad6e - Add breakthrough transition
- e31a6cc0 - Add x/fractals

### Changed
- 844a2f69 - Rename x/ directory to exp/
- 066a5fb5 - Bump go version to 1.26.0
- 0fba2548 - overhaul CLI flags and improve help menu organization
- e955e6c8 - split mandelbox and mandelbulb shaders into separate files
- 9ce0e6e3 - rename cmd/3d to cmd/fractalis-ebiten-wasm
- 1d6816e0 - Update go deps
- 522b4813 - Modularize core logic and reorganize project structure
- ab943166 - Update dependencies and project structure
- fe5dd80e - Tweak Taskfile
- 421cf59d - Bump deps
- ae3b8716 - Tiny simplification
- d8c57113 - Various perf improvements
- 3c8ae0c1 - Rename SRS.md -> SPEC.md
- 7e11abb2 - Add reinstall task
- 1f9b9dee - Move url handling to persistence package
- 7bdb1591 - Add SRS
- c916d9fe - Move bookmarks and screenshots to persistence
- 3eb56a4b - Bump go deps
- 6ce380aa - Various updates
- 2e4076e9 - Modularize transitions
- 62d3430f - Modularize fractal logic into individual files
- 616a6bf4 - Bump go version
- c7277bda - Update deps

### Fixed
- e8dfd2c6 - normalize outpaint mode values in shader
- af1e97e8 - prevent Ctrl+num from triggering fractal switches in fractalis
- 38e652c9 - correct paint mode conditions in shader for fire inpaint
- 17db0728 - Correct 2D fractal rendering path detection
- a1da4d2d - Center 2D fractals correctly for non-square viewports
- 87f6ca4d - reduce Ebiten movement speeds to match bubbletea behavior
- f9d4596e - respect initial config and 2D defaults in Ebiten engine
