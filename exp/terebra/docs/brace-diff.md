# Brace expansion differential: terebra vs mvdan/sh

Gap report for issue #13, feeding the contract cleanup decided in
[ADR 0004][adr-0004]. It answers one question: can terebra's custom
brace expander be deleted in favor of mvdan/sh's `SplitBraces` and
`BracesSeq`?

## Method

The documented behavior matrix (README promises plus boundary cases)
was run through both implementations: `internal/expand`'s `Expand` and
mvdan/sh's parse → `syntax.SplitBraces` → `expand.BracesSeq` pipeline.
The harness is `internal/expand/bracediff_test.go`; it locks in both
sides' outputs so the gap is pinned until deletion.

## Results

| Case                    | Word             | Match |
|-------------------------|------------------|-------|
| list                    | `{a,b,c}`        | yes   |
| list with affixes       | `pre{a,b}post`   | yes   |
| numeric range           | `{1..5}`         | yes   |
| zero-padded range       | `{01..05}`       | yes   |
| alpha range             | `{a..f}`         | yes   |
| uppercase range         | `{A..Z}`         | yes   |
| stepped range           | `{1..10..2}`     | yes   |
| stepped alpha range     | `{a..z..2}`      | yes   |
| reverse numeric range   | `{5..1}`         | yes   |
| nested list             | `{a,{b,c}}`      | yes   |
| cartesian lists         | `{a,b}{1,2}`     | yes   |
| cartesian range + list  | `{1..2}{a,b}`    | yes   |
| empty element           | `{a,}`           | yes   |
| single element literal  | `{a}`            | **no** |
| empty literal           | `{}`             | yes   |
| unclosed literal        | `x{a,b`          | yes   |

Notes:

- The historical mvdan/sh gaps (zero-padded ranges, uppercase
  ranges [sh-issue-1042]) are fixed as of `mvdan.cc/sh/v3` v3.14.1.
- Ordering matches everywhere: cartesian products vary left-most
  groups slowest in both implementations.

## Divergence

One divergence, on the case bash itself defines:

- `{a}` — terebra expands it to `a`; bash and mvdan/sh keep `{a}`
  literal, because braces expand only when they contain a comma or a
  `..` range. This is terebra deviating from bash, not the reverse.

## Decision

**Go: delete the custom brace expander** when the engine swap lands
(issues #16 and #18), subject to the one accepted divergence:

- `{a}` stays literal per bash. ADR 0004's superset rule applies:
  bash wins on collision. No shim is kept.

## Confirmation

- The differential harness passes with the divergence recorded.
- After deletion, the `testdata/brace.txtar` suite must pass with the
  `{a}` expectation flipped to literal.

## Link references

[adr-0004]: ../../adrs/0004-terebra-superset-of-mvdan-sh.md
[sh-issue-1042]: https://github.com/mvdan/sh/issues/1042
