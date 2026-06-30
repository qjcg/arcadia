# Mnemosyne — Agentic Memory over NATS KV

**Technical Design Document**

---

## 1. Overview

Mnemosyne (pronounced *nee-MOZ-uh-nee*) is a CLI tool that gives AI agents
structured, persistent, distributed memory using [NATS KV][nats-kv] as the
backend.  Named after the Greek Titaness of memory, it provides an
agent-friendly abstraction over NATS's key-value store — supporting episodic,
semantic, procedural, and working memory types with search, watch,
importance-based pruning, and multi-tenant isolation.

[nats-kv]: https://docs.nats.io/nats-concepts/jetstream/key-value-store

---

## 2. Architecture

```
┌─────────────────────────────────────────────┐
│              AI Agent Process                │
│                                              │
│  mnemosyne CLI  ◄─────────  subprocess/exec  │
│       │                                      │
│       ▼                                      │
│  Internal Go library (agentic/memory)         │
│       │                                      │
│       ▼                                      │
│  NATS Go client (nats.go + jetstream)         │
├─────────────────────────────────────────────┤
│              NATS Server                      │
│  ┌─────────────────────────────────────────┐ │
│  │  JetStream Domain                       │ │
│  │  ┌────────────────────────────────────┐ │ │
│  │  │  KV_<namespace> Bucket             │ │ │
│  │  │  ├── agent_1.episodic.<id>         │ │ │
│  │  │  ├── agent_1.semantic.<id>         │ │ │
│  │  │  ├── agent_2.episodic.<id>         │ │ │
│  │  │  └── ...                           │ │ │
│  │  └────────────────────────────────────┘ │ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

### Layers

| Layer | Responsibility |
|-------|---------------|
| **CLI** (`exp/mnemosyne/`) | Cobra commands, flags, help text, human/JSON output |
| **Library** (`internal/memory/`) | Core memory model, CRUD, search, serialization |
| **NATS KV** (`internal/natskv/`) | Thin adapter wrapping `jetstream.KeyValue`; handles bucket lifecycle, reconnection |
| **NATS Server** | JetStream-backed KV buckets; persistence, replication, TTL |

---

## 3. Project Structure

```
exp/mnemosyne/
├── main.go                  # Entry point; calls cli.Execute()
├── main_test.go             # testscript integration tests
├── Taskfile.yml             # build, test, lint tasks
├── go.mod / go.sum
├── docs/
│   ├── TDD.md               # this document
│   ├── PITCH.md
│   ├── TAGLINE.md
│   └── PRD.md
├── internal/
│   ├── cli/
│   │   ├── root.go          # Root command, persistent flags, Execute()
│   │   ├── put.go           # mnemosyne put
│   │   ├── get.go           # mnemosyne get
│   │   ├── search.go        # mnemosyne search
│   │   ├── list.go          # mnemosyne list
│   │   ├── forget.go        # mnemosyne forget
│   │   ├── prune.go         # mnemosyne prune
│   │   ├── watch.go         # mnemosyne watch
│   │   ├── status.go        # mnemosyne status
│   │   ├── config.go        # mnemosyne config (init, show, set)
│   │   └── export_import.go # mnemosyne export / import
│   ├── memory/
│   │   ├── entry.go         # Memory type, Entry struct, validation
│   │   ├── store.go         # MemoryStore interface
│   │   ├── filter.go        # Search/filter predicates
│   │   └── consolidate.go   # Importance scoring, pruning logic
│   └── natskv/
│       ├── bucket.go        # Bucket init, lookup, config
│       └── adapter.go       # NATS KV → MemoryStore adapter
└── testdata/
    └── *.txtar              # testscript integration tests
