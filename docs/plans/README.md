# Implementation Plans

This directory contains detailed implementation plans created from feature specifications. Plans break down features into actionable tasks with technical approaches and risk assessments.

## Purpose

Transform feature requirements into concrete implementation roadmaps that development teams can execute, ensuring systematic delivery and risk management.

## How Implementation Planning Works

### Planning Agent

The project uses an AI-powered implementation planning agent configured via `docs/ai/prompts/plan.md` that:

- Analyzes feature documents from `docs/features/`
- Converts features into specific, actionable development tasks
- Defines Go-specific implementation approaches and patterns
- Creates comprehensive testing plans (unit, integration, e2e)
- Identifies technical and business risks with mitigation strategies

### Planning Process

1. **Feature Analysis**: Review feature document thoroughly
2. **Technical Design**: Define package structure, interfaces, API changes, database modifications
3. **Task Decomposition**: Break into manageable tasks (4-8 hours each)
4. **Risk Assessment**: Identify technical risks, integration challenges, mitigation strategies
5. **Review and Approval**: Technical lead review, team estimation, stakeholder approval

## Plan Document Structure

All implementation plans follow the template in `docs/ai/templates/plan.md` and include:

### Essential Components
1. **Feature Summary** - Problem statement, solution approach, users, success criteria
2. **Technical Scope** - Packages, APIs, database changes, dependencies
3. **Task Breakdown** - Phased approach with specific, actionable tasks
4. **Go Implementation** - Package structure following `internal/[domain]/v1/` patterns
5. **Test Strategy** - Unit, integration, and E2E testing requirements
6. **Risk Assessment** - Potential issues with mitigation plans

### Metadata
- **Source**: Links to feature document
- **Date**: Plan creation date
- **Effort**: Story points or time estimate
- **Priority**: High, Medium, Low
- **Status**: Planning, Ready, In Progress, Complete

Refer to the template for detailed structure requirements and validation guidelines.

## Creating a New Plan

```bash
# Copy the template
cp docs/ai/templates/plan.md docs/plans/YYYY-MM-DD-feature-name-plan.md

# Link to source feature
# Update all [PLACEHOLDER] values
# Define specific tasks
# Add risk assessment
```

## Task Breakdown Guidelines

### Phase Structure
Plans typically follow these phases:

#### Phase 1: Foundation
- Set up package structure
- Define interfaces and types
- Create database migrations
- Set up test infrastructure

#### Phase 2: Implementation
- Core business logic
- API endpoints
- Database operations
- Error handling

#### Phase 3: Integration
- Wire up dependencies
- Integration testing
- Performance testing
- Documentation updates

### Task Definition
Each task should specify:
- **Description** - What needs to be done
- **Owner** - Who is responsible
- **Effort** - Time estimate in hours
- **Dependencies** - What must complete first
- **Acceptance** - How we know it's done

## Implementation Patterns

### Following Project Standards
All plans must align with:
- [Interface-implementation pattern](../guidelines/interface-pattern.md)
- [Go coding standards](../guidelines/go-standards.md)
- [Testing strategy](../guidelines/testing-strategy.md)
- [OpenAPI-first development](../development/project-context.md)

### Code Organization
```
internal/
└── <feature>/
    ├── interfaces.go     # Feature interfaces
    ├── types.go         # Domain types
    ├── errors.go        # Feature-specific errors
    └── v1/
        ├── service.go   # Business logic
        └── service_test.go
```

### API Development Flow
1. Update `api/openapi.yml`
2. Run `make generate`
3. Implement controller
4. Add integration tests
5. Update documentation

## Risk Management

### Common Risks
- **Data Migration** - Backward compatibility, rollback plan
- **Performance** - Load testing, optimization strategy
- **Dependencies** - External service reliability
- **Security** - Authentication, authorization, data protection
- **Integration** - Breaking changes, version compatibility

### Mitigation Strategies
- Feature flags for gradual rollout
- Comprehensive testing at all levels
- Performance benchmarks and monitoring
- Rollback procedures documented
- Security review before deployment

## Quality Standards

### Planning Criteria
- **Actionable Tasks**: Every task must be specific, measurable, and assignable
- **Complete Coverage**: Plan covers all aspects from development through deployment
- **Go Idioms**: All technical approaches follow Go best practices and project patterns
- **Test-Driven**: Planning assumes a TDD approach with tests written first

### Definition of Done
Tasks are complete when:
- Code follows conventions
- Tests pass with good coverage
- Code reviewed and approved
- Documentation updated
- No linting errors

## File Organization

```
docs/plans/
├── README.md                          # This file - process documentation
├── YYYY-MM-DD-feature-name-plan.md    # Individual implementation plans
└── archive/                           # Completed plans (optional)
```

## Integration with Development

Implementation plans integrate with the overall development workflow:

- **Feature development**: Convert feature specs into actionable plans
- **Sprint planning**: Break plans into sprint-sized tasks
- **Risk management**: Proactively identify and mitigate risks
- **Quality assurance**: Ensure comprehensive test coverage and standards compliance

## Best Practices

### Effective Planning
1. **Right-size tasks** - Not too big, not too small (4-8 hours each)
2. **Clear ownership** - Every task has an owner
3. **Realistic estimates** - Include testing and review time
4. **Dependency awareness** - Identify critical path
5. **Risk mitigation** - Plan for what could go wrong

### Technical Excellence
- Follow TDD approach
- Maintain >85% test coverage
- Use existing patterns
- Consider performance early
- Plan for observability

## Related Documentation

- [../features/](../features/) - Source feature specifications
- [../ai/templates/plan.md](../ai/templates/plan.md) - Implementation plan template
- [../ai/prompts/plan.md](../ai/prompts/plan.md) - Planning agent configuration
- [../development/project-context.md](../development/project-context.md) - Development workflows
- [../guidelines/](../guidelines/) - Coding standards

---

*This directory ensures features are systematically implemented with clear roadmaps and risk management.*