# Portable project skills

This directory is the canonical, client-neutral source for project skills. Each child
directory follows the [Agent Skills specification](https://agentskills.io/specification)
and starts with a portable `SKILL.md`.

## Discovery layout

```text
.agents/skills/                   canonical skill content
<client skill directory>/<name>  -> .agents/skills/<name>
```

Clients with native `.agents/skills` discovery can read the canonical directory directly.
For clients that require their own discovery directory, add relative per-skill symlinks
to the canonical directories. Do not copy skills into client directories: copies drift
and make evaluation results ambiguous.

## What belongs where

- Put reusable instructions, references, scripts, assets, and evals in the skill directory.
- Keep `SKILL.md` independent of client-only tool names, slash-command syntax, and agent
  delegation APIs.
- Put an optional client command or prompt harness in that client's own configuration
  directory. Keep it thin: select or sequence the generic skills, pass user input, and
  avoid copying detailed rules that belong in the skill core.
- Prefer standard Agent Skills frontmatter. `name` and `description` are required;
  `license`, `compatibility`, and `metadata` are optional. Avoid client extensions unless
  the skill genuinely cannot function without them.

## Validation

Run the portable suite validator from the repository root or this directory:

```sh
node .agents/skills/scripts/validate-skills.mjs
```

It checks frontmatter, directory/name agreement, skill size, local links, references to
other local skills, long-reference navigation, and the shared eval schema:

```json
{
  "skill_name": "example-skill",
  "evals": []
}
```

Script-specific tests use Node's built-in test runner:

```sh
node --test scripts/validate-skills.test.mjs \
  design-tokens/scripts/validate-tokens.test.mjs \
  component-spec/scripts/validate-spec.test.mjs \
  composition-spec/scripts/validate-spec.test.mjs
```
