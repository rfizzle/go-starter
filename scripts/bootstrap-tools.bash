#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/tool-versions.env"

if ! command -v go >/dev/null 2>&1; then
  echo "go could not be found; install the version declared in go.mod first" >&2
  exit 1
fi

tool_bin="${TOOL_BIN:-}"
if [[ -z "${tool_bin}" ]]; then
  tool_bin="$(go env GOBIN)"
fi
if [[ -z "${tool_bin}" ]]; then
  go_path="$(go env GOPATH)"
  tool_bin="${go_path%%:*}/bin"
fi

mkdir -p "${tool_bin}"

install_tool() {
  local name="$1"
  local module="$2"
  local version="$3"

  printf 'Installing %s %s into %s\n' "${name}" "${version}" "${tool_bin}"
  GOBIN="${tool_bin}" go install "${module}@${version}"
}

install_tool goimports "${GOIMPORTS_MODULE}" "${GOIMPORTS_VERSION}"
install_tool golangci-lint "${GOLANGCI_LINT_MODULE}" "${GOLANGCI_LINT_VERSION}"
install_tool yamlfmt "${YAMLFMT_MODULE}" "${YAMLFMT_VERSION}"
install_tool goreleaser "${GORELEASER_MODULE}" "${GORELEASER_VERSION}"
install_tool syft "${SYFT_MODULE}" "${SYFT_VERSION}"
install_tool sqlc "${SQLC_MODULE}" "${SQLC_VERSION}"
install_tool goose "${GOOSE_MODULE}" "${GOOSE_VERSION}"
install_tool govulncheck "${GOVULNCHECK_MODULE}" "${GOVULNCHECK_VERSION}"

printf 'Pinned development tools are installed in %s\n' "${tool_bin}"
