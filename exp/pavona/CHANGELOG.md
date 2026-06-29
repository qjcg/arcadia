# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Changed
- 3c164cdf - Update changelog
- 7596abf0 - Add BDD feature tests with godog
- fd4726e2 - move built-in templates to top-level templates directory

## [v0.2.0] - 2026-06-27

### Added
- 75d6dbf3 - add shell completion for -t flag and normalize project_name
- 0c4c512c - use CUE type syntax for config.cue and bubbletea v2 for prompts

### Changed
- 16baed62 - Update deps
- 6d714541 - Run go fix
- 29681f6c - reimagine pavona as a cookiecutter-inspired template engine
- 138d51bd - Add README.md for pavona scaffold and framework
- 6731ec0a - Update all changelogs

## [v0.1.0] - 2026-06-26

### Added
- 25d4efdc - replace seed SQL with embedded migrations
- 1fd37001 - seed database with ~100 demo entries on first start
- b0b23f7c - make links table filterable and sortable with Alpine.js
- fe085f71 - allow optional custom short code in URL shortener
- 8fefd8ab - persist QR codes in DB and upgrade to Tailwind 4 + DaisyUI 5
- 8a1ffd38 - add --demo flag scaffolding a URL shortener with expiry and QR
- 82f8765c - scaffold full-stack web app with templ, SQLite, HTMX
- 7500471d - make --format flag exclusive to `pavona new site`
- 049dd9ac - add hamburger menu for mobile navigation
- 2967013e - add Ctrl/Cmd+K shortcut to focus search bar
- a91cc9b4 - add full-text search via FlexSearch to default site theme
- 0f8574d8 - add live reload to dev server via fsnotify
- ace8db7d - add --name flag for site title, remove welcome page from nav
- f8099277 - add dark mode toggle with DaisyUI 5 to site templates
- acbeba20 - preserve -p flag page order in navbar for scaffolded sites
- 6a3a1f8a - add --pages flag for site scaffold with BDD tests
- 1ca16bab - add brace expansion engine and site scaffold pages
- 61e28973 - add build.go and go.mod templates to site scaffold
- ba877e21 - add templ-based site theming with exported pkg/site
- 2fc72084 - implement comprehensive site builder BDD test steps
- 2115e16b - add tree-based site navigation with frontmatter support
- b57298b6 - implement site theme system with beautiful default template
- 54494979 - auto-detect content format in site builder by file extension
- c7b0fb7d - implement static site builder with markdown and org-mode support
- 7a1c7faa - implement all six Pavona project type scaffolds

### Changed
- 6b44bd2f - migrate from modernc.org/sqlite to ncruces/go-sqlite3
- 442c3aa5 - add trailing newlines to 9 template files for editorconfig
- 92a0e1cc - swap skip2/go-qrcode for yeqown/go-qrcode
- 771f6682 - Rename pavona_test.go -> main_test.go
- a7ba5700 - strip site scaffold to just content files
- 42f99077 - update design doc and feature tests for new site features
- d081ec0e - Add final newlines
- c0d2f1db - apply go fix modernizations to frontmatter parser
- 0ec72530 - move site package from internal to public pkg/
- 37329fc6 - add filesystem-as-URL sections with frontmatter and ordering
- 47435e95 - clean up stale test directory
- a28692ff - add site theme system with templ to design and feature specs
- f0fdc41f - clean up editor backup files
- e1be3248 - remove stale test directory
- 8c7aefc2 - add org-specific HTML assertions and bold/emphasis rendering test
- 1d98055f - expand BDD coverage for scaffolding and site builder
- 42cc530f - remove stale test directory foo/
- dd0ee4c2 - add Taskfile.yaml for pavona
- 8d87d2df - move scaffold to internal/, remove empty testdata/
- 859a956c - move pavona CLI commands from cmd/ to internal/cli
- c95ac85b - replace cobra with boa for Pavona CLI
- 57ed998f - reorganize Pavona features into one file per domain
- 5c5dacd4 - add BDD build plan for Pavona implementation
- f49f103b - polish Pavona design doc for consistency
- 9ce6ea01 - add godog/gherkin BDD testing to Pavona design
- db6d08f0 - polish Pavona design document for coherence
- 2e7119b2 - add org-mode format option to Pavona static site type
- c20c500c - add NATS Agent Protocol project type to Pavona
- 55fd5417 - add static site as fifth Pavona project type
- 4c3c2552 - add 10 design principles to Pavona philosophy
- b3d168bf - add extensible template system and NATS template to Pavona design
- 61f3c32d - add Pavona framework design document

### Fixed
- 579f9104 - prevent seed data scan failure from DATETIME type mismatch
- 86fca57e - check migration and query errors, inline Alpine init
- 0a785264 - register migration files in scaffold embed and fix name clash
- e9be77ea - wire Alpine table store to window.__LINKS__ and HTMX callback
- 835580ac - open QR preview on click instead of hover
- 7fff5c6e - wire garbage can delete button to server-side DELETE handler
- 25ee0573 - remove no-op append to fix go vet warning
- f115f5e2 - update Go module path from bolocera/pavona to qjcg/arcadia/exp/pavona
- e886f1d6 - match test assertion to theme template with attributes
- a7a43607 - disable auto-generated table of contents in org-mode output
- a64cd6fd - use correct org-mode syntax in site scaffold template
