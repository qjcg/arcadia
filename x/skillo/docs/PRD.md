# Product Requirements Document (PRD)

## 1. Overview
**skillo** enables Go-native management of agent skills.

### 1.1 Goals
- Version skills via `go get` semantics.
- Extract to agent-readable dir.
- Cross-platform binary.

### 1.2 Users
- AI agent CLI users (Crush, etc.).
- Skill authors (GitHub repos).

### 1.3 Success Metrics
- Installs 10+ skills/min.
- Validates SKILL.md YAML.
- 100% test coverage.

## 2. Features
| Priority | Feature           | Description                          |
|----------|-------------------|--------------------------------------|
| P0       | `skillo get`      | `go get repo@v` + extract SKILL.md/* |
| P0       | `skillo init`     | `go mod init ~/.skillo`              |
| P1       | `skillo search`   | GitHub API + local grep              |
| P1       | `skillo update`   | `go get -u` + re-extract             |
| P2       | Multi-skill repos | Extract all dirs w/ SKILL.md         |

## 3. Non-Functional
- Single binary (&lt;10MB).
- Go 1.21+.
- Cobra CLI.

## 4. Out of Scope
- Private repos (add later).
- Non-Git VCS.
