#!/usr/bin/env bash
set -euo pipefail

# Application variables
APPLICATION_NAME="starter"

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"
source "${PROJECT_PATH}/scripts/version.bash"

# Setup application variables. A git tag (vX.Y.Z) is the single source of truth
# for the release version (see scripts/version.bash). `.version` and
# api/openapi.yml are files derived from it for contexts that read them
# directly; this script instead calls read_version itself, so no separate
# `make sync-version` step is required for the Docker tag to be correct.
GIT_COMMIT=$(git rev-list -1 HEAD)
VERSION=$(read_version "$PROJECT_PATH")

# Derive the Go version from go.mod so the image compiles against the same
# toolchain the module requires (e.g. "go 1.25.7" -> "1.25.7").
GO_VERSION=$(awk '/^go [0-9]/ { print $2; exit }' "${PROJECT_PATH}/go.mod")
if [[ -z "${GO_VERSION}" ]]; then
  echo -e "${RED}Could not derive Go version from go.mod${RESET}" >&2
  exit 1
fi

# Build docker image with tags
docker build "${PROJECT_PATH}" \
  --build-arg "GO_VERSION=${GO_VERSION}" \
  -t "${APPLICATION_NAME}:${GIT_COMMIT}" \
  -t "${APPLICATION_NAME}:${VERSION}" \
  -t "${APPLICATION_NAME}:latest" \
  -f "${PROJECT_PATH}/build/docker/Dockerfile"