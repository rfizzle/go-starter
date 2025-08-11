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

# Build docker image with tags
docker build "${PROJECT_PATH}" \
  -t "${APPLICATION_NAME}:${GIT_COMMIT}" \
  -t "${APPLICATION_NAME}:${VERSION}" \
  -t "${APPLICATION_NAME}:latest" \
  -f "${PROJECT_PATH}/build/docker/Dockerfile"