#!/usr/bin/env bash

set -uo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/tool-versions.env"

failures=0

report_present() {
  local name="$1"
  local installed="$2"
  local required="$3"

  printf '%-16s installed: %-14s required: %s\n' "${name}" "${installed}" "${required}"
}

report_missing() {
  local name="$1"
  local required="$2"

  report_present "${name}" "missing" "${required}"
  failures=$((failures + 1))
}

version_at_least() {
  local installed="${1#go}"
  local required="${2#go}"

  awk -v installed="${installed}" -v required="${required}" 'BEGIN {
    split(installed, actual, ".")
    split(required, wanted, ".")
    for (i = 1; i <= 3; i++) {
      actual[i] += 0
      wanted[i] += 0
      if (actual[i] > wanted[i]) {
        exit 0
      }
      if (actual[i] < wanted[i]) {
        exit 1
      }
    }
    exit 0
  }'
}

check_command() {
  local name="$1"

  if command -v "${name}" >/dev/null 2>&1; then
    report_present "${name}" "present" "present"
    return
  fi

  report_missing "${name}" "present"
}

check_go() {
  local required
  required="$(awk '$1 == "go" {print "go" $2; exit}' "${PROJECT_DIR}/go.mod")"

  if ! command -v go >/dev/null 2>&1; then
    report_missing go "${required} or newer"
    return
  fi

  local installed
  installed="$(go env GOVERSION)"
  report_present go "${installed}" "${required} or newer"
  if ! version_at_least "${installed}" "${required}"; then
    failures=$((failures + 1))
  fi
}

check_go_tool() {
  local name="$1"
  local required_module="$2"
  local required_version="$3"

  if ! command -v "${name}" >/dev/null 2>&1; then
    report_missing "${name}" "${required_version}"
    return
  fi

  local binary
  binary="$(command -v "${name}")"

  local metadata
  metadata="$(go version -m "${binary}" 2>/dev/null)"

  local installed_module
  installed_module="$(awk '$1 == "mod" {print $2; exit}' <<<"${metadata}")"

  local installed_version
  installed_version="$(awk '$1 == "mod" {print $3; exit}' <<<"${metadata}")"

  if [[ -z "${installed_version}" ]]; then
    installed_version="unknown"
  fi

  report_present "${name}" "${installed_version}" "${required_version}"
  if [[ "${installed_module}" != "${required_module%/cmd/*}" ]] ||
    [[ "${installed_version}" != "${required_version}" ]]; then
    failures=$((failures + 1))
  fi
}

check_go
check_command gofmt
check_command git
check_command make
check_command docker
# make skills-check runs the Node-based skill-library validator.
check_command node

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  report_present "docker compose" "present" "present"
else
  report_missing "docker compose" "present"
fi

if command -v go >/dev/null 2>&1; then
  check_go_tool goimports "${GOIMPORTS_MODULE}" "${GOIMPORTS_VERSION}"
  check_go_tool golangci-lint "${GOLANGCI_LINT_MODULE}" "${GOLANGCI_LINT_VERSION}"
  check_go_tool yamlfmt "${YAMLFMT_MODULE}" "${YAMLFMT_VERSION}"
  check_go_tool goreleaser "${GORELEASER_MODULE}" "${GORELEASER_VERSION}"
  check_go_tool syft "${SYFT_MODULE}" "${SYFT_VERSION}"
  check_go_tool sqlc "${SQLC_MODULE}" "${SQLC_VERSION}"
  check_go_tool goose "${GOOSE_MODULE}" "${GOOSE_VERSION}"
  check_go_tool govulncheck "${GOVULNCHECK_MODULE}" "${GOVULNCHECK_VERSION}"

  ogen_version="$(
    cd -- "${PROJECT_DIR}" &&
      go list -m -f '{{.Version}}' github.com/ogen-go/ogen
  )"
  report_present ogen "${ogen_version}" "${ogen_version} (go.mod)"
else
  report_missing ogen "version pinned in go.mod"
fi

if ((failures > 0)); then
  printf 'Build environment has %d missing or incompatible requirement(s).\n' "${failures}" >&2
  printf 'Run make tools, ensure its bin directory is on PATH, and retry.\n' >&2
  exit 1
fi

echo "Build environment is valid"
