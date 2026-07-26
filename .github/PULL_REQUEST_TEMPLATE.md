<!--
Keep the pull request focused. Do not mix generated output, dependency
updates, format-only edits, and behavioral changes unless they belong to the
same change. See CONTRIBUTING.md.
-->

## Summary

<!-- What user-visible outcome does this change produce, and why? -->

## Related issues

<!-- e.g. Closes #123. Write "none" if this is not tracked by an issue. -->

## Verification

<!-- Commands you actually ran, with their outcome. -->

- [ ] `make check` passes.
- [ ] Focused package tests pass for the changed code.
- [ ] `make race` (concurrency, lifecycle, middleware, storage, or test
      infrastructure changes).
- [ ] `make test-integration` (database, migration, or generated-API contract
      changes).
- [ ] `make generate` was run and its output is committed (generator input
      changes).

## Checklist

- [ ] The commit subjects are conventional commits, for example
      `feat(auth): add token rotation` (see CONTRIBUTING.md).
- [ ] Generated files under `internal/api/v1/` and
      `internal/platform/database/postgres/db/` were not hand-edited.
- [ ] Documentation, `README.md`, and `AGENTS.md` are updated where the change
      makes them inaccurate.
- [ ] No credentials, private vulnerability details, build artifacts, or
      unrelated worktree changes are included.

## Notes for reviewers

<!-- Migrations, compatibility concerns, follow-up work, or deliberate
non-goals. -->
