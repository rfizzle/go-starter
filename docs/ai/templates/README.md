# AI Documentation Templates

Templates for creating consistent AI documentation that follows this
repository's module/platform boundaries.

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
- In-scope outcomes and non-goals
- Simplest viable Go design and package contracts
- Complexity gates and operational bounds
- Risk-based verification and definition of done

### 3. Feature Template (`feature.md`)
**Use for:** Feature specifications and requirements documentation.

**Key features:**
- User stories and acceptance criteria
- Public contracts and operational bounds
- Explicit scope and non-goals
- Verification mapped to requirements and risks

### 4. Review Template (`review.md`)
**Use for:** Code review documentation with consistent structure and quality standards.

**Key features:**
- Structured issue categorization (Critical/Major/Minor/Positive)
- Go-specific review criteria and checklists
- Actionable recommendations and next steps
- Review quality validation guidelines


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

### Step 4: Validate Using Template Checklist
Each template includes a focused validation checklist covering:

**Content Completeness**
- All `[PLACEHOLDER]` values replaced
- Template comments removed
- Required sections completed

**Technical Accuracy**
- Go patterns and conventions followed
- File paths and project structure correct
- Integration with existing systems considered

**Quality Standards**
- Requirements are specific and measurable
- Tests protect named contracts, boundaries, regressions, or risks
- New complexity and resilience mechanisms are justified

**AI Agent Effectiveness** *(for prompts)*
- Instructions are actionable and clear
- Scope boundaries are enforceable
- Quality gates are measurable

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
- Module/platform placement and import-direction considerations
- Contract-first package boundaries
- Consumer-owned interfaces and Go conventions

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
1. Reference core Go principles from `docs/development/foundations.md`
2. Use scope enforcement consistently
3. Provide relevant context from development documentation
4. Keep documents current with project evolution
5. Follow the bounded workflow: feature → plan → smallest implementation → review → fix
6. Use template validation checks in agent prompts

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

### Code Review Documentation
```bash
# Creating a code review document
cp docs/ai/templates/review.md docs/reviews/auth-package-review.md
# Document review findings with structured feedback
```

---

*These templates enable effective AI-human collaboration in Go development projects while maintaining consistency and quality standards.*
