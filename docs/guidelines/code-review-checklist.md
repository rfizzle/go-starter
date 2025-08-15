# Code Review Checklist

This checklist provides guidelines for conducting thorough and consistent code reviews across the project.

---

## Review Process

### Before Starting Review
- [ ] **Pull Request Description** - Clear description of changes and rationale
- [ ] **Linked Issues** - Associated tickets or requirements referenced
- [ ] **Tests Included** - New tests for new functionality, updated tests for changes
- [ ] **CI Status** - All automated checks (tests, linting) are passing
- [ ] **Size Appropriate** - PR is reasonably sized for effective review (< 500 lines preferred)

### Review Categories
Focus your review on these key areas in order of priority:

1. **Correctness** - Does the code work as intended?
2. **Security** - Are there any security vulnerabilities?
3. **Architecture** - Does it fit well with existing design?
4. **Testing** - Are there appropriate tests?
5. **Performance** - Any performance implications?
6. **Style** - Does it follow coding standards?

---

## Correctness Review

### Logic and Behavior
- [ ] **Business Logic** - Implementation matches requirements
- [ ] **Edge Cases** - Handles boundary conditions appropriately
- [ ] **Error Paths** - All error conditions are handled properly
- [ ] **Data Validation** - Input validation is comprehensive
- [ ] **State Management** - Correct state transitions and consistency

### Go-Specific Correctness
- [ ] **Error Handling** - All errors are properly handled or propagated
- [ ] **Context Usage** - Context is threaded through appropriately
- [ ] **Resource Management** - Files, connections, goroutines are properly cleaned up
- [ ] **Nil Checks** - Appropriate nil checks before dereferencing pointers
- [ ] **Type Safety** - Correct type usage and conversions

```go
// Good: Proper error handling
result, err := processData(input)
if err != nil {
    return fmt.Errorf("processing data: %w", err)
}

// Bad: Ignoring errors
result, _ := processData(input)
```

---

## Security Review

### Security Fundamentals
- [ ] **Input Validation** - All user inputs are validated and sanitized
- [ ] **SQL Injection** - Using parameterized queries or ORMs properly
- [ ] **Authentication** - Proper authentication checks in place
- [ ] **Authorization** - Access control enforced at appropriate levels
- [ ] **Secrets Management** - No hardcoded secrets, proper secret handling

### Go Security Patterns
- [ ] **Context Timeout** - Appropriate timeouts to prevent resource exhaustion
- [ ] **Rate Limiting** - Protection against abuse where applicable
- [ ] **Safe Concurrency** - No race conditions or data races
- [ ] **Memory Safety** - No buffer overflows or unsafe operations
- [ ] **Dependency Security** - Only trusted, maintained dependencies

```go
// Good: Parameterized query
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = $1", userID)

// Bad: String concatenation (SQL injection risk)
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = " + userID)
```

---

## Architecture Review

### Design Patterns
- [ ] **Interface Compliance** - Follows interface-implementation pattern
- [ ] **Package Structure** - Appropriate package organization
- [ ] **Dependency Direction** - Dependencies point in correct direction
- [ ] **Separation of Concerns** - Clear responsibility boundaries
- [ ] **SOLID Principles** - Adherence to SOLID design principles

### Go Architecture
- [ ] **Package Visibility** - Appropriate use of exported vs unexported items
- [ ] **Interface Design** - Minimal, focused interfaces
- [ ] **Constructor Pattern** - Proper use of New* constructors
- [ ] **Error Types** - Appropriate error types and sentinel errors
- [ ] **Configuration** - Proper configuration injection and management

```go
// Good: Interface in domain package, implementation in versioned package
// internal/user/interfaces.go
type Repository interface {
    Create(ctx context.Context, user User) error
}

// internal/user/v1/repository.go
type repository struct { ... }
```

---

## Testing Review

### Test Coverage
- [ ] **Unit Tests** - New functions have corresponding unit tests
- [ ] **Edge Cases** - Tests cover boundary conditions and error cases
- [ ] **Integration Tests** - Complex interactions are integration tested
- [ ] **Test Quality** - Tests are clear, maintainable, and deterministic
- [ ] **Coverage Metrics** - Adequate test coverage (≥80% for new code)

### Test Patterns
- [ ] **Table-Driven Tests** - Multiple scenarios tested efficiently
- [ ] **Interface Testing** - Tests against interfaces when possible
- [ ] **Mock Usage** - Appropriate use of mocks vs real implementations
- [ ] **Test Helpers** - Common test setup extracted to helpers
- [ ] **Test Data** - Use of factories or fixtures for test data

```go
// Good: Table-driven test with clear scenarios
func TestCalculate(t *testing.T) {
    tests := []struct {
        name    string
        input   int
        want    int
        wantErr bool
    }{
        {"positive number", 5, 25, false},
        {"zero", 0, 0, false},
        {"negative number", -1, 0, true},
    }
    // ... test implementation
}
```

