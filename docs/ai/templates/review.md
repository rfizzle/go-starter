# Code Review Template

<!-- Replace [PLACEHOLDER] text and remove these comments -->

## Summary
[BRIEF_OVERVIEW_OF_CHANGES]

## Critical Issues (🚨)
<!-- Security vulnerabilities, potential bugs or crashes, breaking changes -->
- [CRITICAL_ISSUE_1]
- [CRITICAL_ISSUE_2]

## Major Issues (⚠️)
<!-- Performance problems, architectural concerns, missing tests -->
- [MAJOR_ISSUE_1]
- [MAJOR_ISSUE_2]

## Minor Issues (💡)
<!-- Style improvements, documentation gaps, optimization opportunities -->
- [MINOR_ISSUE_1]
- [MINOR_ISSUE_2]

## Positive Notes (✅)
<!-- Good patterns used, proper implementation, well-tested code -->
- [POSITIVE_NOTE_1]
- [POSITIVE_NOTE_2]

## Recommendations
<!-- Specific actionable next steps -->
1. [RECOMMENDATION_1]
2. [RECOMMENDATION_2]
3. [RECOMMENDATION_3]

## Review Criteria Checklist

### Code Quality
- [ ] **Correctness**: Logic errors, edge cases, error handling
- [ ] **Scope**: Changes map to approved acceptance criteria without adjacent features
- [ ] **Simplicity**: Design is the least complex one that satisfies the contract
- [ ] **Bounds**: Inputs, resources, concurrency, retries, and failure behavior are bounded
- [ ] **Performance**: Evidence supports any performance mechanism or optimization
- [ ] **Concurrency**: Race conditions, goroutine leaks, proper channel usage
- [ ] **Security**: Input validation, SQL injection, secrets exposure
- [ ] **Style**: gofmt compliance, naming conventions, code organization

### Go-Specific Issues
- [ ] **Package Contract**: Purpose, non-goals, public API, bounds, and ownership are clear
- [ ] **Interface Design**: Interfaces are small, justified, and consumer-owned where possible
- [ ] **Implementation Shape**: One implementation stays local; real variants use named subpackages
- [ ] **Error Handling**: Proper error wrapping, context preservation
- [ ] **Context Usage**: Context as first parameter, proper cancellation handling
- [ ] **Package Design**: Clear responsibilities, minimal public surface
- [ ] **Documentation**: Package documentation and README are present when required
- [ ] **Testing**: Each test protects a named contract, regression, boundary, or risk

### Project Standards
- [ ] **Architecture**: Standard Go Project Layout compliance
- [ ] **Dependencies**: New dependencies and layers have current concrete justification
- [ ] **Documentation**: Godoc comments, README updates, architectural decisions
- [ ] **Configuration**: Proper config management, environment handling
- [ ] **Build**: Makefile targets, CI/CD compatibility

## TEMPLATE VALIDATION CHECKLIST

Before finalizing this code review, verify:

### Content Completeness
- [ ] All `[PLACEHOLDER]` values replaced with actual findings
- [ ] Template comments (`<!-- -->`) removed
- [ ] Summary provides clear overview of changes
- [ ] Issues are categorized appropriately (Critical/Major/Minor)
- [ ] Positive notes acknowledge good practices

### Review Quality
- [ ] Issues reference specific lines and provide examples
- [ ] Suggestions are constructive and include solutions
- [ ] Explanations include the "why" behind recommendations
- [ ] Language is respectful and focuses on code, not person
- [ ] Issues are prioritized correctly (must-fix vs. nice-to-have)

### Technical Accuracy
- [ ] Go patterns and conventions are correctly applied
- [ ] Security and performance concerns are valid
- [ ] Architectural feedback aligns with project standards
- [ ] Test feedback identifies the contract or realistic defect being protected
- [ ] Review does not request tests solely to raise coverage or test count
- [ ] Speculative abstractions, configurability, and resilience mechanisms are rejected
- [ ] Dependencies and build considerations are accurate

### Review Effectiveness
- [ ] Critical issues would block deployment
- [ ] Major issues significantly impact code quality
- [ ] Minor issues provide meaningful in-scope improvements
- [ ] Recommendations are specific and actionable
- [ ] Recommendations do not expand the approved feature scope
- [ ] Review criteria checklist is completed accurately
