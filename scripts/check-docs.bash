#!/usr/bin/env bash

set -euo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
failures=0

report() {
  printf 'documentation check: %s\n' "$1" >&2
  failures=$((failures + 1))
}

documentation_files() {
  find "${project_dir}" \
    \( -path "${project_dir}/.git" \
      -o -path "${project_dir}/.agents" \
      -o -path "${project_dir}/.claude" \
      -o -path "${project_dir}/vendor" \
      -o -path "${project_dir}/dist" \) -prune \
    -o -type f -name '*.md' -print
}

current_guidance_files() {
  documentation_files | while IFS= read -r file; do
    case "${file#${project_dir}/}" in
      # Reviews and the changelog are historical records; they may name
      # retired targets and paths without making them current guidance.
      docs/reviews/* | CHANGELOG.md)
        ;;
      *)
        printf '%s\n' "${file}"
        ;;
    esac
  done
}

while IFS= read -r file; do
  while IFS= read -r markdown_link; do
    target="${markdown_link#*](}"
    target="${target%)}"
    target="${target#<}"
    target="${target%>}"
    target="${target%%#*}"
    target="${target%%\?*}"

    case "${target}" in
      "" | \#* | http://* | https://* | mailto:*)
        continue
        ;;
      /*)
        resolved="${project_dir}${target}"
        ;;
      *)
        resolved="$(dirname -- "${file}")/${target}"
        ;;
    esac

    if [[ ! -e "${resolved}" ]]; then
      report "${file#${project_dir}/}: local link target does not exist: ${target}"
    fi
  done < <(grep -Eo '\[[^][]+\]\([^()]+\)' "${file}" || true)
done < <(documentation_files)

make_targets="$(awk -F: '/^[[:alnum:]_.-]+:/ { print $1 }' "${project_dir}/Makefile")"
while IFS= read -r file; do
  while IFS= read -r documented_target; do
    if ! grep -qx "${documented_target}" <<<"${make_targets}"; then
      report "${file#${project_dir}/}: documents missing Make target: ${documented_target}"
    fi
  done < <(
    {
      grep -Eo '`make [[:alnum:]_.-]+' "${file}" | sed 's/^`make //' || true
      grep -E '^[[:space:]]*make [[:alnum:]_.-]+' "${file}" |
        sed -E 's/^[[:space:]]*make ([[:alnum:]_.-]+).*/\1/' || true
    } | sort -u
  )
done < <(current_guidance_files)

module_go_version="$(awk '$1 == "go" { print $2; exit }' "${project_dir}/go.mod")"
while IFS= read -r file; do
  while IFS= read -r documented_version; do
    if [[ "${documented_version}" != "Go ${module_go_version}" ]]; then
      report "${file#${project_dir}/}: ${documented_version} does not match go.mod (${module_go_version})"
    fi
  done < <(grep -Eo 'Go 1\.[0-9]+(\.[0-9]+)?\+?' "${file}" | sort -u || true)
done < <(current_guidance_files)

required_paths=(
  ".env.example"
  ".github/dependabot.yml"
  ".github/workflows/ci.yml"
  ".github/workflows/codeql.yml"
  ".github/workflows/release.yml"
  ".github/workflows/security.yml"
  ".gitleaks.toml"
  ".goreleaser.yaml"
  "api/openapi.yml"
  "build/docker/Dockerfile"
  "build/docker/Dockerfile.release"
  "compose.yaml"
  "internal/api/v1"
  "internal/app"
  "internal/controller"
  "internal/domain"
  "internal/modules"
  "internal/platform"
  "internal/platform/database/postgres/db"
  "internal/platform/database/postgres/migrations"
  "internal/platform/database/postgres/queries"
  "scripts/container-build.bash"
  "scripts/coverage-baseline.env"
  "scripts/tool-versions.env"
  "sqlc.yaml"
  "tools/initialize"
  "web/build/index.html"
)

for required_path in "${required_paths[@]}"; do
  if [[ ! -e "${project_dir}/${required_path}" ]]; then
    report "required documented path does not exist: ${required_path}"
  fi
done

retired_pattern='Dockerfile\.dev|pkg/config|internal/platform/(middleware|web)|make (clean-gen|test-cover|test-race|docker)|docker compose |docker-build\.bash|GOSTARTER_REQUIRE_DOCKER'
while IFS= read -r file; do
  while IFS= read -r retired_reference; do
    report "${file#${project_dir}/}: retired reference: ${retired_reference}"
  done < <(grep -Eo "${retired_pattern}" "${file}" || true)
done < <(current_guidance_files)

if ((failures > 0)); then
  printf '%d documentation issue(s) found\n' "${failures}" >&2
  exit 1
fi

printf 'documentation links, commands, paths, and version verified\n'
