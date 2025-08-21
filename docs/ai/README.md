# AI Documentation

This directory contains AI agent documentation, prompts, and templates designed to enable effective AI-human collaboration on the Go Starter project.

## Purpose

Provide structured guidance for AI agents working with this codebase, ensuring consistent, high-quality contributions that follow project patterns and standards.

## Directory Structure

### 📁 prompts/
AI agent role definitions with specific scopes and responsibilities.

- **[ai.md](prompts/ai.md)** - AI documentation specialist for maintaining docs
- **[feature.md](prompts/feature.md)** - Product owner agent for gathering requirements
- **[plan.md](prompts/plan.md)** - Senior engineer for implementation planning
- **[review.md](prompts/review.md)** - Code reviewer for quality assurance

### 📁 templates/
Reusable templates for creating consistent documentation.

- **[README.md](templates/README.md)** - Guide for using templates
- **[feature.md](templates/feature.md)** - Feature specification template
- **[plan.md](templates/plan.md)** - Implementation plan template
- **[prompt.md](templates/prompt.md)** - AI agent prompt template

### 📄 Agent Coordination
Documentation for effective AI agent collaboration.

- **[agent-responsibility-matrix.md](agent-responsibility-matrix.md)** - Agent roles, boundaries, and coordination patterns
- **[escalation-criteria.md](escalation-criteria.md)** - When and how agents should hand off tasks

## Working with AI Agents

### Quick Reference for AI Agents

#### Key File Paths
- **Project overview**: `docs/development/project-context.md`
- **Go patterns**: `docs/development/foundations.md`
- **Coding standards**: `docs/guidelines/go-standards.md`
- **Interface pattern**: `docs/guidelines/interface-pattern.md`
- **OpenAPI spec**: `api/openapi.yml`
- **Main entry**: `main.go` → `cmd/serve.go`

#### Essential Commands
```bash
make generate  # Generate API code from OpenAPI spec
make fmt       # Format code
make lint      # Check code quality
make test      # Run tests
make build     # Build binaries
```

#### Development Workflow
1. Read feature from `docs/features/`
2. Apply standards from `docs/guidelines/`
3. Use templates from `docs/ai/templates/`
4. Reference context in `docs/development/`

### Agent Capabilities and Scopes

#### Documentation Agent (`ai.md`)
- **Scope**: AI documentation in `docs/ai/` directory
- **Tools**: Read, Write, Edit, Glob, LS, Grep
- **Responsibilities**: Create and maintain AI-facing documentation

#### Feature Agent (`feature.md`)
- **Scope**: Feature requirements in `docs/features/`
- **Tools**: Read, Write, Bash
- **Responsibilities**: Gather requirements, create user stories

#### Planning Agent (`plan.md`)
- **Scope**: Implementation plans in `docs/plans/`
- **Tools**: Read, Bash
- **Responsibilities**: Convert features to technical tasks (no coding)

#### Review Agent (`review.md`)
- **Scope**: Code review and quality checks
- **Tools**: Read, Bash
- **Responsibilities**: Analyze code changes, provide feedback

## Best Practices for AI Collaboration

### Understanding the Codebase
1. **Start with context**: Always read `docs/development/project-context.md` first
2. **Follow patterns**: Study `docs/guidelines/interface-pattern.md` for package structure
3. **Check existing code**: Use similar implementations as reference
4. **Respect conventions**: Follow naming and style in `docs/guidelines/go-standards.md`

### Making Changes
1. **OpenAPI first**: API changes start in `api/openapi.yml`
2. **Generate code**: Run `make generate` after API changes
3. **Test-driven**: Write tests before implementation
4. **Quality checks**: Always run `make fmt lint test`
5. **Document changes**: Update relevant documentation

### Code Organization Pattern
```
internal/<domain>/
├── interfaces.go     # Domain interfaces
├── types.go         # Shared types
├── errors.go        # Domain errors
└── v1/              # Implementation version
    ├── service.go   # Business logic
    └── service_test.go
```

### Common Tasks

#### Adding an API Endpoint
1. Update `api/openapi.yml`
2. Run `make generate`
3. Implement controller in `internal/controller/`
4. Write tests
5. Update documentation

#### Creating a Service
1. Define interface in `internal/<domain>/interfaces.go`
2. Implement in `internal/<domain>/v1/`
3. Write unit tests
4. Wire up in `internal/app/app.go`

