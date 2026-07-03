#!/usr/bin/env bash

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"
source "${PROJECT_PATH}/scripts/version.bash"

# `.version` is the single source of truth for the release version. This script
# propagates it into every file that must agree with it. The build ldflags
# (scripts/build.bash) and the Docker image tags (scripts/docker-build.bash) also
# read `.version` directly (via read_version), so this only needs to touch files
# that embed a literal.
VERSION=$(read_version "$PROJECT_PATH")
echo -e "${MAGENTA}Updating files to use version: ${VERSION}...${RESET}"

# OpenAPI spec `info.version` (the API version surfaced to clients). Anchored to
# the two-space `info:` indentation so nested `version:` scalars elsewhere in the
# spec are never touched, and the value is quoted so bare numerics like 1.0 stay
# strings rather than being parsed as YAML floats.
sed -i.tmp "s#^\(  version: \).*\$#\1\"${VERSION}\"#" "${PROJECT_PATH}/api/openapi.yml"
rm -f "${PROJECT_PATH}/api/openapi.yml.tmp"

# package.json — kept only so npm tooling reports the right version; it is no
# longer read by the release path. Anchored to the two-space top-level indentation
# so a dependency named "version" would not be rewritten.
sed -i.tmp "s#^\(  \"version\": \"\)[^\"]*\(\"\)#\1${VERSION}\2#" "${PROJECT_PATH}/package.json"
rm -f "${PROJECT_PATH}/package.json.tmp"
