#!/usr/bin/env bash

# Regenerates the agent discovery symlinks under .claude/.
#
# .agents/skills/ is the canonical, client-neutral skill library and
# docs/ai/prompts/ holds the prompt agents. Claude Code only discovers skills
# under .claude/skills/ and agents under .claude/agents/, so those directories
# hold relative symlinks back into the tracked sources instead of copies, which
# would drift.
#
# The links must be relative: absolute targets resolve only on the machine that
# generated them and break in every clone.
#
# Usage:
#   scripts/setup-agents.bash          regenerate the links
#   scripts/setup-agents.bash --check  verify the links without writing

set -euo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

check_only=false
case "${1-}" in
  --check)
    check_only=true
    ;;
  "")
    ;;
  *)
    printf 'usage: %s [--check]\n' "$0" >&2
    exit 2
    ;;
esac

failures=0

report() {
  printf 'agent link check: %s\n' "$1" >&2
  failures=$((failures + 1))
}

# The skills a Go service developer needs, plus agent-skills for maintaining the
# library itself. Any other canonical skill stays canonical-only.
linked_skills() {
  local entry name
  for entry in "${project_dir}"/.agents/skills/*/; do
    name="$(basename -- "${entry}")"
    case "${name}" in
      backend-spec | go-project-layout | podman-operations | project-conventions | agent-skills | golang-*)
        printf '%s\n' "${name}"
        ;;
    esac
  done
}

linked_prompts() {
  local file
  for file in "${project_dir}"/docs/ai/prompts/*.md; do
    [[ -f "${file}" ]] || continue
    basename -- "${file}"
  done
}

# Each entry is "<link path relative to the repository root>|<symlink target>".
desired=()
while IFS= read -r name; do
  desired+=(".claude/skills/${name}|../../.agents/skills/${name}")
done < <(linked_skills)
while IFS= read -r name; do
  desired+=(".claude/agents/${name}|../../docs/ai/prompts/${name}")
done < <(linked_prompts)

is_desired() {
  local candidate="$1" entry
  for entry in "${desired[@]}"; do
    if [[ "${entry%%|*}" == "${candidate}" ]]; then
      return 0
    fi
  done
  return 1
}

verify_link() {
  local link="$1" target="$2" actual
  local full="${project_dir}/${link}"

  if [[ ! -L "${full}" ]]; then
    report "${link}: missing symlink; run make agents-setup"
    return
  fi

  actual="$(readlink "${full}")"
  if [[ "${actual}" != "${target}" ]]; then
    report "${link}: links to ${actual}, expected ${target}"
    return
  fi

  if [[ ! -e "${full}" ]]; then
    report "${link}: symlink does not resolve"
  fi
}

write_link() {
  local link="$1" target="$2"
  local full="${project_dir}/${link}"

  if [[ -e "${full}" && ! -L "${full}" ]]; then
    printf 'refusing to replace non-symlink: %s\n' "${link}" >&2
    exit 1
  fi

  mkdir -p -- "$(dirname -- "${full}")"
  rm -f -- "${full}"
  ln -s "${target}" "${full}"
}

# Symlinks this script owns are exactly the ones pointing at the two managed
# source trees. Anything else under .claude/ belongs to the developer.
is_managed_target() {
  case "$1" in
    ../../.agents/skills/* | ../../docs/ai/prompts/*)
      return 0
      ;;
  esac
  return 1
}

sweep_existing() {
  local directory="$1" entry link
  [[ -d "${project_dir}/${directory}" ]] || return 0

  for entry in "${project_dir}/${directory}"/*; do
    [[ -L "${entry}" ]] || continue
    link="${directory}/$(basename -- "${entry}")"
    is_desired "${link}" && continue

    if [[ "${check_only}" == true ]]; then
      # An unmanaged link is a local choice; only a dangling one is a defect.
      if [[ ! -e "${entry}" ]]; then
        report "${link}: symlink does not resolve"
      fi
      continue
    fi

    if is_managed_target "$(readlink "${entry}")"; then
      rm -f -- "${entry}"
    fi
  done
}

for entry in "${desired[@]}"; do
  if [[ "${check_only}" == true ]]; then
    verify_link "${entry%%|*}" "${entry#*|}"
  else
    write_link "${entry%%|*}" "${entry#*|}"
  fi
done

sweep_existing ".claude/skills"
sweep_existing ".claude/agents"

if ((failures > 0)); then
  printf '%d agent link issue(s) found\n' "${failures}" >&2
  exit 1
fi

if [[ "${check_only}" == true ]]; then
  printf 'agent discovery links verified (%d)\n' "${#desired[@]}"
else
  printf 'agent discovery links regenerated (%d)\n' "${#desired[@]}"
fi
