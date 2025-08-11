#!/usr/bin/env bash

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
source "${PROJECT_PATH}/scripts/colors.bash"

# Loop through array and ensure the command is installed
for command in go gofmt goimports golangci-lint yamlfmt git grep head awk sed tr echo make; do
  if ! command -v $command &> /dev/null
  then
    echo "$command could not be found"
    exit 1
  fi
done

echo "Build environment is valid"