# AI Documentation Templates

Simple templates for creating consistent AI documentation in Go projects following Standard Go Project Layout.

## Available Templates

### 1. Prompt Template (`prompt.md`)
**Use for:** Creating AI agent role prompts with clear scope and responsibilities.

**Key features:**
- Scope enforcement
- Go workflow guidance
- Quality standards
- Common task examples

### 2. Plan Template (`plan.md`)
**Use for:** Implementation plans for features and development work.

**Key features:**
- Feature analysis and breakdown
- Go-specific implementation patterns
- Test strategy and definition of done
- Risk assessment and rollout plans

### 3. Feature Template (`feature.md`)
**Use for:** Feature specifications and requirements documentation.

**Key features:**
- User stories and acceptance criteria
- Technical requirements
- Implementation guidance
- Success metrics and testing strategy

## How to Use

### Step 1: Copy Template
```bash
cp docs/ai/templates/[template-type].md docs/ai/[target-directory]/[new-file-name].md
```

### Step 2: Replace Placeholders
- Find `[PLACEHOLDER]` text in ALL_CAPS
- Replace with actual content
- Remove template comments (`<!-- -->`)

### Step 3: Customize
- Add or remove sections as needed
- Keep essential structure and formatting
- Maintain scope definitions and quality standards

### Step 4: Validate
- All placeholders replaced
- Links work within project documentation
- Content follows Go project conventions
- Quality standards met

### Step 5: Update Documentation
- Update cross-references
- Update any related documentation

## Design Principles

### Simplicity
- Templates are concise and focused
- Essential sections only
- Clear, actionable content

### Consistency
- Similar structure across templates
- Common sections in predictable places
- Consistent formatting patterns

### Go-Specific
- Standard Go Project Layout considerations
- Go development patterns embedded
- TDD and Go conventions reinforced

### AI-Focused
- Designed for AI consumption
- Clear scope definitions
- Actionable guidance for AI agents

## Best Practices

### Creating New Documents
1. Start with the template - don't create from scratch
2. Follow naming conventions (use established patterns)
3. Be specific - replace placeholders with concrete content
4. Cross-reference related documents appropriately
5. Validate scope definitions are clear

### Using with AI Agents
1. Reference `docs/development/foundations.md` for core principles
2. Use scope enforcement consistently
3. Provide relevant context from `docs/development/` and `docs/architecture/`
4. Keep documents current with project evolution

## Examples

### Prompt Creation
```bash
# Creating a code reviewer prompt
cp docs/ai/templates/prompt.md docs/ai/prompts/reviewer.md
# Customize for code review tasks and quality standards
```

### Feature Documentation
```bash
# Documenting an authentication feature
cp docs/ai/templates/feature.md docs/features/authentication.md
# Capture requirements, acceptance criteria, and implementation guidance
```

---

*These templates enable effective AI-human collaboration in Go development projects while maintaining consistency and quality standards.*