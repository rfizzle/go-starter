---
name: agent-skills
description: Use this skill when creating, scaffolding, updating, reviewing, or validating a reusable Agent Skill that follows the Agent Skills open standard. Trigger for requests involving SKILL.md, agentskills.io, skill directories, skill frontmatter, trigger descriptions, progressive disclosure, bundled scripts/references/assets, or skills-ref validation. Not for one-off prompts, subagent definitions, slash commands, or general project instructions that do not need a reusable skill package.
---

# Create Agent Skills

Create portable, focused skill packages that follow the
[Agent Skills specification](https://agentskills.io/specification). Treat the
current specification as authoritative if it differs from this guidance.

## Workflow

1. Define the capability.
2. Inspect the target environment.
3. Plan the package.
4. Author or update the skill.
5. Validate the complete package.
6. Report the result.

Proceed without clarification when the request, target location, and intended
capability are already clear. Ask only when a missing choice would materially
change the skill.

## 1. Define the capability

Identify:

- The repeatable task the skill should enable.
- Concrete user requests that should activate it.
- Near-miss requests that should not activate it.
- Domain knowledge, procedures, or tools an agent would not reliably infer.
- The requested scope and target directory.

Keep one coherent responsibility per skill. Prefer a reusable procedure over
instructions that merely reproduce one answer.

## 2. Inspect the target environment

Before writing:

- Read repository-level agent instructions and relevant local conventions.
- Inspect nearby skills for established discovery paths and client metadata.
- Inspect the target directory and preserve existing user content.
- Determine whether the skill is new or an update.
- Confirm the intended skill name from the target directory.

Do not overwrite an existing skill blindly. When updating, keep compatible
behavior unless the request explicitly changes it.

## 3. Plan the package

Every skill requires `SKILL.md`. Add optional resources only when they improve
reliability or context efficiency:

```text
skill-name/
├── SKILL.md
├── scripts/       # deterministic or repeatedly recreated operations
├── references/    # detailed information loaded only when needed
└── assets/        # templates and static files used in produced output
```

Additional client-specific files are allowed, but they are not part of the
portable core. Create them only when the target client or repository uses them.

Do not add auxiliary files such as `README.md`, changelogs, or installation
guides unless they directly support the agent executing the skill.

## 4. Author the skill

### Frontmatter

Begin `SKILL.md` with YAML frontmatter containing:

- `name` (required): 1-64 characters; lowercase ASCII letters, digits, and
  hyphens only; no leading, trailing, or consecutive hyphens; exactly matches
  the parent directory.
- `description` (required): 1-1024 characters; explains both what the skill
  does and when to use it. Include concrete intent and domain keywords.
- `license` (optional): a short license name or reference to a bundled license.
- `compatibility` (optional): 1-500 characters describing genuine environment
  requirements.
- `metadata` (optional): a string-to-string map for client-specific metadata.
- `allowed-tools` (optional, experimental): a space-separated string of
  pre-approved tools. Support varies by client.

Use only fields supported by the specification or required by the target
client. A minimal header is:

```yaml
---
name: skill-name
description: Use this skill when [specific user intent]. It [capability and scope].
---
```

Write the description as the activation contract. Put all activation guidance
there because agents normally see the description before loading the body.
State a boundary when similar requests should not activate the skill.

### Body

Write instructions for an agent that already understands general concepts:

- Use imperative language and an executable sequence.
- Include project- or domain-specific facts the agent would otherwise miss.
- Give a clear default when several approaches are possible.
- Be exact where order, safety, or reproducibility matters.
- Include concrete edge cases and corrections in a `Gotchas` section when
  applicable.
- Provide an output template when exact structure matters.
- Reference bundled resources with relative paths from the skill root.
- State exactly when to load each reference or run each script.

Keep `SKILL.md` under 500 lines and preferably under 5,000 tokens. Move
conditional detail into focused, one-level-deep files under `references/`.
Avoid duplicating the same information in the body and a reference.

### Bundled resources

For every script:

- Document inputs, outputs, dependencies, and invocation.
- Make failures actionable and handle expected edge cases.
- Execute representative cases before finalizing.

For every reference:

- Keep it focused and loadable on demand.
- Link it directly from `SKILL.md` with a condition for reading it.
- Add a table of contents when it is long.

For every asset:

- Include only files used to produce the skill's output.
- Explain from `SKILL.md` when and how to use it.

## 5. Validate

Run the official reference validator when available:

```bash
skills-ref validate path/to/skill-name
```

If it is unavailable, manually verify:

- `SKILL.md` exists and contains valid YAML frontmatter followed by Markdown.
- `name` satisfies every naming rule and matches its directory.
- `description` is non-empty, within its limit, and states what and when.
- Optional fields have the specified types and limits.
- Every relative resource path resolves.
- No placeholder text or unused example file remains.
- Scripts run successfully on representative inputs.
- The body is concise and its steps can be followed without hidden context.

Then sanity-check activation using at least:

- Three realistic prompts that should use the skill.
- Two plausible near-miss prompts that should not use it.

For a high-value or frequently reused skill, evaluate a larger fixed set of
positive and negative prompts across multiple runs and refine the description
based on activation errors.

Fix validation failures and repeat until all checks pass.

## 6. Report

Summarize:

- The skill path and files created or changed.
- The capability and activation boundary.
- Validation commands run and their outcomes.
- Any client-specific discovery step that remains.

Do not claim validation that was not actually run.

## Gotchas

- The directory name and frontmatter `name` must match exactly.
- A broad description causes false activation; a vague one prevents activation.
- The body cannot repair missing activation language because it loads later.
- Client-specific discovery files and symlinks do not replace `SKILL.md`.
- The open standard permits optional directories but does not require empty
  `scripts/`, `references/`, or `assets/` folders.
- A reference that is not linked from `SKILL.md` is unlikely to be discovered
  at the right time.
