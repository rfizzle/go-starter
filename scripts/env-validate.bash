#!/usr/bin/env bash
ROOT_DIR=$(git rev-parse --show-toplevel)

# Loop through array and ensure the command is installed
for command in go gofmt goimports golangci-lint yamlfmt swagger git grep head awk sed tr echo make; do
  if ! command -v $command &> /dev/null
  then
    echo "$command could not be found"
    exit 1
  fi
done

echo "Build environment is valid"