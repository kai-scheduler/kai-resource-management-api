#!/usr/bin/env bash
# Copyright 2026 NVIDIA CORPORATION
# SPDX-License-Identifier: Apache-2.0

# Turns "## [Unreleased]" in CHANGELOG.md into "## [VERSION] - <today>" and opens a
# new, empty "## [Unreleased]" above it.
#
# Usage: hack/changelog-release.sh vX.Y.Z [CHANGELOG.md]

set -euo pipefail

version="${1:-}"
changelog="${2:-CHANGELOG.md}"

if ! [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: version must look like vX.Y.Z, got '$version'" >&2
  exit 1
fi
if grep -qF "## [$version]" "$changelog"; then
  echo "error: $changelog already has a $version section" >&2
  exit 1
fi
if ! grep -qx '## \[Unreleased\]' "$changelog"; then
  echo "error: $changelog has no '## [Unreleased]' heading" >&2
  exit 1
fi

# A release without entries would publish blank notes.
entries="$(awk '
  /^## \[Unreleased\]$/ { inside = 1; next }
  /^## \[/ { inside = 0 }
  inside && /^- / { count++ }
  END { print count + 0 }
' "$changelog")"
if [ "$entries" -eq 0 ]; then
  echo "error: '## [Unreleased]' in $changelog has no entries to release" >&2
  exit 1
fi

today="$(date -u +%Y-%m-%d)"
tmp="$(mktemp "$changelog.XXXXXX")"
awk -v heading="## [$version] - $today" '
  /^## \[Unreleased\]$/ && !done { print; print ""; print heading; done = 1; next }
  { print }
' "$changelog" > "$tmp"
mv "$tmp" "$changelog"

echo "Moved $entries entries under $version in $changelog."
