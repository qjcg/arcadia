# Software Requirements Specification: sv

## 1. Functional Requirements
### FR-1: Tag Parsing
The system SHALL parse Git tags and identify version strings following the pattern `[path/]vMAJOR.MINOR.PATCH`.

### FR-2: Commit Filtering
The system SHALL filter Git log output to include only commits that modified files within the target module's directory.

### FR-3: Conventional Commit Analysis
The system SHALL interpret commit subjects (feat, fix, breaking change) according to the Conventional Commits 1.0.0 specification to determine version bumps.

### FR-4: Prefix Management
The system SHALL support custom prefixes and automatically handle path-based prefixes for monorepo modules.

## 2. Non-Functional Requirements
### NFR-1: Performance
Version calculation for a module in a repo with 10k+ commits SHOULD take less than 200ms.

### NFR-2: Portability
The tool SHOULD be distributed as a single static binary for Linux, macOS, and Windows.

## 3. Interface Requirements
### IR-1: CLI Commands
- `sv current`: Show the latest tag for the current module.
- `sv next`: Calculate and print the next version.
- `sv major|minor|patch`: Force a specific increment.
