#!/usr/bin/env bash

set -euo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
module_path="$(awk '$1 == "module" { print $2; exit }' "${project_dir}/go.mod")"
failures=0
architecture_graph="$(mktemp -t go-starter-architecture.XXXXXX)"
trap 'rm -f "${architecture_graph}"' EXIT

if [[ -z "${module_path}" ]]; then
  printf 'architecture check: go.mod has no module directive\n' >&2
  exit 1
fi

report() {
  printf 'architecture violation: %s\n' "$1" >&2
  failures=$((failures + 1))
}

package_dependencies() {
  (
    cd "${project_dir}"
    go list -f '{{.ImportPath}}{{range .Imports}} {{.}}{{end}}{{range .TestImports}} {{.}}{{end}}{{range .XTestImports}} {{.}}{{end}}' "$@"
  )
}

package_dependencies ./... >"${architecture_graph}"

while IFS=' ' read -r package dependencies; do
  case "${package}" in
    "${module_path}/internal/"*)
      internal_path="${package#${module_path}/internal/}"
      top_level="${internal_path%%/*}"
      case "${top_level}" in
        api | app | auth | cli | controller | domain | modules | platform | templateinit)
          ;;
        *)
          report "${package} is unclassified; place business code in modules or infrastructure in platform"
          ;;
      esac

      case "${internal_path}" in
        api | modules | platform)
          report "internal/${internal_path} is a grouping directory and must not contain Go files directly"
          ;;
      esac
      ;;
  esac

  for dependency in ${dependencies}; do
    case "${package}" in
      "${module_path}/internal/platform" | "${module_path}/internal/platform/"*)
        case "${dependency}" in
          "${module_path}/internal/domain" | \
            "${module_path}/internal/domain/"* | \
            "${module_path}/internal/modules" | \
            "${module_path}/internal/modules/"*)
            report "${package} must not import ${dependency}; platform is independent of business code"
            ;;
        esac
        ;;
    esac

    case "${package}" in
      "${module_path}/internal/domain" | "${module_path}/internal/domain/"*)
        case "${dependency}" in
          "${module_path}/internal/"*)
            if [[ "${dependency}" != "${module_path}/internal/domain" ]]; then
              report "${package} must not import ${dependency}; domain is an internal leaf"
            fi
            ;;
        esac
        ;;
    esac

    case "${package}" in
      "${module_path}/internal/modules" | "${module_path}/internal/modules/"*)
        case "${dependency}" in
          "${module_path}/internal/api" | "${module_path}/internal/api/"*)
            report "${package} must not import generated transport ${dependency}"
            ;;
        esac
        ;;
    esac

    case "${dependency}" in
      "${module_path}/internal/api" | "${module_path}/internal/api/"*)
        case "${package}" in
          "${module_path}/internal/app" | \
            "${module_path}/internal/auth" | \
            "${module_path}/internal/controller" | \
            "${module_path}/internal/platform/web" | \
            "${module_path}/internal/platform/web/"*)
            ;;
          *)
            report "${package} imports generated transport ${dependency} outside an approved edge package"
            ;;
        esac
        ;;
    esac
  done
done <"${architecture_graph}"

while IFS= read -r directory; do
  report "${directory#${project_dir}/} is a forbidden grab-bag package"
done < <(
  find "${project_dir}/internal" -type d \( \
    -name utils -o \
    -name helpers -o \
    -name common -o \
    -name shared \
  \) -print
)

if ((failures > 0)); then
  printf '%d architecture violation(s) found\n' "${failures}" >&2
  exit 1
fi

printf 'architecture boundaries verified\n'
