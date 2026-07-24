#!/usr/bin/env bash
#
# Reports total and maintained-code statement coverage from a go test
# coverage profile, then enforces the maintained-code floor recorded in
# scripts/coverage-baseline.env. "Maintained" excludes the ogen and sqlc
# generated packages (kept in sync with the Makefile's generate-check
# target) so generated boilerplate cannot inflate or hide real coverage
# regressions.
#
# Usage: check-coverage.bash <raw-profile> <maintained-profile-out>

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

if [[ $# -ne 2 ]]; then
  printf 'usage: %s <raw-profile> <maintained-profile-out>\n' "$0" >&2
  exit 1
fi

raw_profile="$1"
maintained_profile="$2"

if [[ ! -s "${raw_profile}" ]]; then
  printf 'coverage: profile not found or empty: %s (run "make cover")\n' "${raw_profile}" >&2
  exit 1
fi

# Generated packages, kept in sync with the Makefile's generate-check target.
generated_paths=(
  "internal/api/v1/"
  "internal/platform/database/postgres/db/"
)
exclude_pattern="$(IFS='|'; echo "${generated_paths[*]}")"

{
  head -n 1 "${raw_profile}"
  tail -n +2 "${raw_profile}" | grep -v -E "${exclude_pattern}" || true
} >"${maintained_profile}"

echo "Total coverage (including generated code):"
go tool cover -func="${raw_profile}"

echo
echo "Maintained coverage (excluding ${generated_paths[*]}):"
maintained_report="$(go tool cover -func="${maintained_profile}")"
printf '%s\n' "${maintained_report}"

maintained_total="$(printf '%s\n' "${maintained_report}" | awk '/^total:/ { print $NF }' | tr -d '%')"
if [[ -z "${maintained_total}" ]]; then
  printf 'coverage: could not parse a maintained-code total from "go tool cover -func"\n' >&2
  exit 1
fi

source "${PROJECT_DIR}/scripts/coverage-baseline.env"
if [[ -z "${MAINTAINED_COVERAGE_MIN:-}" ]]; then
  printf 'coverage: MAINTAINED_COVERAGE_MIN is not set in scripts/coverage-baseline.env\n' >&2
  exit 1
fi

echo
if awk -v actual="${maintained_total}" -v floor="${MAINTAINED_COVERAGE_MIN}" \
    'BEGIN { exit (actual + 0 >= floor + 0) ? 0 : 1 }'; then
  printf 'coverage: maintained-code coverage %s%% meets the %s%% floor\n' \
    "${maintained_total}" "${MAINTAINED_COVERAGE_MIN}"
else
  printf 'coverage: maintained-code coverage %s%% is below the %s%% floor in scripts/coverage-baseline.env\n' \
    "${maintained_total}" "${MAINTAINED_COVERAGE_MIN}" >&2
  exit 1
fi
