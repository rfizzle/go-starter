#!/usr/bin/env bash

# Shared version helper. `.version` is the single source of truth for the release
# version; every script that needs the version sources this file and calls
# read_version so the parse is identical everywhere (build ldflags, Docker tags,
# and sync-version propagation).

# read_version <project_path>
# Prints the release version from <project_path>/.version, trimming surrounding
# whitespace and a single leading "v" prefix (e.g. "v1.2.3" -> "1.2.3"). Internal
# characters are preserved so prerelease/build metadata like "1.0.0-dev" survives.
read_version() {
  local project_path="$1"
  local raw
  raw=$(tr -d '[:space:],' < "${project_path}/.version")
  # Strip only a leading v, not every v in the string.
  printf '%s' "${raw#v}"
}