---

## Performance Review

### Performance Considerations
- [ ] **Algorithm Efficiency** - Appropriate time/space complexity
- [ ] **Memory Allocations** - Minimized unnecessary allocations
- [ ] **Goroutine Management** - Proper goroutine lifecycle management
- [ ] **Database Queries** - Efficient queries without N+1 problems
- [ ] **Caching Strategy** - Appropriate caching where beneficial

### Go Performance Patterns
- [ ] **Slice Pre-allocation** - Pre-allocate slices when size is known
- [ ] **String Building** - Use strings.Builder for concatenation
- [ ] **Object Pooling** - Consider sync.Pool for frequently allocated objects
- [ ] **Context Cancellation** - Proper context usage for cancellation
- [ ] **Profiling Readiness** - Code is ready for profiling if needed

```go
// Good: Pre-allocated slice
items := make([]Item, 0, expectedCount)

// Good: String builder for concatenation
var builder strings.Builder
for _, part := range parts {
    builder.WriteString(part)
}
```

---

## Style and Standards Review

### Code Style
- [ ] **Naming Conventions** - Clear, descriptive names following Go conventions
- [ ] **Function Size** - Functions are appropriately sized and focused
- [ ] **Code Organization** - Logical organization within files and packages
- [ ] **Comments** - Appropriate godoc comments on exported items
- [ ] **Formatting** - Code is properly formatted (gofmt)

### Go Style Specifics
- [ ] **Package Names** - Short, lowercase, descriptive package names
- [ ] **Receiver Names** - Consistent, short receiver names
- [ ] **Error Messages** - Lowercase error messages without punctuation
- [ ] **Constants** - Appropriate use of const vs var
- [ ] **Imports** - Standard library first, then third-party, then local

```go
// Good: Clear naming and structure
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    if err := s.validateCreateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    // ... implementation
}
```

---

## Documentation Review

### Code Documentation
- [ ] **Godoc Comments** - All exported items have appropriate documentation
- [ ] **Package Documentation** - Package purpose is clearly documented
- [ ] **Complex Logic** - Non-obvious code has explanatory comments
- [ ] **Examples** - Code examples provided where helpful
- [ ] **API Documentation** - Public APIs are well documented

### Project Documentation
- [ ] **Architecture Changes** - Significant changes documented in architecture docs
- [ ] **Breaking Changes** - Breaking changes clearly documented
- [ ] **Migration Guides** - Instructions for adapting to changes
- [ ] **README Updates** - Project README updated if necessary

---

## Review Feedback Guidelines

### Providing Feedback
- **Be Constructive** - Focus on the code, not the person
- **Be Specific** - Provide clear, actionable suggestions
- **Explain Why** - Include rationale for suggested changes
- **Categorize Importance** - Distinguish between critical issues and suggestions
- **Suggest Solutions** - Offer alternatives when pointing out problems

### Feedback Categories
- **🚨 Critical** - Security issues, bugs, breaking changes
- **⚠️ Major** - Architecture concerns, performance issues, missing tests
- **💡 Minor** - Style improvements, documentation, suggestions
- **✅ Positive** - Good practices worth highlighting

### Example Feedback
```
💡 Minor: Consider using strings.Builder here for better performance

Current:
var result string
for _, item := range items {
    result += item.String()
}

Suggested:
var builder strings.Builder
for _, item := range items {
    builder.WriteString(item.String())
}
result := builder.String()
```

---

## Common Issues Checklist

### Frequent Problems
- [ ] **Goroutine Leaks** - Goroutines have clear exit conditions
- [ ] **Resource Leaks** - Files, connections properly closed
- [ ] **Error Handling** - Errors not ignored with `_`
- [ ] **Context Propagation** - Context passed through call chains
- [ ] **Race Conditions** - Shared state properly synchronized
- [ ] **Nil Pointer Panics** - Appropriate nil checks
- [ ] **Infinite Loops** - Loops have proper termination conditions

### Anti-Patterns
- [ ] **Premature Optimization** - Avoid optimizing without profiling
- [ ] **God Objects** - Large structs with too many responsibilities
- [ ] **Deep Nesting** - Excessive indentation levels
- [ ] **Long Parameter Lists** - Functions with too many parameters
- [ ] **Global State** - Avoiding global variables where possible

---

## Post-Review Actions

### For Reviewers
- [ ] **Mark Approval** - Explicitly approve when ready
- [ ] **Follow Up** - Check that feedback is addressed
- [ ] **Learn and Share** - Note patterns for future reviews

### For Authors
- [ ] **Address Feedback** - Respond to all review comments
- [ ] **Update Tests** - Add tests for any changes made during review
- [ ] **Documentation** - Update documentation if changes affect APIs
- [ ] **Verify CI** - Ensure all checks pass after changes

---

*This checklist should be used as a guide, not a rigid requirement. Use judgment to focus on the most important aspects for each specific change.*