#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
while [[ $# -gt 0 ]]; do
  case $1 in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    *)
      echo "Usage: $0 [--dry-run]" >&2
      exit 1
      ;;
  esac
done

# Install sv if missing
if ! command -v sv >/dev/null 2>&1; then
  echo "Installing sv..."
  CGO_ENABLED=0 go install -ldflags '-s -w' ./exp/sv
fi

CHANGES=$(sv next --all)
if [[ -z "$CHANGES" ]]; then
  echo "No version bumps needed."
  [[ $DRY_RUN == true ]] || exit 0
fi

echo "Computed version bumps (from 'sv next --all'):"
echo "$CHANGES"

TAGS=$(echo "$CHANGES" | grep -E '^(.*/)?v[0-9]+\.[0-9]+\.[0-9]+$' || true)
if [[ -z "$TAGS" ]]; then
  echo "No valid semver tags extracted."
  [[ $DRY_RUN == true ]] || exit 0
fi

if [[ $DRY_RUN == true ]]; then
  echo "DRY RUN - Would tag, push, and create GH releases for:"
  echo "$TAGS"
  exit 0
fi

# Real release
git config user.name "github-actions"
git config user.email "github-actions@github.com"

echo "$TAGS" | while read -r tag; do
  [[ -n "$tag" ]] && git tag "$tag"
done

git push origin --tags

echo "$TAGS" | while read -r tag; do
  [[ -n "$tag" ]] && gh release create "$tag" --generate-notes
done

echo "Released tags: $TAGS"
