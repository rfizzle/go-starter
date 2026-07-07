#!/usr/bin/env bash

# Shared version helper. A git tag (`vX.Y.Z`) is the single source of truth for
# the release version; every script that needs the version sources this file and
# calls read_version so the parse is identical everywhere (build ldflags, Docker
# tags, and sync-version propagation).
#
# `.version` is a checked-in, derived fallback — not the source of truth. It
# exists for contexts that can't see git history (a shallow clone, an exported
# tarball, or a Docker build stage without a `.git` directory). `make
# sync-version` writes the resolved git-tag version into it, so it should never
# be hand-edited.

# read_version <project_path>
# Prints the release version, preferring the most recent `vX.Y.Z`-style git tag
# reachable from HEAD. Falls back to <project_path>/.version when no such tag
# exists (or git/history isn't available). A leading "v" is stripped; internal
# characters are preserved so prerelease/build metadata like "1.0.0-dev" survives.
read_version() {
  local project_path="$1"
  local tag
  if tag=$(git -C "${project_path}" describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null); then
    printf '%s' "${tag#v}"
    return
  fi
  if [[ ! -f "${project_path}/.version" ]]; then
    echo "read_version: no vX.Y.Z git tag found and ${project_path}/.version does not exist" >&2
    return 1
  fi
  local raw
  raw=$(tr -d '[:space:],' < "${project_path}/.version")
  if [[ -z "${raw}" ]]; then
    echo "read_version: ${project_path}/.version is empty" >&2
    return 1
  fi
  # Strip only a leading v, not every v in the string.
  printf '%s' "${raw#v}"
}
