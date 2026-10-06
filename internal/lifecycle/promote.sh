#!/bin/sh
# Promote an experiment to a stable command.
#
# Usage: promote.sh <name>
#
# Moves exp/<name> to cmd/<name>, rewrites the module path and imports, and
# cuts the cmd/<name>/v1.0.0 tag in the same act. See
# docs/agents/command-lifecycle.md for the full promotion checklist: this
# script enforces the mechanical items and performs the move; the judgement
# items (soak, breaking-change debt) are the human's call when running it.
set -eu

name="${1:?usage: promote.sh <name>}"
case "$name" in
  *[!a-z0-9-]*) echo "promote: name must contain only lowercase letters, digits, and hyphens" >&2; exit 1 ;;
esac
src="exp/$name"
dst="cmd/$name"

die() { echo "promote: $*" >&2; exit 1; }

sdir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
root=$(git rev-parse --show-toplevel)
cd "$root"

# --- checklist (mechanical part) ---
[ -d "$src" ] || die "$src does not exist"
[ -f "$src/go.mod" ] || die "$src is not a Go module (every experiment is its own module; see docs/agents/command-lifecycle.md)"
[ ! -e "$dst" ] || die "$dst already exists"
[ -f "$src/README.md" ] || die "checklist: $src/README.md is missing"
[ -f "$src/CHANGELOG.md" ] || die "checklist: $src/CHANGELOG.md is missing (run: task changelogs)"
[ -f go.work ] || die "go.work not found"
if [ -n "$(git status --porcelain)" ]; then
  die "checklist: working tree is not clean (promotion lands its own commit)"
fi
(cd "$src" && go test ./...) || die "checklist: tests are red in $src"
(cd "$src" && go vet ./...) || die "checklist: go vet is red in $src"
sh "$sdir/check.sh" || die "checklist: policy check is red (task check:policy)"

# --- move and rewrite ---
echo "promote: moving $src -> $dst"
mkdir -p cmd
git mv "$src" "$dst"
echo "promote: rewriting module path and imports"
git grep -l "github.com/qjcg/arcadia/exp/$name" \
  | grep -Ev '(^|/)CHANGELOG\.md$|\.sum$' \
  | xargs -r sed -i "s|github.com/qjcg/arcadia/exp/$name|github.com/qjcg/arcadia/cmd/$name|g"
go work edit -dropuse="./$src"
go work edit -use="./$dst"
(cd "$dst" && go mod tidy)

# --- verify and land ---
(cd "$dst" && go test ./...) || die "tests are red after the move; finish by hand (see docs/agents/command-lifecycle.md)"
(cd "$dst" && go vet ./...) || die "go vet is red after the move; finish by hand (see docs/agents/command-lifecycle.md)"
git add -A
git commit -m "feat($name)!: promote $src to $dst at v1.0.0"
git tag -a "cmd/$name/v1.0.0" -m "release cmd/$name v1.0.0"
echo "promote: promoted $src -> $dst at v1.0.0"
echo "promote: tag created locally; push it through the repository's normal review and release workflow"
