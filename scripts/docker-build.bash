#!/usr/bin/env bash

# Application variables
APPLICATION_NAME="starter"

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"

# Setup application variables
GIT_COMMIT=$(git rev-list -1 HEAD)
VERSION=$(grep version < "${PROJECT_PATH}/package.json" \
  | head -1 \
  | awk -F: '{ print $2 }' \
  | sed 's/[",]//g' \
  | tr -d '[:space:]')

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