#!/usr/bin/env bash
set -euo pipefail

# Compute next patch version based on the latest GitHub release tag of jfrog/jfrog-cli.
# - Uses: gh release view --repo jfrog/jfrog-cli --json tagName
# - Output does NOT include a leading 'v' prefix

if ! command -v gh >/dev/null 2>&1; then
  echo "Error: 'gh' (GitHub CLI) is not installed or not in PATH." >&2
  exit 1
fi

# Get latest release tag from the known repository
tag=$(gh release view --repo jfrog/jfrog-cli --json tagName -q .tagName)

if [ -z "${tag}" ]; then
  echo "Error: Could not determine latest release tag from jfrog/jfrog-cli." >&2
  exit 1
fi

version_core="${tag#v}"

# Extract major, minor, patch from semantic version
IFS='.' read -r major minor patch_raw <<EOF
${version_core}
EOF

if [ -z "${major:-}" ] || [ -z "${minor:-}" ]; then
  echo "Error: Tag '${tag}' is not a valid semver (expected MAJOR.MINOR.PATCH)." >&2
  exit 1
fi

# Strip any suffix from patch (e.g. 1-rc.1 -> 1)
patch=${patch_raw%%[^0-9]*}
patch=${patch:-0}

# Increment patch
next_patch=$((patch + 1))
next_version="${major}.${minor}.${next_patch}"

echo "${next_version}"


