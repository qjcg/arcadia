---
name: pyramid
description: Create reversible multi-level summaries at 5 zoom levels for AI agent responses and implementation plans. Use when users need to (1) scale response detail up or down, (2) get a quick decision/outcome without full reasoning, (3) review approaches before diving into execution details, (4) compress long outputs for high-level review, or (5) expand brief answers into comprehensive specifications.
---

# Pyramid Summaries

Compress or expand AI agent responses across 5 detail levels. Start at the level appropriate for the user's context budget, then zoom in or out as needed.

## The 5 Levels

| Level | Scope | Purpose |
|-------|-------|---------|
| **0** | 1 sentence (~10-20 words) | The decision/outcome only—no reasoning |
| **1** | 1 paragraph (~40-60 words) | What + Why—essential context and rationale |
| **2** | Structured outline (~150-250 words) | What + Why + How—key steps and considerations |
| **3** | Full response (~500-800 words) | Complete reasoning—alternatives, trade-offs, dependencies |
| **4** | Exhaustive detail | Full specification—production-ready depth |

## Navigation Patterns

### Quick Decisions
Use Level 0-1 when the user trusts the agent and just needs the answer.

**Example progression:**
- Level 0: "Refactor the auth module into a separate service."
- Level 1: "Refactor auth into a microservice to resolve circular dependencies blocking the payments team. This unblocks 3 pending PRs and aligns with Q3 architecture goals."

### Review Mode
Use Level 2 to understand the approach without drowning in detail.

**Example:**
```
Approach: Extract auth logic into standalone service

Steps:
1. Create auth-service repo with FastAPI scaffold
2. Migrate /login and /token endpoints (preserve API contracts)
3. Update gateway to proxy auth routes
4. Add service-to-service authentication
5. Deploy to staging with feature flag

Risks:
- Session state migration (mitigation: dual-write period)
- Latency increase (mitigation: connection pooling)

Success criteria:
- Zero-downtime migration
- <50ms latency overhead
```

### Approval Mode
Use Level 3 when the user needs to evaluate trade-offs and alternatives.

Include:
- Rejected approaches and why
- Detailed step sequence with estimates
- Explicit dependencies (what must happen first)
- Risk mitigation strategies
- Rollback plan

### Execution Mode
Use Level 4 when all details are needed to implement.

Include:
- Exact file changes (create/modify/delete)
- Code samples for critical sections
- Test cases with expected inputs/outputs
- Error handling for each edge case
- Rollout plan with monitoring checks

## Applying to Any Response Type

### Code Review
- Level 0: "Approve with 2 minor changes."
- Level 1: "Approve—fix null check and add docstring."
- Level 2: Outline of issues by category (security, style, logic)
- Level 3: Full review with suggestions and rationale
- Level 4: Line-by-line commentary with proposed diffs

### Bug Analysis
- Level 0: "Race condition in cache invalidation—fix with mutex."
- Level 1: "Cache reads stale data during concurrent updates. Add locking around read-modify-write cycle."
- Level 2: Reproduction steps, root cause, fix approach, verification plan
- Level 3: Include similar bugs found, prevention strategies, monitoring
- Level 4: Full RCA with logs, traces, proposed code change, test cases

### Research Summary
- Level 0: "Use PostgreSQL over MySQL for JSON operations."
- Level 1: "PostgreSQL's JSONB indexing outperforms MySQL for our query patterns and supports partial updates."
- Level 2: Comparison matrix, benchmark results, migration effort estimate
- Level 3: Deep dive on JSONB internals, benchmark methodology, risk assessment
- Level 4: Full evaluation report with citations, POC code, detailed migration plan

## Zoom Commands

When the user wants to change levels:

| Command | Action |
|---------|--------|
| "Summarize" / "TL;DR" | Compress to Level 0-1 |
| "Expand" / "More detail" | Increase by 1-2 levels |
| "Full detail" / "Comprehensive" | Jump to Level 4 |
| "Outline only" | Target Level 2 |

## Level 0 Writing Tips

- Lead with the verb: "Deploy", "Reject", "Refactor", "Investigate"
- Include the subject and key constraint
- Omit articles and filler words when needed

**Weak:** "I think we should probably consider refactoring..."
**Strong:** "Refactor auth service to eliminate circular dependency."

## When to Stop Expanding

- The user says "that's enough"
- Additional detail doesn't change the decision
- Further expansion would duplicate source material

## Combining with Other Patterns

**MapReduce + Pyramid:**
1. Map: Generate Level 0 summaries for 50 items in parallel
2. Cluster: Group by similarity using Level 0
3. Reduce: Expand interesting clusters to Level 2-3

This lets a context-limited model "see" 50 items at Level 0 (~1000 words) instead of 50 full responses (~25000 words).
