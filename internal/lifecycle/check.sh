#!/bin/sh
# Check command-lifecycle policy.
#
# Rules (see docs/agents/command-lifecycle.md and docs/adrs/0002):
#   1. Every exp/* directory holding Go sources is its own Go module.
#   2. Every cmd/* directory is a Go module whose latest version tag is
#      v1.0.0 or later: a command may enter cmd/ only by promotion.
#   3. No Go source or go.mod outside exp/ depends on exp/ packages/modules:
#      dependencies point exp -> stable only.
#   4. Experiments do not depend on one another.
set -eu

root=$(git rev-parse --show-toplevel)
cd "$root"

fail=0
err() { echo "check: $*" >&2; fail=1; }

# Rule 1: every exp dir with Go sources is a module.
for d in exp/*/; do
  [ -d "$d" ] || continue
  if find "$d" -name '*.go' -print -quit | grep -q .; then
    [ -f "$d/go.mod" ] || err "$d holds Go sources but no go.mod (every experiment is its own module)"
  fi
done

# Rule 2: cmd/ modules exist only past promotion (latest tag >= v1.0.0).
for d in cmd/*/; do
  [ -d "$d" ] || continue
  n=$(basename "$d")
  if [ ! -f "$d/go.mod" ]; then
    err "cmd/$n is not a Go module"
    continue
  fi
  latest=$(git tag -l "cmd/$n/v[0-9]*" | sort -V | tail -1)
  if [ -z "$latest" ]; then
    err "cmd/$n has no version tags (a command may enter cmd/ only by promotion)"
    continue
  fi
  floor="cmd/$n/v1.0.0"
  [ "$(printf '%s\n%s\n' "$floor" "$latest" | sort -V | tail -1)" = "$latest" ] \
    || err "cmd/$n latest tag $latest is below v1.0.0 (pre-1.0 commands live in exp/)"
done

# Rule 3: dependencies point exp -> stable only (including go.mod tool pins).
if hits=$(git grep -l 'github.com/qjcg/arcadia/exp/' -- '*.go' ':(exclude)exp/**' ':(exclude)**/testdata/**') && [ -n "$hits" ]; then
  err "stable Go source imports exp/ packages (dependencies must point exp -> stable only):"
  echo "$hits" | sed 's/^/check:   /' >&2
fi
if hits=$(find . -type f -name go.mod -not -path './exp/*' -not -path './.git/*' -exec grep -l 'github.com/qjcg/arcadia/exp/' {} + 2>/dev/null) && [ -n "$hits" ]; then
  err "Go modules outside exp/ require an exp/ module (including tool pins):"
  echo "$hits" | sed 's/^/check:   /' >&2
fi

# Rule 4: experiments are independent; one experiment cannot depend on another.
for f in exp/*/go.mod; do
  [ -f "$f" ] || continue
  module=$(awk '$1 == "module" { print $2; exit }' "$f")
  refs=$(grep -n 'github.com/qjcg/arcadia/exp/' "$f" | grep -Fv ":module $module" || true)
  if [ -n "$refs" ]; then
    err "$f depends on another experiment module (experiments must not depend on each other)"
  fi
done

if [ "$fail" -ne 0 ]; then
  echo "check: command-lifecycle policy check FAILED" >&2
  exit 1
fi
echo "check: command-lifecycle policy OK"