## Template Usage

### Creating New Documents

#### For Features
```bash
cp templates/feature.md ../../features/YYYY-MM-DD-feature-name.md
# Fill in all [PLACEHOLDER] values
```

#### For Plans
```bash
cp templates/plan.md ../../plans/YYYY-MM-DD-feature-plan.md
# Link to source feature and define tasks
```

#### For AI Agents
```bash
cp templates/prompt.md prompts/new-agent.md
# Define scope and responsibilities
```

### Template Guidelines
1. Replace all `[PLACEHOLDER]` text
2. Remove template comments
3. Ensure all sections are complete
4. Cross-reference related documents
5. Follow existing patterns

## Agent Coordination

### Responsibility Matrix
See [agent-responsibility-matrix.md](agent-responsibility-matrix.md) for detailed agent roles and boundaries:
- **AI Documentation Agent**: `docs/ai/` directory only
- **Feature Agent**: Requirements in `docs/features/`  
- **Plan Agent**: Implementation plans in `docs/plans/` (READ-ONLY elsewhere)
- **Review Agent**: Code analysis and feedback (READ-ONLY)
- **General Purpose Agent**: Full implementation capabilities

### Escalation Guidelines
See [escalation-criteria.md](escalation-criteria.md) for detailed handoff procedures:

**Immediate Escalation Triggers:**
- Task requires tools agent doesn't have
- Work falls outside defined scope
- Quality standards cannot be met

**Planned Handoffs:**
- Feature complete → Plan Agent
- Plan complete → General Purpose Agent
- Implementation complete → Review Agent

### Communication Patterns
- Use specific file paths when referencing code
- Quote exact error messages
- Provide complete context for handoffs
- Link to relevant documentation
- Follow escalation procedures

### Quality Standards
All AI contributions must:
- Follow Go conventions (`gofmt`, `golint`)
- Include appropriate tests
- Update documentation
- Handle errors properly
- Use existing patterns
- Stay within defined agent scope

## Prompt Engineering

### Effective Prompts Include
1. **Clear scope** - What the agent can and cannot do
2. **Specific tools** - Which tools are available
3. **Quality standards** - Expected output quality
4. **Common tasks** - Examples of typical work
5. **Error handling** - How to handle edge cases

### Scope Boundaries
- **READ-ONLY**: Agents that analyze but don't modify
- **PLANNING ONLY**: Agents that plan but don't implement
- **IMPLEMENTATION**: Agents that write code
- **REVIEW**: Agents that assess quality

## Integration with Development Workflow

### Pre-Implementation
1. Feature agent gathers requirements
2. Planning agent creates technical approach
3. Review agent validates plan

### During Implementation
1. Development agent writes code
2. Follow TDD approach
3. Regular quality checks

### Post-Implementation
1. Review agent analyzes changes
2. Documentation updates
3. Quality validation

## Troubleshooting AI Issues

### Common Problems

#### Scope Confusion
- **Issue**: Agent tries to work outside its area
- **Solution**: Check prompt scope definition
- **Prevention**: Clear boundaries in prompt

#### Pattern Violations
- **Issue**: Code doesn't follow project patterns
- **Solution**: Reference `docs/guidelines/`
- **Prevention**: Include pattern examples in context

#### Missing Context
- **Issue**: Agent lacks necessary information
- **Solution**: Provide relevant file paths
- **Prevention**: Include key references upfront

## Implemented Features

AI collaboration framework includes:
- ✅ **Agent Responsibility Matrix** - Clear role definitions and boundaries
- ✅ **Escalation Criteria** - Detailed handoff procedures and triggers
- ✅ **Template Validation** - Comprehensive checklists for quality
- ✅ **Scope Enforcement** - Clear boundaries and handoff points
- ✅ **Workflow Integration** - Seamless development process integration

## Future Enhancements

Planned improvements for AI collaboration:
- Automated scope validation tools
- Enhanced error recovery patterns
- Performance optimization guides
- Security review automation
- Agent coordination metrics

## Related Documentation

- [../development/](../development/) - Core development docs
- [../guidelines/](../guidelines/) - Coding standards
- [../architecture/](../architecture/) - System design
- [../features/](../features/) - Feature specifications
- [../plans/](../plans/) - Implementation plans

---

*This directory enables effective AI-human collaboration by providing clear guidance, templates, and patterns for AI agents working with the Go Starter codebase.*