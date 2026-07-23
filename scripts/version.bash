#!/usr/bin/env bash

# Shared version helper. A valid `vX.Y.Z` tag on the current commit is the only
# source of a release version. Untagged local builds are identified explicitly
# as development builds derived from the commit instead of inheriting an old
# release number.

is_release_tag() {
  local tag="$1"
  if [[ ! "${tag}" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$ ]]; then
    return 1
  fi

  local version="${tag#v}"
  if [[ "${version}" != *-* ]]; then
    return 0
  fi

  local prerelease="${version#*-}"
  local identifier
  while IFS= read -r identifier; do
    if [[ "${identifier}" =~ ^[0-9]+$ && "${identifier}" == 0[0-9]* ]]; then
      return 1
    fi
  done < <(tr '.' '\n' <<<"${prerelease}")
}

# read_version <project_path>
# Prints the release version when HEAD has a valid release tag. Otherwise prints
# a deterministic development version based on HEAD.
read_version() {
  local project_path="$1"
  local tag
  if tag=$(git -C "${project_path}" describe --tags --exact-match --match 'v*' HEAD 2>/dev/null); then
    if ! is_release_tag "${tag}"; then
      echo "read_version: tag ${tag} is not a valid v-prefixed semantic version" >&2
      return 1
    fi
    printf '%s' "${tag#v}"
    return
  fi

  local revision
  if ! revision=$(git -C "${project_path}" rev-parse --short=12 HEAD 2>/dev/null); then
    echo "read_version: ${project_path} is not a Git worktree" >&2
    return 1
  fi

  local dirty=""
  if ! git -C "${project_path}" diff --quiet --ignore-submodules --; then
    dirty=".dirty"
  fi
  printf '0.0.0-dev.%s%s' "${revision}" "${dirty}"
}
