#!/usr/bin/env bash

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"

# Get the version from .version
VERSION=$(tr -d '[:space:],v' < "${PROJECT_PATH}/.version")
echo -e "${MAGENTA}Updating swagger files to use version: ${VERSION}...${RESET}"

# Now do the replacement in-place (MacOS/Unix compatible)
REPLACE=' version: .*$'
WITH=" version: \"${VERSION}\""
sed -i.tmp "s#${REPLACE}#${WITH}#g" "${PROJECT_PATH}/api/swagger.yaml"
rm -f "${PROJECT_PATH}/api/swagger.yaml.tmp"

REPLACE=' "version": ".*"'
WITH=" \"version\": \"${VERSION}\""
sed -i.tmp "s#${REPLACE}#${WITH}#g" "${PROJECT_PATH}/internal/api/embedded_spec.go"
rm -f "${PROJECT_PATH}/internal/api/embedded_spec.go.tmp"