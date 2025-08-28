# AI Agent Documentation

This directory contains AI agent configurations, prompts, templates, and reference materials for consistent AI interactions with the Go Starter codebase.

## Overview

The AI documentation system provides structured guidance for AI agents working on this Go project. It includes role-specific prompts, reusable templates, and comprehensive reference materials.

## Directory Structure

```
ai/
├── README.md                    # This file - AI system overview
├── agent-responsibility-matrix.md # Agent roles and boundaries
├── escalation-criteria.md       # When to escalate beyond AI capabilities
├── prompts/                     # AI agent role definitions
│   ├── ai.md                   # Documentation agent
│   ├── feature.md              # Product owner agent
│   ├── plan.md                 # Implementation planning agent
│   ├── review.md               # Code review agent
│   └── implement.md            # Three-phase implementation agent
└── templates/                   # Reusable document templates
    ├── README.md               # Template usage guide
    ├── feature.md              # Feature specification template
    ├── plan.md                 # Implementation plan template
    └── prompt.md               # AI agent prompt template
```

## Agent Roles

### AI Documentation Agent (`prompts/ai.md`)
- **Scope**: Documentation within `docs/ai/` directory
- **Responsibilities**: Create and refine prompts, templates, AI-specific documentation
- **Boundaries**: Cannot modify code or non-AI documentation

### Feature Agent (`prompts/feature.md`)
- **Scope**: Feature specifications and requirements
- **Responsibilities**: Create user stories, acceptance criteria, technical requirements
- **Boundaries**: Planning only, no implementation

### Planning Agent (`prompts/plan.md`)
- **Scope**: Implementation planning and task breakdown
- **Responsibilities**: Convert features to actionable development plans
- **Boundaries**: Planning only, no code implementation

### Review Agent (`prompts/review.md`)
- **Scope**: Code quality and architectural review
- **Responsibilities**: Ensure code standards, patterns, and best practices
- **Boundaries**: Review only, no implementation changes

### Implementation Agent (`prompts/implement.md`)
- **Scope**: Three-phase implementation with user approval between phases
- **Responsibilities**: Interface design, test creation, and code implementation
- **Boundaries**: Must execute all three phases in order, requires user approval between phases

## Using AI Agents

### Getting Started

1. **Choose the right agent** based on the task type
2. **Reference the agent's prompt** to understand scope and capabilities
3. **Provide context** using the reference files in `docs/development/`
4. **Use templates** from `templates/` for consistent outputs

### Best Practices

- **Stay in scope**: Each agent has defined boundaries
- **Use references**: Point to specific file paths and documentation
- **Follow templates**: Use provided templates for consistency
- **Document decisions**: Update documentation when making changes

## Common Workflows

#### Adding an API Endpoint
1. Update `api/openapi.yml`
2. Run `make generate`
3. Implement controller in `internal/controller/`
4. Write tests
5. Update documentation

#### Creating a Service
1. Define interface in `internal/<domain>/interfaces.go`
2. Implement directly in `internal/<domain>/` (single impl) or `internal/<domain>/<provider>/` (multiple impls)
3. Write unit tests
4. Wire up in `internal/app/app.go`

## Template Usage

### Creating New Documents

#### For Features
```bash
cp templates/feature.md ../../features/YYYY-MM-DD-feature-name.md
```

#### For Plans
```bash
cp templates/plan.md ../../plans/YYYY-MM-DD-plan-name.md
```

#### For AI Agents
```bash
cp templates/prompt.md prompts/new-agent.md
```

### Template Guidelines

1. **Replace all placeholders** marked with `[BRACKETS]`
2. **Remove template comments** after completion
3. **Follow validation checklists** included in templates
4. **Update cross-references** to ensure links work

## Integration with Development

### With Standard Go Project Layout
- All agent prompts reference the Standard Go Project Layout
- Templates follow project conventions and patterns
- Documentation stays synchronized with code structure

### With Interface Pattern
Follow the standard interface-implementation pattern:
- Interfaces in `internal/<domain>/interfaces.go`
- Single implementations directly in `internal/<domain>/`
- Multiple implementations in named subdirectories

### With OpenAPI Workflow
- API changes start with OpenAPI specification
- Code generation follows spec updates
- Controllers implement generated interfaces

## Quality Standards

### Documentation Quality
- **Clear scope definitions** for each agent
- **Specific boundaries** and limitations
- **Actionable templates** with validation checklists
- **Consistent formatting** and structure

### Technical Accuracy
- **Up-to-date references** to current codebase structure
- **Correct Go patterns** and conventions
- **Valid file paths** and cross-references
- **Working examples** and code snippets

### Maintainability
- **Version controlled** templates and prompts
- **Automated validation** where possible
- **Clear change processes** for updates
- **Regular review** for accuracy and relevance

## Error Handling

### Common Issues
- **Scope violations**: Agents working outside defined boundaries
- **Outdated references**: File paths or patterns that have changed
- **Template inconsistencies**: Variations from standard formats
- **Missing context**: Insufficient project context for decisions

### Resolution Process
1. **Check agent scope** in the relevant prompt file
2. **Verify file paths** and references are current
3. **Use escalation criteria** for complex decisions
4. **Update templates** if patterns have changed

## Maintenance

### Regular Updates
- **Review agent boundaries** quarterly for relevance
- **Update file references** when project structure changes
- **Validate templates** against current project standards
- **Test workflows** with sample tasks

### Version Management
- **Track changes** to prompts and templates
- **Document reasoning** for major updates
- **Maintain backwards compatibility** where possible
- **Coordinate with project changes**

---

*This AI documentation system enables consistent, high-quality interactions between AI agents and the Go Starter codebase while maintaining clear boundaries and responsibilities.*