```

---

## 4. Data Model

### 4.1 Key Hierarchy

```
{namespace}.{agent_id}.{memory_type}.{memory_id}
```

NATS KV keys use dot-separated segments (same as NATS subjects), enabling
wildcard listing and filtering.

| Segment | Example | Description |
|---------|---------|-------------|
| `namespace` | `arcadia` | Multi-tenant isolation; maps to one KV bucket `KV_arcadia` |
| `agent_id` | `agent-a7x` | Unique agent identifier |
| `memory_type` | `episodic` | One of: `episodic`, `semantic`, `procedural`, `working`, `preference` |
| `memory_id` | `mem_01j2...` | Unique ID (ULID or UUID), generated client-side |

### 4.2 Memory Types

| Type | Purpose | TTL | Example |
|------|---------|-----|---------|
| `episodic` | Past conversations, observations | 7 days | "User said they prefer concise answers" |
| `semantic` | Facts, knowledge, concepts | 90 days | "The app supports both SQLite and PostgreSQL" |
| `procedural` | How to perform tasks | 180 days | "To deploy, first run tests, then build..." |
| `working` | Current context, scratch space | 1 hour | "Currently processing user's login request" |
| `preference` | User/agent preferences | Permanent | "User prefers dark mode in responses" |

### 4.3 Entry Value (JSON)

```json
{
  "id": "mem_01J2XYZ...",
  "agent_id": "agent-a7x",
  "type": "episodic",
  "content": "The user prefers answers under 4 lines of prose.",
  "metadata": {
    "created_at": "2026-06-30T12:00:00Z",
    "last_accessed": "2026-06-30T14:30:00Z",
    "access_count": 3,
    "importance": 0.85,
    "ttl": "168h",
    "tags": ["preference", "style", "user-123"],
    "source": "session_s7g2k9",
    "parent_id": "",
    "embedding": null
  }
}
```

**Go struct:**

```go
type Entry struct {
    ID        string   `json:"id"`
    AgentID   string   `json:"agent_id"`
    Type      MemType  `json:"type"`
    Content   string   `json:"content"`
    Metadata  Metadata `json:"metadata"`
}

type Metadata struct {
    CreatedAt    time.Time         `json:"created_at"`
    LastAccessed time.Time         `json:"last_accessed"`
    AccessCount  int64             `json:"access_count"`
    Importance   float64           `json:"importance"`    // 0.0 - 1.0
    TTL          time.Duration     `json:"ttl,omitempty"`
    Tags         []string           `json:"tags,omitempty"`
    Source       string            `json:"source,omitempty"`
    ParentID     string            `json:"parent_id,omitempty"`
    Embedding    []float64         `json:"embedding,omitempty"`
}

type MemType string

const (
    Episodic   MemType = "episodic"
    Semantic   MemType = "semantic"
    Procedural MemType = "procedural"
    Working    MemType = "working"
    Preference MemType = "preference"
)
```

### 4.4 NATS KV Bucket Configuration

```go
bucketConfig := jetstream.KeyValueConfig{
    Bucket:       "KV_" + namespace,     // e.g. "KV_arcadia"
    Description:  "Agentic memory store for " + namespace,
    MaxValueSize: 1 << 20,                // 1 MB per value
    History:      5,                      // keep 5 revisions for CAS
    TTL:          30 * 24 * time.Hour,    // bucket-level max TTL
    MaxBytes:     0,                      // unlimited bucket size
    Storage:      jetstream.FileStorage,  // persist to disk
    Replicas:     1,                      // adjustable
}
```

---

## 5. CLI Interface

### 5.1 Root Command

```
mnemosyne — agentic memory for AI agents, backed by NATS KV

Usage:
  mnemosyne [flags] [command]

Persistent Flags:
  -s, --server string      NATS server URL (default "nats://localhost:4222")
  -n, --namespace string   Memory namespace / tenant (default "default")
  -c, --creds string       NATS credentials file (.creds or .jwt)
      --json               Output in JSON format
  -q, --quiet              Suppress non-error output
  -V, --verbose            Enable verbose logging

Available Commands:
  init      Initialize a new memory namespace
  put       Store a new memory entry
  get       Retrieve a specific memory entry
  list      List memory entries with optional filters
  search    Search memories by content, tags, or importance
  forget    Delete a memory entry
  prune     Remove expired or low-importance entries
  watch     Watch for real-time memory changes
  config    Show or update bucket configuration
  status    Show bucket stats and health
  export    Export memories to JSON/JSONL
  import    Import memories from JSON/JSONL
  help      Help about any command

