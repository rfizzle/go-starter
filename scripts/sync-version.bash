#!/usr/bin/env bash
set -euo pipefail

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"
source "${PROJECT_PATH}/scripts/version.bash"

# A git tag (`vX.Y.Z`) is the single source of truth for the release version
# (see scripts/version.bash). This script resolves it and propagates it into
# every file that must agree: the checked-in `.version` fallback and the
# OpenAPI spec. The build ldflags (scripts/build.bash) and the Docker image
# tags (scripts/docker-build.bash) call read_version directly, so they always
# see the same value without needing this script to run first.
VERSION=$(read_version "$PROJECT_PATH")
echo -e "${MAGENTA}Updating files to use version: ${VERSION}...${RESET}"

# `.version` — a derived fallback for contexts without git history (shallow
# clones, exported tarballs). Not hand-edited; regenerated here.
printf '%s' "${VERSION}" > "${PROJECT_PATH}/.version"

# OpenAPI spec `info.version` (the API version surfaced to clients). Anchored to
# the two-space `info:` indentation so nested `version:` scalars elsewhere in the
# spec are never touched, and the value is quoted so bare numerics like 1.0 stay
# strings rather than being parsed as YAML floats.
sed -i.tmp "s#^\(  version: \).*\$#\1\"${VERSION}\"#" "${PROJECT_PATH}/api/openapi.yml"
rm -f "${PROJECT_PATH}/api/openapi.yml.tmp"
