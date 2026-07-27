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

# Set up application variables. A valid git tag on HEAD is the single source of
# a release version; untagged local images receive a commit-derived development
# version (see scripts/version.bash).
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

BUILD_ARGS=(
  --file "${PROJECT_PATH}/build/docker/Dockerfile"
  --build-arg "BUILD_BRANCH=${GIT_BRANCH}"
  --build-arg "BUILD_DATE=${BUILD_DATE}"
  --build-arg "BUILD_DATE_RFC3339=${BUILD_DATE_RFC3339}"
  --build-arg "BUILD_ENV=production"
  --build-arg "BUILD_REVISION=${GIT_COMMIT}"
  --build-arg "BUILD_VERSION=${VERSION}"
)

VERSION_TAG="${APPLICATION_NAME}:${VERSION}"

if [[ "${CONTAINER_PLATFORMS}" == *,* ]]; then
  # Podman stores a multi-platform build as a manifest list in local storage.
  # Building into an existing list appends duplicate entries, so any previous
  # list is removed first. The non-native platform needs binfmt emulation
  # (qemu-user-static).
  if podman manifest exists "${VERSION_TAG}"; then
    podman manifest rm "${VERSION_TAG}"
  fi
  podman build "${PROJECT_PATH}" \
    --platform "${CONTAINER_PLATFORMS}" \
    --manifest "${VERSION_TAG}" \
    "${BUILD_ARGS[@]}"
  podman tag "${VERSION_TAG}" \
    "${APPLICATION_NAME}:${GIT_COMMIT}" \
    "${APPLICATION_NAME}:latest"
  if [[ -n "${CONTAINER_ARCHIVE:-}" ]]; then
    podman manifest push --all "${VERSION_TAG}" "oci-archive:${CONTAINER_ARCHIVE}"
  fi
else
  podman build "${PROJECT_PATH}" \
    --platform "${CONTAINER_PLATFORMS}" \
    "${BUILD_ARGS[@]}" \
    -t "${APPLICATION_NAME}:${GIT_COMMIT}" \
    -t "${VERSION_TAG}" \
    -t "${APPLICATION_NAME}:latest"
  if [[ -n "${CONTAINER_ARCHIVE:-}" ]]; then
    podman save --format oci-archive -o "${CONTAINER_ARCHIVE}" "${VERSION_TAG}"
  fi
fi
