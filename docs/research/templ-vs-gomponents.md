# templ vs gomponents — research notes

**Question:** How do Go's [templ](https://templ.guide) and
[gomponents](https://github.com/maragudk/gomponents) compare as
HTML-component / templating solutions for Go backends?

**Method:** Claims are traced to primary sources only. The two projects
publish canonical machine-readable doc bundles, which were used as the
backbone: [templ.guide/llms.md](https://templ.guide/llms.md) (the templ
project's own consolidated docs for coding assistants) and
[gomponents.com/llms.txt](https://www.gomponents.com/llms.txt) (same for
gomponents). These were cross-checked against the templ docs site and
source repo, the gomponents repo / README / source, and the GitHub API
for ecosystem stats. No secondary comparison posts were used.
Self-reported numbers are labeled as such.

**Location convention:** This repo had no prior research-notes
convention (only `docs/adrs/` and `docs/agents/`), so this note creates
`docs/research/`, per AGENTS.md's "store documentation in a `docs/`"
rule.

---

## Summary

Both libraries solve the same problem — server-rendered HTML components
in Go — with opposite architectures:

- **templ** is a *DSL + compiler*. You write components in `.templ`
  files; `templ generate` compiles them into Go functions returning a
  `templ.Component` interface. You get a custom toolchain (CLI, LSP,
  formatter, live-reload proxy) and purpose-built injection defenses,
  at the cost of a build step, generated code in your tree, and a non-Go
  language to learn.
- **gomponents** is a *plain Go library*. A component is a Go function
  returning a `Node` interface that renders itself to an `io.Writer`.
  There is no code generation and no external dependencies; every Go
  tool works out of the box. The cost is more verbose nesting,
  type-only error surfacing (the Go compiler catches type errors, but
  nothing validates the HTML), and DIY solutions for things templ
  bundles (scripts, CSS scoping, hot reload).

The gomponents author's own position is that `html/template`, templ,
and gomponents "are all good choices, and it largely comes down to
preference" (README FAQ, "Should I choose html/template, Templ, or
gomponents?" — see [README][gomp-readme]).

---

## Comparison

### 1. Authoring model

| Aspect       | templ                             | gomponents           |
| ------------ | --------------------------------- | -------------------- |
| Component    | `templ` block → `templ.Component` | Go func → `Node`     |
| Build step   | `templ generate` required         | none — pure Go       |
| File types   | `*.templ` + `*_templ.go`          | ordinary `*.go`      |
| Dependencies | templ runtime module              | none (zero requires) |

Sources: [templ.guide: Components][templ-components],
[gomponents package doc][gomp-pkg], [README, Features][gomp-readme].

Details and quotes:

- templ component signature and generation: `templ generate` turns a
  `templ` block into `func(...) templ.Component`; `templ.Component` is
  `Render(ctx context.Context, w io.Writer) error`
  ([Components][templ-components]).
- gomponents: "pure Go, with no extra build step like Templ, so it
  works with all tools that already support Go" (README FAQ; see
  [README][gomp-readme]).
- Dependencies (verified in repo source): gomponents' `go.mod` is just
  `module maragu.dev/gomponents` / `go 1.18`, zero requires.

templ example (from [templ.guide][templ-guide]):

```go
templ Hello(name string) {
  <div>Hello, { name }</div>
}
```

gomponents example (from [README, Usage][gomp-readme]):

```go
func NavbarLink(href, name, currentPath string) Node {
	return A(Href(href), Classes{"is-active": currentPath == href}, Text(name))
}
```

gomponents' element/attribute creators mirror HTML names directly
(`A`, `Hr`, `Div`…), but name clashes require `El`/`Attr` suffixes for
`cite`, `data`, `form`, `label`, `style`, `title` (README FAQ; see
[README][gomp-readme]). gomponents ships a complete-document component:
`HTML5(HTML5Props{Title, Language, Head, Body})` plus `Doctype()`
([llms.txt][gomp-llms], `components` package). Because gomponents is
plain Go, control flow is ordinary `if`/`for` plus helpers `If`/`Iff`,
`Map`, `Group` ([README, Features][gomp-readme]); templ also allows Go
`if`/`switch`/`for` inside its blocks ([templ.guide][templ-guide]). The
recommended gomponents style is dot-imports so calls read as an HTML
DSL (or single-letter aliases `g`/`h`/`c` if you dislike dot-imports)
([llms.txt][gomp-llms]).

### 2. Type safety and error surfacing

- **Both** give compile-time type safety on component inputs:
  components are Go functions with typed parameters in both models
  ([Components][templ-components], [README, Features][gomp-readme] —
  "Type safety from the compiler").
- **templ** additionally parses `.templ` files, so *syntax* errors in
  the template layer surface at `templ generate` time (with
  source-mapped errors) or live in the editor via the LSP
  ([CLI][templ-cli], [IDE support][templ-ide]). Unknown HTML structure
  is still not validated — templ checks tag closure but not document
  validity ([Basic syntax][templ-basic] — "templ requires that all HTML
  elements are closed").
- **gomponents** has no template layer to get wrong: any type error is
  an ordinary Go compile error. But nothing checks the resulting
  markup: "gomponents does not check that the result is valid HTML, by
  design" ([skills][gomp-skills], `skills/gomponents/SKILL.md`).

### 3. Composition and data passing

- templ: compose via `@component()` calls and `children...` slots
  ([Template composition][templ-composition]); components can be passed
  as parameters typed `templ.Component` (same page); method components
  let you attach rendering to types ([Components][templ-components]).
- templ explicitly addresses *prop drilling*: data can be set on
  `context.Context` and read in descendants ([Context][templ-context])
  — "The technique of passing data through a stack" is called out as a
  problem there.
- gomponents composes by plain function application: everything is a
  `Node`, elements take child `Node`s variadically
  ([skills][gomp-skills]). Data passing is ordinary Go parameters; the
  library exists largely because the author found data-passing awkward
  in `html/template` (README FAQ; see [README][gomp-readme]).

### 4. Rendering, streaming, HTTP integration

- templ's `templ.Component` renders to an `io.Writer` with a
  `context.Context`: `Render(ctx, w) error`. The docs note a component
  "may write partial output to the `io.Writer` if it returns an error"
  — i.e., rendering is streaming by nature, with buffering left to the
  caller ([Components][templ-components]). A built-in `templ.Handler`
  adapts components to `net/http`, with functional options
  `templ.WithStatus`, `templ.WithContentType`, and
  `templ.WithErrorHandler` ([llms.md][templ-llms], "Creating an HTTP
  server with templ"). templ's deployment story is "deploy like any
  other Go application" — Docker, serverless (AWS Lambda via
  `algnhsa`), or a plain binary ([llms.md][templ-llms], "Hosting").
- templ's own guidance is that good components are "idempotent, pure
  functions" — same parameters always yield the same HTML, no network
  or disk access inside components ([llms.md][templ-llms], "Creating an
  HTTP server with templ").
- gomponents' `Node.Render(w io.Writer) error` also renders straight to
  the writer ("gomponents renders directly to an `io.Writer`, making it
  efficient for server-side rendering" — [README, FAQ][gomp-readme],
  "Is gomponents fast?"). The [`http` package][gomp-http] provides
  `Adapt`, turning `func(w, r) (Node, error)` into an
  `http.HandlerFunc`, with status-code-aware error handling.

Both therefore stream naturally and integrate with `net/http` with one
small adapter.

### 5. Escaping and XSS safety

**templ** has purpose-built, layered defenses ([Injection
attacks][templ-injection]):

- text expressions are escaped with `templ.EscapeString`;
- no Go expressions allowed inside `<script>`/`<style>` tags at all;
- `on*` attributes only accept `templ.ComponentScript`; script-template
  parameters are JSON-escaped per Google's SafeHTML
  ([discussion #206][templ-disc-206] — first-party maintainer
  statement);
- `style` attributes cannot be expressions (constants only; use `css`
  templates);
- class names are sanitized by default (`templ.SafeClass` to bypass);
- `href` values are sanitized to strip `javascript:` URLs unless
  wrapped in `templ.SafeURL` (see also [issue #908][templ-issue-908]
  reporting docs drift on this exact point);
- CSS expression values pass through `templ.SanitizeCSS`.
- Known limitation: escaping is simple (`templ.Escape`), not fully
  context-aware ([issue #6][templ-issue-6], first-party issue tracker).

**gomponents** escaping rules, from the [package documentation][gomp-pkg]
and source (`gomponents.go`):

- text via `Text`/`Textf` and attribute *values* are HTML-escaped (the
  escape set matches `text/template.HTMLEscapeString`'s six
  replacements — see `htmlEscapeSet` in `gomponents.go`);
- element and attribute *names* are rendered unescaped and "must be a
  trusted value, never user-controlled data" (`El`/`Attr` docs);
- `Raw`/`Rawf` insert unescaped strings deliberately.

Net: templ's protection is broader (scripts, styles, URLs, classes get
dedicated rules); gomponents' is simpler but requires the same
discipline about `Raw` and about never passing user data as
tag/attribute names. templ also supports CSP nonces for inline scripts
([Content security policy][templ-csp]).

### 6. Tooling and developer experience

- **templ** ships an official toolchain ([CLI][templ-cli]):
  `templ generate`, `templ fmt`, `templ lsp`, plus built-in live reload
  via `templ generate --watch --proxy=... --cmd="go run ."` ([Live
  reload][templ-reload]) that regenerates on `.templ` changes, restarts
  the server on `.go` changes, and reloads the browser — via a proxy on
  port 7331. IDE support: official VS Code extension and JetBrains
  plugin, LSP integrates with gopls for Go code inside templates
  ([IDE support][templ-ide]).
- Caveat: code from `templ generate --watch` "is not optimised for
  production use" ([Components][templ-components], "Template
  generation"); production builds should run plain `templ generate`.
- **gomponents** needs no tooling: "Automatic formatting with
  `gofmt`/`goimports`", "Easy debugging with the standard Go debugger"
  ([README, Features][gomp-readme]). Hot reload is whatever you already
  use for Go (e.g. `air` — which AGENTS.md in this repo mandates
  anyway).

Extra toolchain facts from [llms.md][templ-llms]:

- `templ fmt -fail` exits non-zero if any file needed reformatting — a
  ready-made CI formatting gate. `script`/`style` element formatting is
  handed to `prettier`/`prettierd`/`npx` when present.
- Recommended workflow is to **commit generated `*_templ.go` files** so
  the repo builds without a generate step, and guard freshness in CI
  with `templ generate && git diff --exit-code`.
- `templ lsp` embeds its own gopls instance (with `-gopls-remote`
  shared-daemon mode available).
- Live reload internals: the proxy injects reload JS before `</body>`
  and listens on SSE (`/_templ/reload/events`); in watch mode, text-only
  `.templ` changes are hot-swapped via `TEMPL_DEV_MODE` text files
  *without recompiling Go*, so only Go-code changes restart the server.
  External build tools (`air`, `wgo`) can trigger reloads via
  `templ generate --notify-proxy`.

### 7. Performance

Evidence here is thin and asymmetric — treat it accordingly:

- templ's only first-party performance claim found is qualitative:
  components are "compiled into performant Go code"
  ([templ.guide][templ-guide]). No official benchmark table exists on
  templ.guide.
- gomponents publishes self-reported benchmark numbers in its [README
  FAQ][gomp-readme] ("Is gomponents fast?"; Apple M4, rendering to
  `io.Discard`): ~16 ns/op to construct+render an `El`, and a
  "realistic full page" at ~63.6 µs/op, 126 KB, 2858 allocs/op. The
  repo carries an extensive benchmark suite (`gomponents_benchmark_test.go`,
  `html/benchmark_test.go`, `components/benchmark_test.go`), and a
  first-party [diary entry][gomp-diary] (`2026-03-25-benchmark-optimization.md`)
  documents an optimization pass concluding that ~73% of CPU in that
  workload was GC/runtime overhead, so allocation count dominates.
- Architecturally, templ's generated code writes directly with known
  strings (compile-time layout), while gomponents builds a runtime
  closure tree per render — but neither project publishes a head-to-head
  comparison, and any claim that one "is faster" would be speculation at
  the library level. In practice both are comfortably fast for
  server-rendered HTML; gomponents' own profiling suggests GC pressure
  is the only lever that matters.

### 8. Testing

- **templ** documents two first-party strategies (testing section of
  [Components][templ-components]): expectation testing with `goquery`
  over rendered output (piping `Render` into
  `goquery.NewDocumentFromReader`, using `data-testid` attributes), and
  snapshot testing with the `generator/htmldiff` package
  (`htmldiff.Diff`, used by templ itself for release regression
  checks). The pure-function component guidance (§4) is what makes both
  strategies viable.
- **gomponents** testing is trivially ordinary Go: `Node` implements
  `fmt.Stringer` (`NodeFunc.String()` in `gomponents.go`), and the
  first-party testing pattern is render-to-buffer-then-compare-strings
  ([llms.txt][gomp-llms], "Testing"). Useful rendering semantics from
  the same source: nil `Node`s are safely ignored, and attributes
  render in the order specified. The project itself claims "fully
  tested with 100% coverage" ([README, FAQ][gomp-readme], "Is
  gomponents production-ready?") and ships fuzz tests
  (`components/components_fuzz_test.go`), but provides no testing guide
  — there is little to teach beyond Go norms.

### 9. Maturity and ecosystem (first-party observable facts)

| Fact             | templ                     | gomponents                  |
| ---------------- | ------------------------- | --------------------------- |
| First commit     | 2021-04-10                | 2020-09-12                  |
| GitHub stars     | 10,576 (2026-10-09)       | 1,905 (2026-10-09)          |
| Latest release   | v0.3.1070 (2026-10-04)    | v1.4.0 (2026-10-06)         |
| License          | MIT                       | MIT                         |
| Release cadence  | frequent pre-1.0 releases | stable v1; feature complete |
| Stability stance | pre-1.0 API may change    | no breaking changes         |

Notes on the last two rows:

- templ is pre-1.0; a v1→v2 syntax migration exists
  ([FAQ][templ-faq], "How can I migrate from templ version 0.1.x to
  templ 0.2.x syntax?").
- gomponents: the author considers the core "pretty much feature
  complete" (README FAQ; see [README][gomp-readme]) and promises "Mature
  and stable, no breaking changes" ([README, Features][gomp-readme]);
  experimental `x/...` packages are excepted.

Both are MIT-licensed and actively pushed-to as of October 2026 (GitHub
API). templ has visibly more ecosystem gravity (stars, IDE extensions,
talks, a playground at [play.templ.guide][templ-play]); gomponents
counters with zero-dependency stability and a deliberately frozen core.

### 10. Learning curve

- gomponents: "Simple API that's easy to learn and use (you know most
  already if you know HTML)" ([README, Features][gomp-readme]); the
  author acknowledges the nested-call aesthetic takes adjustment
  ("think of gomponents like a DSL for HTML… Give it a day" — README
  FAQ; see [README][gomp-readme]).
- templ: HTML-shaped syntax reads close to plain HTML, but you learn
  the templ language, its escaping rules, and the generate/LSP
  toolchain. The payoff is richer guardrails (see §5) and first-party
  editor support.

### 11. AI-assisted development (relevant to agent-driven repos)

Both projects ship first-party material for coding agents:

- **templ** maintains a single consolidated doc for LLM assistants at
  [templ.guide/llms.md][templ-llms] (llms.txt convention), explicitly
  intended for Copilot/Cursor-style tools ([llms.md][templ-llms],
  "Coding assistants / LLMs").
- **gomponents** publishes [gomponents.com/llms.txt][gomp-llms] *and* a
  curated agent skill in the repo ([skills][gomp-skills],
  `skills/gomponents/SKILL.md`), installable as a Claude Code / Codex
  plugin (README; see [README][gomp-readme]). The skill encodes
  gomponents conventions (dot imports, `Node` composition,
  `Map`/`If`/`Iff`, HTML5 documents) that ordinary Go habits get wrong.

Practical upshot: both are well-documented for agents; gomponents is
plain Go so agents need no format-specific training to *read* the code,
while templ's `.templ` syntax benefits from its llms.md being loaded.

---

## Evidence gaps

- **No authoritative head-to-head benchmark** of templ vs gomponents
  exists in either first-party source. gomponents' numbers are
  self-reported and measure only itself; templ publishes none. Do not
  cite performance as a differentiator without running your own
  benchmark on your own workload.
- templ.guide had no dedicated "performance" page at time of writing
  (checked 2026-10-09 via templ.guide docs index and search).
- Popularity/ecosystem comparisons beyond GitHub counts (stars,
  releases) are impressions, not measured facts; community libraries
  for both were not catalogued.

## Sources (primary)

- Canonical LLM doc bundles: [templ.guide/llms.md][templ-llms] (full
  templ docs, consolidated first-party),
  [gomponents.com/llms.txt][gomp-llms] (consolidated first-party
  gomponents docs)
- templ docs: [Introduction][templ-guide], [Components][templ-components],
  [Basic syntax][templ-basic], [Template composition][templ-composition],
  [Context][templ-context], [Injection attacks][templ-injection],
  [Content security policy][templ-csp], [CLI][templ-cli],
  [Live reload][templ-reload], [IDE support][templ-ide],
  [Installation][templ-install], [FAQ][templ-faq]
- templ repo (first-party): [a-h/templ][templ-repo],
  issues [#6][templ-issue-6], [#908][templ-issue-908],
  [discussion #206][templ-disc-206]
- gomponents (first-party): [README][gomp-readme],
  [package docs][gomp-pkg], source (`gomponents.go`,
  [`http/handler.go`][gomp-http], `go.mod`), [agent skill][gomp-skills],
  [benchmark diary][gomp-diary]
- GitHub REST API (repos/releases endpoints), queried 2026-10-09

## Link references

[templ-guide]: https://templ.guide
[templ-llms]: https://templ.guide/llms.md
[templ-components]: https://templ.guide/core-concepts/components/
[templ-basic]: https://templ.guide/syntax-and-usage/basic-syntax/
[templ-composition]: https://templ.guide/syntax-and-usage/template-composition
[templ-context]: https://templ.guide/syntax-and-usage/context/
[templ-injection]: https://templ.guide/security/injection-attacks/
[templ-csp]: https://templ.guide/security/content-security-policy/
[templ-cli]: https://templ.guide/developer-tools/cli/
[templ-reload]: https://templ.guide/developer-tools/live-reload/
[templ-ide]: https://templ.guide/developer-tools/ide-support/
[templ-install]: https://templ.guide/quick-start/installation/
[templ-faq]: https://templ.guide/faq/
[templ-play]: https://play.templ.guide
[templ-repo]: https://github.com/a-h/templ
[templ-issue-6]: https://github.com/a-h/templ/issues/6
[templ-issue-908]: https://github.com/a-h/templ/issues/908
[templ-disc-206]: https://github.com/a-h/templ/discussions/206
[gomp-readme]: https://github.com/maragudk/gomponents
[gomp-llms]: https://www.gomponents.com/llms.txt
[gomp-pkg]: https://pkg.go.dev/maragu.dev/gomponents
[gomp-http]: https://github.com/maragudk/gomponents/blob/main/http/handler.go
[gomp-skills]: https://github.com/maragudk/gomponents/tree/main/skills
[gomp-diary]: https://github.com/maragudk/gomponents/blob/main/docs/diary
