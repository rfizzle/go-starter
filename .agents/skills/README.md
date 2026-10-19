# Project skills

This directory is the one place skills live. Each child directory is an
[Agent Skills](https://agentskills.io/specification) skill with a `SKILL.md`
whose frontmatter has a `name` (matching the directory) and a `description`.

```text
.agents/skills/       the skills themselves
.claude/skills -> ../.agents/skills
```

- `.claude/skills` is a relative symlink to this directory, tracked in git.
  Never copy a skill into a client directory; copies drift.
- Shared skills are copies from `../skills`, listed with their source commit
  and content hash in `shared.sum`. `make skills-check` fails on a copy edited
  here; edit the source, commit it there, then `make skills-sync`. Add a new
  one with `make skills-sync SKILLS="<name>"`.
- Project skills are edited here. `make skills-lint` checks their frontmatter,
  links, size and client-neutral wording (`scripts/validate-skills.mjs` in
  this directory); it skips the shared ones.
- A skill is worth adding when a tool or a reviewer acts on it. Descriptions
  name the paths or situations that trigger the skill.

| Skill | Shared | Use it for |
|---|---|---|
| `project-docs` | yes | The document set, comments, where a fact belongs |
| `sound-patterns` | yes | Settling style questions against the Go stdlib |
| `go-conformance-suite` | yes | One test suite over several implementations of an interface |
| `golang-benchmark` | yes | Writing and comparing benchmarks, pprof |
| `golang-concurrency` | yes | Goroutines, channels, locks, worker pools |
| `golang-context` | yes | Context propagation, cancellation, deadlines |
| `golang-design-patterns` | yes | Constructors, options, lifecycle, shutdown |
| `golang-error-handling` | yes | Defining, wrapping and translating errors |
| `golang-observability` | yes | Logs, metrics, traces, profiling endpoints |
| `golang-performance` | yes | Fixing a measured bottleneck |
| `golang-security` | yes | Injection, crypto, secrets, untrusted input |
| `golang-spf13-cobra` | yes | Cobra command trees and flags |
| `golang-spf13-viper` | yes | Viper configuration and precedence |
| `golang-testing` | yes | Test design, integration tests, fuzzing |
| `golang-troubleshooting` | yes | Finding the root cause of a defect |
| `go-project-layout` | | Where a package goes under `internal/`; `make architecture-check` |
| `golang-database` | | sqlc, squirrel, Goose and pgx changes with their make targets |
| `podman-operations` | | Local PostgreSQL, testcontainers and image builds with Podman |