Use "mnemosyne <command> --help" for more information about a command.
```

### 5.2 Subcommand Details

#### `mnemosyne init`

Initialize a new namespace / KV bucket.

```
mnemosyne init [namespace] [flags]

  --description string    Bucket description
  --history uint          Max revisions per key (default 5)
  --ttl duration          Default entry TTL (default 720h)
  --max-value-size int    Max value size in bytes (default 1MB)
  --replicas int          Number of replicas (default 1)
  --storage string        Storage type: file, memory (default "file")
```

#### `mnemosyne put`

Store a new memory entry.

```
mnemosyne put [flags]

  -a, --agent string       Agent ID (required)
  -t, --type string        Memory type: episodic, semantic, procedural, working, preference (default "episodic")
  -c, --content string     Memory content (required, or use --content-file)
      --content-file string  Read content from file
      --importance float    Importance score 0.0-1.0 (default 0.5)
      --ttl duration        Entry TTL (overrides bucket default)
      --tags strings        Comma-separated tags
      --source string       Source identifier (session ID, etc.)
      --parent-id string    Parent memory entry ID
```

#### `mnemosyne get`

Retrieve a specific entry by ID.

```
mnemosyne get <memory-id> [flags]

  -a, --agent string    Agent ID filter (required if memory-id is not globally unique)
      --revision uint   Specific revision number
```

#### `mnemosyne list`

List entries with optional filters. Defaults to listing all entries in the
namespace.

```
mnemosyne list [flags]

  -a, --agent string     Agent ID filter
  -t, --type string      Memory type filter
      --tags strings     Filter by tags (AND)
      --min-importance float  Minimum importance (0.0-1.0)
      --max-age duration Max age of entries (e.g. 24h)
      --limit int        Max entries to return (default 100)
      --offset int       Skip N entries
      --sort string      Sort field: created, access_count, importance (default "created")
      --order string     Sort order: asc, desc (default "desc")
```

#### `mnemosyne search`

Semantic / text-based search across entries.

```
mnemosyne search <query> [flags]

  -a, --agent string     Agent ID filter
  -t, --type string      Memory type filter
      --tags strings     Filter by tags (AND)
      --min-importance float  Minimum importance
      --limit int        Max results (default 20)
      --fuzzy            Enable fuzzy text matching
```

#### `mnemosyne forget`

Delete a memory entry.

```
mnemosyne forget <memory-id> [flags]

  -a, --agent string   Agent ID (required if not globally unique)
      --purge          Permanently remove all revisions
```

#### `mnemosyne prune`

Remove expired entries and entries below an importance threshold.

```
mnemosyne prune [flags]

  -a, --agent string         Agent ID filter
  -t, --type string          Memory type filter
      --min-importance float Importance threshold (entries below this are pruned) (default 0.1)
      --expired-only         Only remove TTL-expired entries
      --dry-run              Show what would be pruned without removing
```

#### `mnemosyne watch`

Watch for real-time changes to memories.

```
mnemosyne watch [flags]

  -a, --agent string     Agent ID filter
  -t, --type string      Memory type filter
      --updates-only     Only show new changes (skip current state)
      --json             Output each event as one JSON line
```

#### `mnemosyne status`

Show bucket stats and configuration.

```
mnemosyne status [namespace] [flags]

      --verbose   Show detailed stream info
```

Example output:

```
Namespace:   arcadia
Bucket:      KV_arcadia
Entries:     1,234 (active)
Revisions:   5,102 (total)
Size:        48.2 MB
Storage:     File
Replicas:    3
TTL:         720h
Created:     2026-03-15
```

#### `mnemosyne config`

Show or update namespace configuration.

```
mnemosyne config [flags]

  --ttl duration          Set default TTL
  --history uint          Set max history per key
  --max-value-size int    Set max value size (bytes)
  --description string    Update bucket description
```

#### `mnemosyne export` / `mnemosyne import`

Bulk data transfer.

```
mnemosyne export [flags]

  -a, --agent string     Agent ID filter
  -t, --type string      Memory type filter
  -o, --output string    Output file ("-" for stdout)
  -f, --format string    Output format: json, jsonl (default "jsonl")

