# Builtins: mvdan/sh wins wholesale

Decision record for issue #15, feeding the engine swap in #16 and #17.
The rule, confirmed by the operator: **terebra overrides no bash-named
builtin.** Every builtin `interp` provides is used verbatim, output
formats and error messages included. ADR 0004's conflict rule ("bash
wins on collision") is applied in its strongest form: there are no
renames and no terebra-flavored formats.

## Consequence: deliberate expectation changes

The txtar suite encodes today's custom formats. These expectations
change on purpose when #16 lands (all are bash wins):

| Case              | Today (terebra)      | After (bash/interp)     |
|-------------------|----------------------|-------------------------|
| `alias ll=...`    | `ll=ls -la`          | `alias ll='ls -la'`     |
| `set -x` trace    | `[+] echo hello`     | `+ echo hello`          |
| `help`            | topic system         | bash builtin help       |
| `which echo`      | builtin, prints path | external `which` runs   |
| `echo -n` / `-e`  | printed literally    | honored like bash       |
| `export` (bare)   | raw `VAR=val`        | bash `declare -x` form  |
| `exit foo`        | message, exits 1     | bash message, exit 2    |
| `set` (bare)      | raw environ dump     | bash variable listing   |

Notes:

- No txtar case today depends on the old `echo`, `which`, or `set`
  behaviors; `alias`, `set -x`, and `help` expectations need edits
  (builtins.txtar, options.txtar, help.txtar).
- `type` already matches bash's wording and gains `-t`, `-a`, `-P`
  for free.
- The help topic system is not carried into the builtin; topics live
  in `README.md` and `docs/`.
- `set -o vi` / `set -o emacs` keep working: they are real bash
  options, and their effect (line-editing mode) is wired to terebra's
  readline at the REPL layer.

## The custom residue

Only what mvdan/sh cannot provide:

| Name                       | Why custom                                  |
|----------------------------|---------------------------------------------|
| `jobs`, `fg`, `bg`         | interp returns "unsupported builtin"; needs  |
|                            | OS job tables (issue #12 layer)             |
| `history`, `bind`, `fc`    | interp unsupported; backed by terebra's     |
|                            | readline                                    |
| `drill`, `cue`, `plugin`,  | terebra product commands; no bash analogue  |
| `state`                    |                                             |
| `__terebra_auger__`        | the desugared `|>` stage                    |

Open flag for a human: the `cue` builtin shadows the external `cue`
CLI. Recommendation: keep the name and treat CLI compatibility as a
product promise (superset of the documented subcommands), since CUE is
terebra's identity.

## Confirmation

- `interp.IsBuiltin` + dispatch audit against `mvdan.cc/sh/v3` v3.14.1:
  every bash-named builtin resolves to interp's implementation.
- The txtar updates above are queued in #16's acceptance criteria.

## Link references

[adr-0004]: ../../adrs/0004-terebra-superset-of-mvdan-sh.md
