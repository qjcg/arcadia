# Mnemosyne — Product Requirements Document

## 1. Product Overview

Mnemosyne is a CLI tool and Go library for AI agent memory management,
backed by NATS KV. It enables agents to store, retrieve, search, and
prune memories across sessions in a distributed, fault-tolerant manner.

## 2. Target Users

- **AI agent developers** building multi-step or multi-session agents
- **Agent orchestration platforms** needing shared memory across agents
- **DevOps/SRE** managing agent infrastructure on NATS

## 3. User Personas

| Persona | Needs |
|---------|-------|
| **Agent Developer** | Simple API to store/recall what an agent learned or did |
| **Platform Engineer** | Multi-tenant isolation, monitoring, bulk operations |
| **Researcher** | Experiment with memory decay curves, importance models |

## 4. Success Metrics

- **Time to first memory**: < 30 seconds from `mnemosyne init` to `put`
- **P99 latency**: < 10ms for get/put against local NATS server
- **Integration surface**: Single `go get` for Go agents; CLI for any language
- **Zero data loss**: KV bucket replication + history provides durability

## 5. Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| F1 | Store memory entries with type, content, tags, importance | P0 |
| F2 | Retrieve entries by ID or filter (agent, type, tags, importance) | P0 |
| F3 | List entries with pagination, sort, and filter | P0 |
| F4 | Delete entries by ID | P0 |
| F5 | Watch for real-time memory changes | P1 |
| F6 | Search entries by text content (fuzzy matching) | P1 |
| F7 | Prune expired or low-importance entries automatically | P1 |
| F8 | Export/import memories in JSON/JSONL format | P2 |
| F9 | Initialize and configure new namespaces/buckets | P0 |
| F10 | Show bucket status and health | P1 |
| F11 | Authenticate via NATS credentials, NKEY, or JWT | P1 |

## 6. Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| N1 | CLI startup time | < 100ms |
| N2 | Max memory per process | < 50MB |
| N3 | JSON output always parseable | All commands support `--json` |
| N4 | Help text includes examples | Each subcommand |
| N5 | Error messages include cause + fix suggestion | All user-facing errors |
| N6 | Exit codes follow conventions | 0=ok, 1=err, 2=not found, 3=invalid input |

## 7. Release Criteria

- All P0 and P1 commands implemented
- `mnemosyne put | get | list | search | forget` fully functional
- Testscript integration tests for all commands
- `--help` output reviewed for clarity per clig.dev guidelines
