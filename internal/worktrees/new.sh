#!/bin/sh
# Create an agent worktree.
#
# Usage: new.sh <slug> [branch]
#
# Creates ../arcadia.worktrees/<slug> on <branch> (default: <slug>).
# Branches off main when the branch does not exist yet; reattaches an
# existing branch so a crashed session can pick its work back up. See
# docs/agents/worktrees.md for the convention this enforces.
set -eu

slug="${1:?usage: new.sh <slug> [branch]}"
branch="${2:-$slug}"
case "$slug" in
  *[!a-z0-9-]*) echo "worktree:new: slug must contain only lowercase letters, digits, and hyphens" >&2; exit 1 ;;
  -*) echo "worktree:new: slug must not start with a hyphen" >&2; exit 1 ;;
esac

die() { echo "worktree:new: $*" >&2; exit 1; }

root=$(git rev-parse --show-toplevel)
base=$(dirname "$root")
dir="$base/arcadia.worktrees/$slug"

[ ! -e "$dir" ] || die "$dir already exists"

if git -C "$root" show-ref --verify --quiet "refs/heads/$branch"; then
  git -C "$root" worktree add "$dir" "$branch"
  echo "worktree:new: reattached existing branch $branch at $dir"
else
  git -C "$root" worktree add -b "$branch" "$dir" main
  echo "worktree:new: created branch $branch off main at $dir"
fi

echo "worktree:new: work in $dir; it is branch-only territory (see docs/agents/worktrees.md)"
