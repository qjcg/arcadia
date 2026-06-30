# Mnemosyne Press Release & FAQ

---

## Press Release

**FOR IMMEDIATE RELEASE**

### Mnemosyne Gives AI Agents Distributed, Persistent Memory Over NATS

**ARCADIA PROJECT — June 30, 2026** — The Arcadia Project today announced
Mnemosyne, a new command-line tool that provides structured, typed memory
for AI agents using NATS KV as a distributed backend.

AI agents today are stateless by design — each interaction starts from
scratch. Developers work around this with ad-hoc databases, Redis caches,
or flat files, none of which are designed for the way agents actually
use memory: episodic (what happened), semantic (what I know), procedural
(how to do it), and working (what I'm doing now).

Mnemosyne gives agents a purpose-built memory system. A single CLI command
stores a memory with type, tags, importance, and TTL. Another retrieves it.
Agents can search across thousands of memories, watch for real-time changes
from other agents, and automatically prune what's no longer relevant.

"Agents should remember what matters and forget what doesn't," said the
Arcadia team. "NATS KV gives us a rock-solid distributed foundation with
replication, TTL, and history built in. Mnemosyne is the agent-friendly
layer on top."

Mnemosyne runs against any NATS server with JetStream enabled, including
embedded single-node setups and production clusters. A Go library for
in-process agent use and a standalone CLI for any language or workflow
are both available.

"Before Mnemosyne, we were juggling three different storage backends for
our agent fleet," said a beta user. "Now it's one command, one data model,
one place to look. Our agents actually remember what happened last session."

Mnemosyne is open source under the MIT License. Get started with
`mnemosyne init` and your first `mnemosyne put`.

---

## Frequently Asked Questions

### External FAQs

**Q: Do I need a running NATS server?**
A: Yes. Mnemosyne requires a NATS server with JetStream enabled. You can
run one locally (`nats-server -js`), use a cloud-hosted service like Synadia
NATS Cloud, or embed one in-process during development (see the NATS Go
examples in this repo).

**Q: How is this different from just using NATS KV directly?**
A: NATS KV is a generic binary key-value store. Mnemosyne adds an agentic
data model (memory types, importance, tags, TTL by type), structured JSON
values, search, watch, and pruning — all through a purpose-built CLI.

**Q: Can I use this from languages other than Go?**
A: Yes. The CLI can be called from any language as a subprocess. Pipe JSON
in for `put` and read JSON out for `list` / `search`. A Go library is also
available for in-process use.

**Q: What happens when NATS is down?**
A: Mnemosyne returns a clear error and non-zero exit code. Agents should
handle this gracefully (retry, degrade, or fail). On reconnect, NATS KV
re-syncs automatically.

**Q: How are memories encrypted?**
A: In transit via NATS TLS. At rest, NATS FileStorage encryption is
available. Mnemosyne passes through whatever NATS security configuration
you provide.

### Internal FAQs

**Q: Why NATS KV instead of Redis, SQLite, or a vector database?**
A: NATS KV was chosen for its unique combination of distributed replication,
built-in TTL, history support, real-time watch, and multi-tenant isolation
via buckets — all without needing a separate database server. It's already
used elsewhere in the Arcadia project. A vector DB adapter could be added
later for embedding-based search.

**Q: How do we prevent memory leaks / unbounded growth?**
A: Two mechanisms: TTL (per-type defaults + per-entry overrides) and
importance-based pruning. Both are automated. NATS KV also supports
`MaxBytes` bucket limits for a hard cap.

**Q: What's the scale target?**
A: Thousands of agents, millions of entries per namespace. NATS KV handles
this with JetStream clustering. Mnemosyne's client-side filtering becomes
the bottleneck first; future versions may push filtering into the KV layer.

**Q: Is this meant for production use?**
A: Yes. NATS KV is production-grade. Mnemosyne adds minimal overhead — it's
essentially a thin client with a data model on top. The `mnemosyne status`
command reports bucket health.
