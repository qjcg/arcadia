#!/bin/sh
# Retire an experiment.
#
# Usage: retire.sh <name>
#
# Deletes exp/<name>. Retirement is deletion, without ceremony: git history
# and the exp/<name>/v0.x tags preserve everything worth keeping. See
# docs/agents/command-lifecycle.md.
set -eu

name="${1:?usage: retire.sh <name>}"
case "$name" in
  *[!a-z0-9-]*) echo "retire: name must contain only lowercase letters, digits, and hyphens" >&2; exit 1 ;;
esac
src="exp/$name"

die() { echo "retire: $*" >&2; exit 1; }

root=$(git rev-parse --show-toplevel)
cd "$root"

[ -d "$src" ] || die "$src does not exist"
if [ -n "$(git status --porcelain)" ]; then
  die "working tree is not clean (retirement lands its own commit)"
fi

git rm -r -q "$src"
if [ -f go.work ] && grep -q "\./$src" go.work; then
  go work edit -dropuse="./$src"
  go work sync
fi
git add -A
git commit -m "chore: retire $src"
echo "retire: retired $src (deleted; history and tags preserve the past)"
