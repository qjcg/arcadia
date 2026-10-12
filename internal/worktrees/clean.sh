#!/bin/sh
# Remove an agent worktree.
#
# Usage: clean.sh <slug>
#
# Removes ../arcadia.worktrees/<slug>, prunes stale worktree
# registrations, and deletes the worktree's branch. Refuses to remove
# work with commits that are not in main, or a dirty worktree, unless
# FORCE=1 is set. See docs/agents/worktrees.md.
set -eu

slug="${1:?usage: clean.sh <slug>}"

die() { echo "worktree:clean: $*" >&2; exit 1; }

root=$(git rev-parse --show-toplevel)
base=$(dirname "$root")
dir="$base/arcadia.worktrees/$slug"

[ -d "$dir" ] || die "$dir does not exist"

branch=$(git -C "$dir" symbolic-ref --quiet --short HEAD) ||
  die "$dir is in detached-HEAD state; resolve it manually"

if [ "${FORCE:-}" = "1" ]; then
  git -C "$root" worktree remove --force "$dir"
  git -C "$root" branch -D "$branch"
else
  git -C "$root" merge-base --is-ancestor "$branch" main ||
    die "branch $branch has commits not in main; integrate it first, or FORCE=1 to discard"
  git -C "$root" worktree remove "$dir" ||
    die "worktree is dirty; commit or stash first, or FORCE=1 to discard"
  git -C "$root" branch -d "$branch"
fi

git -C "$root" worktree prune
echo "worktree:clean: removed $dir (branch $branch)"
