# awesome-lint

**Lint an Awesome list for compliance with [awesome.re](https://awesome.re) guidelines.**

A Go port of [sindresorhus/awesome-lint](https://github.com/sindresorhus/awesome-lint). Checks that your awesome list follows the
[awesome manifesto](https://github.com/sindresorhus/awesome/blob/main/awesome.md) — proper heading structure, a valid badge,
correctly formatted list items, and more.

## Install

```sh
go install github.com/qjcg/arcadia/cmd/awesome-lint@latest
```

## Usage

```sh
# Lint README.md in the current directory
awesome-lint

# Lint a specific file
awesome-lint my-awesome-list.md

# Output results as JSON
awesome-lint --json

# Specify a different file
awesome-lint --filename path/to/README.md
```

### Exit codes

| Code | Meaning                                |
|------|----------------------------------------|
| 0    | No errors found                        |
| 1    | Errors were found or an error occurred |

## Rules

| Rule ID                | Description                                                                                                                            |
|------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `awesome-heading`      | Main heading must exist, be depth 1, and use title case. Only one level-1 heading allowed.                                             |
| `awesome-badge`        | Official Awesome badge (`[![Awesome](https://awesome.re/badge.svg)](https://awesome.re)`) must be present next to the main heading.    |
| `awesome-list-item`    | Each list item must have a valid URL link, a dash separator (` - `), and a description ending with proper punctuation (`.`, `!`, `?`). |
| `awesome-license`      | License section must not appear in the readme (GitHub handles license detection).                                                      |
| `awesome-no-ci-badge`  | CI badges (Travis CI, CircleCI) must not appear in the readme.                                                                         |
| `awesome-contributing` | `contributing.md` (or `.github/contributing.md`) must exist and be non-empty.                                                          |
| `awesome-toc`          | If a Table of Contents exists, it must be the first section. ToC links are validated against actual headings.                          |
| `double-link`          | Duplicate links in the document are flagged.                                                                                           |
| `awesome-spell-check`  | Checks for common misspellings and incorrect technology names (200+ rules from upstream awesome-lint).                                 |
| `definition-case`      | Definition labels (`[label]: URL`) must be lowercase.                                                                                  |
| `no-repeat-item-in-description` | List item descriptions must not start by repeating the item name.                                                             |

## GitHub Actions

```yaml
name: CI
on:
  pull_request:
    branches: [main]
jobs:
  awesome-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7
        with:
          go-version: stable
      - run: go run github.com/qjcg/arcadia/cmd/awesome-lint@latest
```
