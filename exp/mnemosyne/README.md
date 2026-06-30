# Mnemosyne

> Your agents never forget.

**Mnemosyne** is a CLI tool and Go library for AI agent memory, backed by
[NATS KV](https://docs.nats.io/nats-concepts/jetstream/key-value-store).

Store, retrieve, search, and prune structured memories — episodic, semantic,
procedural, or working — across distributed NATS infrastructure.

## Quickstart

```bash
# Start NATS (if you don't have one running)
nats-server -js &

# Initialize a namespace
mnemosyne init my-agents

# Store a memory
mnemosyne put \
  --agent agent-1 \
  --type episodic \
  --content "User prefers JSON responses" \
  --tags preference,user-42

# Retrieve it
mnemosyne get --agent agent-1 mem_01J2XYZ...

# List recent memories
mnemosyne list --agent agent-1 --limit 10

# Search
mnemosyne search "prefers JSON" --agent agent-1

# Watch for changes in real-time
mnemosyne watch --agent agent-1 --updates-only
```

## Installation

```bash
go install github.com/qjcg/arcadia/exp/mnemosyne@latest
```

## Requirements

- Go 1.26+
- NATS server 2.11+ with JetStream enabled

## Documentation

- [Technical Design Document](docs/TDD.md)
- [Product Requirements](docs/PRD.md)

## Related Tools

- [horeb](../../cmd/horeb) — Random unicode generator
- [sv](../../cmd/sv) — Semantic versioning for monorepos
