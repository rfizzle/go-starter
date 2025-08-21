# Interface Package Pattern

> **Scope:** Go package architecture and interface design patterns  
> **Applies to:** All Go packages in `internal/` and `pkg/` directories  
> **Version:** 1.0.0

## Purpose
- Enforce consistent, predictable package structure across Go project
- Enable clean separation between contract definition and implementation details
- Support multiple implementations, versioning, and dependency inversion principles

## Requirements

### Core Principles
- **Interface Segregation**: Interfaces defined at top level of domain package, separate from implementation
- **Implementation Isolation**: All concrete implementations reside in versioned sub-packages (e.g., `v1/`, `v2/`)
- **Dependency Inversion**: Higher-level packages depend on interfaces, not concrete implementations
- **Single Responsibility**: Each interface defines cohesive set of related operations

### Specific Rules
1. Domain interfaces defined in `internal/<domain>/interfaces.go`
2. Supporting types and constants in `internal/<domain>/types.go`
3. Primary implementation in `internal/<domain>/v1/`
4. Constructor functions use pattern `New<InterfaceName>(options...)`
5. Implementation structs should not be exported unless necessary

## Implementation Guidelines

### For Development Teams
- **Interface First**: Always define or reference interface before implementing
- **Version Awareness**: Place new implementations in appropriate version sub-packages
- **Testing Strategy**: Write tests against interfaces, not concrete implementations
- **Documentation**: Document interface contracts thoroughly, including error conditions

### For Code Reviews
- **Contract Compliance**: Verify interface contracts followed and implementations properly isolated
- **Backward Compatibility**: Maintain backward compatibility or bump version appropriately
- **Integration Testing**: Ensure different implementations can be swapped without breaking consumers

### For Go Projects
- **Interface Composition**: Use interface embedding to compose larger interfaces from smaller ones
- **Type Assertions**: Minimize type assertions by designing interfaces appropriately
- **Error Handling**: Follow Go error handling conventions in all interface methods
- **Context Propagation**: Always thread context through interface methods for cancellation and timeouts

## Compliance

### Checklist
- [ ] Domain interfaces defined in `internal/<domain>/interfaces.go`
- [ ] Implementation in versioned subpackage (`internal/<domain>/v1/`)
- [ ] Constructor follows `New<InterfaceName>` pattern
- [ ] All methods accept `context.Context` as first parameter where appropriate
- [ ] Interface methods return explicit errors
- [ ] Implementation struct is unexported
- [ ] Tests written against interfaces, not concrete types
- [ ] Documentation covers interface contracts and error conditions

### Violations
- **Detection**: CI checks for package structure violations and interface compliance
- **Correction**: Refactor non-compliant code to follow interface-implementation pattern

## Exceptions
- **Simple Utility Packages**: Pure utility functions without state may not require interfaces
- **Generated Code**: API client code generated from OpenAPI specs may follow different patterns
- **Third-party Integrations**: Wrapper packages for external libraries may use adapted patterns

## Success Metrics
- **Interface Coverage**: >90% of public functionality exposed through interfaces—Target: 100%
- **Package Compliance**: Package structure follows pattern—Target: 100%
- **Test Coverage**: Interface-based tests vs implementation tests ratio—Target: 3:1

## Related Documentation
- [docs/development/foundations.md](../development/foundations.md) - Go interface design patterns and core principles
- [docs/development/project-context.md](../development/project-context.md) - Project architecture overview
- [docs/guidelines/go-standards.md](../guidelines/go-standards.md) - Go-specific coding standards