mnemosyne import <file> [flags]

  -n, --namespace string   Target namespace (default "default")
      --merge             Merge with existing (update by ID)
      --dry-run           Validate without importing
```

---

## 6. MemoryStore Interface

The core abstraction in `internal/memory/store.go`:

```go
type MemoryStore interface {
    // Lifecycle
    Init(ctx context.Context) error
    Close(ctx context.Context) error

    // CRUD
    Put(ctx context.Context, entry Entry) (revision uint64, err error)
    Get(ctx context.Context, agentID, memoryID string) (Entry, error)
    List(ctx context.Context, filter Filter) ([]Entry, error)
    Delete(ctx context.Context, agentID, memoryID string) error

    // Search
    Search(ctx context.Context, query string, filter Filter) ([]Entry, error)

    // Pruning / consolidation
    Prune(ctx context.Context, filter Filter) (pruned int, err error)

    // Watch
    Watch(ctx context.Context, filter Filter) (<-chan EntryEvent, error)

    // Status
    Status(ctx context.Context) (BucketStatus, error)
    Config(ctx context.Context) (BucketConfig, error)
    UpdateConfig(ctx context.Context, cfg BucketConfigUpdate) error
}

type Filter struct {
    AgentID       string
    Type          MemType
    Tags          []string
    MinImportance float64
    MaxAge        time.Duration
    Limit         int
    Offset        int
    Sort          string  // "created" | "access_count" | "importance"
    Order         string  // "asc" | "desc"
}

type EntryEvent struct {
    Entry Entry
    Op    KeyValueOp  // Put, Delete, Purge
}

