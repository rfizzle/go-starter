# Implementation Plans

This directory contains detailed implementation plans created from feature specifications. Each plan breaks down features into specific, actionable development tasks with Go-specific technical approaches.

## Overview

Implementation plans bridge the gap between high-level feature requirements and actual development work. They provide:

- **Task Breakdown**: Features decomposed into manageable development tasks
- **Technical Approach**: Go-specific implementation strategies and patterns
- **Risk Assessment**: Potential issues and mitigation strategies  
- **Verification Strategy**: Acceptance criteria and risks mapped to effective checks
- **Definition of Done**: Clear completion criteria

## Plan Structure

All implementation plans follow the template in `docs/ai/templates/plan.md` and include:

### Essential Components
1. **Feature Summary** - Problem statement, solution approach, users, success criteria
2. **Technical Scope** - Packages, APIs, database changes, dependencies
3. **Task Breakdown** - Phased approach with specific, actionable tasks
4. **Go Implementation** - Package structure following standard patterns
5. **Verification Strategy** - The narrowest stable check for each requirement or risk
6. **Risk Assessment** - Potential issues with mitigation plans

### Metadata
- **Source**: Links to feature document
- **Date**: Plan creation date
- **Effort**: Story points or time estimate
- **Priority**: High, Medium, Low
- **Status**: Planning, Ready, In Progress, Complete

Refer to the template for detailed structure requirements and validation guidelines.

## Creating Implementation Plans

### Prerequisites
1. **Complete feature document** in `docs/features/`
2. **Technical requirements** clearly defined
3. **Acceptance criteria** specified
4. **Dependencies** identified

### Process
1. **Use the planning agent** defined in `docs/ai/prompts/planning.md`
2. **Follow the template** in `docs/ai/templates/plan.md`
3. **Reference existing patterns** from the codebase
4. **Include validation checklist** completion

### Naming Convention
```
YYYY-MM-DD-feature-name-plan.md
```

Examples:
- `2024-01-15-user-authentication-plan.md`
- `2024-02-03-api-rate-limiting-plan.md`

## Go-Specific Considerations

### Package Organization
Follow contract-first package boundaries:
- **Contracts**: Document purpose, non-goals, public behavior, bounds, and ownership
- **Interfaces**: Define small interfaces only at real substitution or consumer seams
- **Single Implementation**: Keep the natural implementation in the contract package
- **Multiple Implementations**: Place real variants in named subpackages
- **Controllers**: Keep generated transport mapping in `internal/controller/`

### Implementation Patterns
- **Error Handling**: Domain-specific errors with proper wrapping
- **Context Usage**: Cancellation and timeout support
- **Testing**: Use table-driven tests only when shared structure makes them clearer
- **Documentation**: Godoc comments on all exported items

### Quality Standards
- **Go Conventions**: `gofmt`, `golangci-lint`, `go vet` compliance
- **Test Value**: Every test protects a contract, regression, boundary, or named risk
- **Interface Design**: Small, focused interfaces
- **Dependency Injection**: Constructors accept narrow interfaces only at real seams

## Plan Lifecycle

### Status Progression
1. **Planning** - Initial plan creation and review
2. **Ready** - Plan approved and ready for development
3. **In Progress** - Active development underway
4. **Complete** - All tasks completed and accepted

### Review Process
1. **Technical Review** - Architecture and approach validation
2. **Effort Review** - Task sizing and timeline validation  
3. **Risk Review** - Risk assessment and mitigation strategies
4. **Approval** - Final approval for development

## Integration with Development

### With Feature Documents
- Plans directly reference source feature documents
- Feature acceptance criteria become plan validation criteria
- Feature requirements drive technical task breakdown

### With Code Implementation
- Plans guide development agent work
- Task completion tracked against plan checklist
- Code reviews verify plan adherence

### With Testing Strategy
- Plans map each acceptance criterion and significant risk to effective verification
- Testing approaches use the narrowest stable boundary that can detect the failure
- Coverage is a diagnostic signal, not a completion target

## Common Plan Templates

### API Feature Plan
- OpenAPI specification updates
- Controller implementation tasks
- Integration test requirements
- Database migration considerations

### Service Feature Plan
- Interface design decisions
- Implementation task breakdown
- Unit testing strategies
- Dependency injection setup

### Infrastructure Plan
- Configuration management updates
- Deployment requirement changes
- Monitoring and observability setup
- Performance testing criteria

## Quality Assurance

### Plan Validation
- [ ] All template sections completed
- [ ] Tasks are specific, necessary, and independently verifiable
- [ ] Go implementation patterns correct
- [ ] Material risks and bounded mitigations identified
- [ ] Verification maps to every acceptance criterion and significant risk

### Technical Review Criteria
- [ ] Architecture decisions sound
- [ ] Package structure follows project patterns
- [ ] Integration points clearly defined
- [ ] Performance implications considered
- [ ] Security requirements addressed

---

*Implementation plans are living documents that guide development while maintaining flexibility for technical decisions during implementation.*
