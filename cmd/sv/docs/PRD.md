# Product Requirements Document: sv

## 1. Introduction
`sv` is a CLI tool designed to automate semantic versioning in monorepos. It extends the philosophy of `caarlos0/svu` by supporting independent versioning for multiple modules within a single repository using path-based tagging.

## 2. Target Personas
- **Monorepo Maintainers**: Developers managing multiple internal libraries or services in a single Git repository.
- **CI/CD Engineers**: Automating release pipelines where different modules release at different cadences.

## 3. Goals
- Automate next-version calculation based on Conventional Commits.
- Support path-prefixed tags (e.g., `pkg/math/v1.0.0`).
- Isolate version increments to the specific module being changed.
- Maintain high compatibility with `svu` CLI patterns.

## 4. Features
### 4.1 Path-Scoped Discovery
The tool must identify the current module's root and filter Git tags that start with that path.

### 4.2 Automated Next Version
Calculate the next semantic version by analyzing commits limited to the module's subdirectory since the last module-specific tag.

### 4.3 Monorepo Awareness
Support a `--recursive` or `all` command to identify all modules with pending changes across the entire repository.

### 4.4 Go Module Integration
Automatically detect module boundaries via `go.mod` files, using the relative directory as the tag prefix.

## 5. Success Metrics
- 100% accurate calculation of next versions for overlapping module paths.
- Execution speed comparable to `svu` on large repositories.