type BucketStatus struct {
    Namespace   string
    Bucket      string
    ActiveKeys  uint64
    TotalRevisions uint64
    Size        uint64
    Storage     string
    Replicas    int
    TTL         time.Duration
    Created     time.Time
}
```

### NATS KV Adapter

The `internal/natskv/adapter.go` file implements `MemoryStore` by mapping
operations to NATS KV calls:

| MemoryStore method | NATS KV mapping |
|---|---|
| `Put` | `kv.Put(ctx, key, jsonBytes)` |
| `Get` | `kv.Get(ctx, key)` + JSON unmarshal |
| `List` | `kv.ListKeys(ctx)` + `kv.Get` per key, filtered client-side |
| `Delete` | `kv.Delete(ctx, key)` |
| `Watch` | `kv.Watch(ctx, pattern)` → decode events |
| `Search` | `kv.ListKeys` + client-side content filtering |
| `Prune` | `kv.ListKeys` + delete matching each expired entry |
| `Status` | `kv.Status(ctx)` → map to BucketStatus |
| `Config` | `kv.Status(ctx).Config()` |
| `UpdateConfig` | Delete + re-create bucket with new config |

---

## 7. Consistency & Concurrency

### Revision-Based Optimistic Locking

NATS KV provides per-key revision numbers. The adapter uses these for
optimistic concurrency:

```go
func (s *KVStore) Update(ctx context.Context, entry Entry, expectedRevision uint64) error {
    key := s.keyFor(entry.AgentID, entry.Type, entry.ID)
    data, _ := json.Marshal(entry)
    _, err := s.kv.Update(ctx, key, data, expectedRevision)
    return err  // fails if revision changed
}
```

Clients (agents) should read-then-update with the known revision.

### Watch-Based Invalidation

Agents watching the same keyspace receive real-time updates. The watcher
channel delivers `KeyValueEntry` objects with `Operation()` indicating
`Put`, `Delete`, or `Purge`.

### Conflict Resolution

- **Last-write-wins** by default (via `kv.Put`)
- **Optimistic locking** via `kv.Update` + revision check for cooperative agents
- **History** (configurable, default 5) allows rolling back to prior revisions

---

## 8. TTL, Pruning, and Consolidation

### TTL Strategy

- **Per-entry TTL** overrides bucket-level TTL (implemented via NATS KV's
  per-key TTL support in `jetstream.KeyValueConfig`)
- **Bucket-level TTL** acts as a ceiling; entries cannot outlive it
- Default TTLs vary by memory type (see §4.2)

### Importance Scoring

After each `Get`, the adapter increments `access_count` and recalculates
importance using an exponential decay:

```
importance = base_importance × decay^(days_since_last_access)
```

### Pruning

`prune` removes entries that:
1. Have exceeded their TTL (handled automatically by NATS KV)
2. Fall below a configurable `--min-importance` threshold
3. Have delete markers older than 30 minutes (via `PurgeDeletes`)

---

## 9. Security

### Authentication

- **NATS Credentials**: `.creds` / `.jwt` files via `--creds` flag
- **NKEY-based auth**: NKEY seed via `--nkey` flag
- **Token auth**: JWT token via `--token` flag

### Authorization

- **Bucket-level isolation**: Each namespace is a separate KV bucket, each
  with its own NATS permissions
- **Agent-level isolation**: Enforced at the application layer via agent_id
  prefix filtering; future: per-bucket ACL

### Encryption in Transit

- NATS TLS via `--tls` / `--tlscert` / `--tlskey` flags
- Standard `nats.Option` TLS configuration

---

## 10. Output Formats

### Human-readable (default)

Table output for lists, summary blocks for status, single-value for gets.

### JSON (`--json`)

Machine-readable JSON output for all commands.

**List output (JSON mode):**
```json
{
  "entries": [...],
  "total": 42,
  "filter": { "agent": "agent-a7x", "type": "episodic" }
}
```

**Watch output (JSON mode, one event per line):**
```json
{"op":"put","entry":{...}}
{"op":"delete","entry":{"id":"mem_..."}}
```

### Quiet mode (`--quiet`)

Suppresses all non-error output; exit code indicates success/failure.
Suitable for scripting.

---

## 11. Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (config, connection, etc.) |
| 2 | Entry not found |
| 3 | Invalid input |
| 4 | NATS connection error |

---

## 12. Implementation Phases

### Phase 1 — Core (Week 1)
- [ ] Project scaffold: `main.go`, `go.mod`, `Taskfile.yml`
- [ ] `internal/natskv/adapter.go`: NATS connection, bucket init
- [ ] `internal/memory/entry.go`: Entry struct, validation
- [ ] `internal/memory/store.go`: MemoryStore interface
- [ ] CLI root with persistent flags
- [ ] `put`, `get`, `list`, `delete` commands

### Phase 2 — Search & Watch (Week 2)
- [ ] `internal/memory/filter.go`: Filter predicates
- [ ] `internal/memory/search.go`: Content search (fuzzy, tag, importance)
- [ ] `search`, `watch` commands
- [ ] JSON output support

### Phase 3 — Maintenance (Week 3)
- [ ] `prune`, `config`, `status` commands
- [ ] Importance scoring and decay
- [ ] TTL management
- [ ] `export`, `import` commands

### Phase 4 — Polish (Week 4)
- [ ] Full testscript integration tests
- [ ] Error messages per clig.dev guidelines
- [ ] Help text and examples
- [ ] `init` command with interactive setup
- [ ] Documentation: PITCH.md, PRD.md, README.md

---

## 13. Technology Choices

| Choice | Rationale |
|--------|-----------|
| **Go 1.26** | Matches workspace, modern stdlib |
| **spf13/cobra** | Standard in project for CLIs |
| **spf13/viper** | Config precedence (flags > env > file) |
| **nats-io/nats.go** | Official Go client; already used in examples |
| **jetstream package** | Newer, context-aware API; preferred over legacy |
| **testscript** | Standard in project for CLI integration tests |
| **Taskfile.yml** | Standard in project for build tasks |
| **ulid** | Sortable, unique IDs for memory entries |

---

## 14. Non-Goals (Explicit Out of Scope)

- **Vector database**: Embedding search requires a vector DB adapter;
  embeddings can be stored in metadata but search is not implemented here
- **Agent runtime**: Mnemosyne provides memory storage, not agent execution
- **Multi-bucket sharding**: One bucket per namespace; no auto-sharding yet
- **Graph relationships**: No explicit graph traversal between memories
  (parent_id is a flat reference)
- **Web UI**: CLI-only; no dashboard or graphical interface
