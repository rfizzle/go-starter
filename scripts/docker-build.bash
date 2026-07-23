#!/usr/bin/env bash
set -euo pipefail

# Application variables
APPLICATION_NAME="${APPLICATION_NAME:-starter}"

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"
source "${PROJECT_PATH}/scripts/version.bash"

# Set up application variables. A git tag (vX.Y.Z) is the single source of truth
# for the release version (see scripts/version.bash). `.version` and
# api/openapi.yml are files derived from it for contexts that read them
# directly; this script instead calls read_version itself, so no separate
# `make sync-version` step is required for the Docker tag to be correct.
GIT_COMMIT=$(git -C "${PROJECT_PATH}" rev-parse HEAD)
GIT_BRANCH=$(git -C "${PROJECT_PATH}" symbolic-ref --short -q HEAD || printf 'detached')
BUILD_DATE=$(git -C "${PROJECT_PATH}" show -s --format=%ct HEAD)
BUILD_DATE_RFC3339=$(git -C "${PROJECT_PATH}" show -s --format=%cI HEAD)
VERSION=$(read_version "${PROJECT_PATH}")
LOCAL_ARCH=$(go env GOARCH)
CONTAINER_PLATFORMS="${CONTAINER_PLATFORMS:-linux/${LOCAL_ARCH}}"

case "${LOCAL_ARCH}" in
  amd64 | arm64) ;;
  *)
    echo -e "${RED}Unsupported local architecture: ${LOCAL_ARCH}; expected amd64 or arm64${RESET}" >&2
    exit 1
    ;;
esac

OUTPUT_ARGS=(--load)
if [[ -n "${CONTAINER_OUTPUT:-}" ]]; then
  OUTPUT_ARGS=(--output "${CONTAINER_OUTPUT}")
elif [[ "${CONTAINER_PLATFORMS}" == *,* ]]; then
  echo -e "${RED}Multi-platform builds require CONTAINER_OUTPUT (for example type=registry or type=oci,dest=dist/starter-image.tar)${RESET}" >&2
  exit 1
fi

# Build docker image with tags
docker buildx build "${PROJECT_PATH}" \
  --file "${PROJECT_PATH}/build/docker/Dockerfile" \
  --platform "${CONTAINER_PLATFORMS}" \
  "${OUTPUT_ARGS[@]}" \
  --build-arg "BUILD_BRANCH=${GIT_BRANCH}" \
  --build-arg "BUILD_DATE=${BUILD_DATE}" \
  --build-arg "BUILD_DATE_RFC3339=${BUILD_DATE_RFC3339}" \
  --build-arg "BUILD_ENV=production" \
  --build-arg "BUILD_REVISION=${GIT_COMMIT}" \
  --build-arg "BUILD_VERSION=${VERSION}" \
  -t "${APPLICATION_NAME}:${GIT_COMMIT}" \
  -t "${APPLICATION_NAME}:${VERSION}" \
  -t "${APPLICATION_NAME}:latest"
