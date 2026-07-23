# AI Agent Definitions and Standards

This document defines core concepts, terminology, and decision criteria used across all AI agents in this project.

## Core Concepts

### Package Stages

**Definition**: A package stage represents a complete, testable, and deployable unit of functionality that can be independently validated.

**Completion Criteria**:
- All code compiles without errors
- All tests pass (unit and integration)
- Linting and formatting standards met
- Documentation updated
- OpenAPI spec updated if API changes made
- Generated code updated via `make generate`

**Stage Boundaries**:
- **Single Feature Package**: One complete feature with all its components (API, business logic, storage, tests)
- **Single Bug Fix Package**: Complete resolution of a specific bug including tests and documentation
- **Single Refactor Package**: Complete refactoring of a specific component or pattern

### Agent Mode Selection

#### Plan Mode vs Edit Mode - Definitive Criteria

**Plan Mode (Opus) - Default Choice**
Use Plan Mode as the default for any non-trivial task. Plan Mode should be used when:

**Always Use Plan Mode For**:
- **All new features** (documentation → planning → implementation)
- **Complex modifications** (architecture, refactoring, system changes)
- **Unfamiliar territory** (new domains, patterns, or technologies)
- **High-risk changes** (security, performance, breaking changes)
- **Multi-step processes** (anything requiring multiple actions)
- **When in doubt** (better to plan than assume)

**Specific Plan Mode Scenarios**:
- User request involves multiple files or components
- Implementation requires architectural decisions
- Dependencies between changes need coordination
- Risk of breaking existing functionality is high
- User explicitly asks for a plan or strategy
- Complex bug analysis required

**Edit Mode (Sonnet) - Limited Use**
Use Edit Mode only for clearly simple tasks:

**Edit Mode Only For**:
- **Simple typo fixes** (single character or word changes)
- **Version number updates** (clear, mechanical changes)
- **Following approved plans** (when plan mode was already used)
- **Trivial updates** (adding a single line of documentation)
- **Configuration tweaks** (known, low-risk changes)

**Specific Edit Mode Scenarios**:
- User explicitly requests edit mode ("just fix this typo")
- Following an existing detailed plan or specification
- Task is genuinely trivial (< 5 minutes, single file, no decisions)
- Emergency hot fixes (but plan mode still preferred)

**Decision Framework**:
```
Is this a trivial, mechanical change with zero decisions required?
├─ YES → Edit Mode (rare)
└─ NO → Plan Mode (default)

Is the user explicitly requesting immediate execution?
├─ YES → Edit Mode (if truly simple)
└─ NO → Plan Mode (default)

When unsure → Always use Plan Mode
```

**Plan → Review → Edit Pattern**:
```
Step a: Agent (Plan Mode with Opus) → Present strategy and approach
         ↓
Step b: User Review → Provide feedback and approval  
         ↓
Step c: Agent (Edit Mode with Sonnet) → Execute approved plan
```

**Why Plan Mode First**:
- **See before you leap**: Understand what will be done before execution
- **Feedback opportunity**: Refine approach based on user input
- **Prevent rework**: Avoid misunderstood requirements
- **Strategic thinking**: Leverage Opus for comprehensive analysis
- **Quality assurance**: Better outcomes through deliberate planning

### Template Validation Levels

#### Level 1: Existence Check
- Verify template file exists at expected path
- Suitable for optional templates or fallback scenarios

#### Level 2: Content Structure Check
- Verify template contains required sections/placeholders
- Validate template format and syntax
- Use for critical templates that must be properly structured

#### Level 3: Content Completeness Check
- Verify all required placeholders are present
- Validate template follows current patterns
- Cross-reference with actual usage patterns
- Use for templates that directly affect output quality

## Agent Terminology

### Standard Terms

- **Implementation Plan**: Detailed breakdown of tasks needed to complete a feature or fix
- **Package Review**: Review of all changes within a single package stage
- **Feature Specification**: Formal document describing feature requirements and implementation approach
- **Quality Gate**: Checkpoint where code must meet all quality standards before proceeding
- **Handoff Criteria**: Conditions that must be met before passing work to another agent

### Workflow States

- **Planning**: Analyzing requirements and creating implementation strategy
- **Implementing**: Actively writing code and making changes
- **Reviewing**: Analyzing code quality and providing feedback
- **Documenting**: Creating or updating documentation
- **Validating**: Running tests and quality checks

## Error Handling Standards

### Template Missing
```
⚠️ Template not found at [path]. Using fallback approach:
[description of fallback behavior]
```

### Template Malformed
```
⚠️ Template at [path] is malformed. Issues found:
- [specific issues]
Using minimal template structure instead.
```

### Mode Selection Conflict
```
🤔 This task could use either Plan or Edit mode. 
Based on [criteria], I'm choosing [mode]. 
Would you prefer the other approach?
```

## Quality Standards Reference

### Code Quality Gates
1. **Compilation**: Code must compile without errors
2. **Testing**: All tests must pass (`make test`)
3. **Linting**: No linting errors (`make lint`)
4. **Formatting**: Code properly formatted (`make fmt`)
5. **Generation**: Generated code up to date (`make generate`)

### Documentation Requirements
- Public APIs must have godoc comments
- Complex functions require implementation comments
- README updates for user-facing changes
- OpenAPI spec updates for API changes

### Review Criteria Priority
1. **Correctness**: Code works as intended
2. **Security**: No security vulnerabilities introduced
3. **Performance**: No significant performance regressions
4. **Maintainability**: Code follows project patterns and is readable
5. **Testing**: Tests protect the changed contracts, regressions, boundaries, and risks

## Inter-Agent Communication

### Handoff Requirements
When one agent hands work to another, include:
- Summary of completed work
- Current state of the codebase
- Outstanding issues or concerns
- Next steps or recommendations

### Status Communication Format
```
📋 Status: [current activity]
📦 Package Stage: [stage description]
✅ Completed: [list of completed tasks]
⏳ In Progress: [current task]
📋 Next: [upcoming tasks]
```

This standardized vocabulary ensures consistent communication and behavior across all AI agents.
