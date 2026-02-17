# Skillo

[![Go](https://img.shields.io/badge/Go-1.21%2B-blue?logo=go)](https://golang.org)
[![Go Report](https://goreportcard.com/badge/github.com/qjcg/arcadia/x/skillo)](https://goreportcard.com/report/github.com/qjcg/arcadia/x/skillo)

`skillo` is a Go CLI for discovering, installing, and managing [Agent
Skills](https://agentskills.io/) using Go modules for versioning Git
repos (pseudo-versions/tags supported, no `go.mod` required in
skills).

## Quickstart

```bash
skillo init                    # Setup ~/.skillo/go.mod
skillo get github.com/user/pdf-processing@v1.0.0
skillo search pdf              # Local + GitHub
skillo update                  # go get -u + re-extract
```

Skills extracted to `~/.config/agents/skills/` for agent consumption (e.g., Crush).

## Install

```bash
go install github.com/qjcg/arcadia/x/skillo/cmd/skillo@latest
```

See [docs/PITCH.md](docs/PITCH.md), [docs/PRD.md](docs/PRD.md).
