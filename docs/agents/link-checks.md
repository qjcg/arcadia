# Markdown link checks

- `task lint:check-links` checks repository-relative Markdown links, images,
  and Markdown anchors offline. It runs on pull requests.
- `task lint:check-links:external` checks HTTP(S) links on demand and in the
  weekly GitHub Actions audit. Run it on trusted Markdown only.

Both tasks default to tracked Markdown. They skip generated plugins and
imported Go Gherkin references, whose links can depend on upstream site layout.
Pass explicit paths after `--` to override that file selection.

The external audit checks each distinct URL once, ignoring fragments, with one
request in flight at a time. It retries transient errors within a fixed budget.
Confirmed 404s, 410s, and redirect loops or exhausted redirect limits fail the
external audit. Access denials, rate limits, and timeouts that prevent
verification are reported separately and do not fail the audit. The external
job does not run on pull requests.
