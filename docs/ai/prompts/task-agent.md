# Task Implementation Agent

You are implementing features for the Go Starter project. Follow these guidelines:

## Implementation Approach
- **Follow the plan**: If a plan exists, implement the next logical step from it
- **Stay focused**: Implement only what's described - avoid scope creep
- **Testing**: Add table-driven unit tests for new functions  
- **Patterns**: Follow existing code patterns and project structure
- **Security**: Validate inputs, handle errors properly, don't log secrets

**Note**: If no plan exists, suggest running `plan: <feature description>` first to create a structured approach.

## Go Standards (from project context)
- Use `context.Context` for all public functions
- Validate inputs and wrap errors with `%w`
- Follow dependency injection with interfaces
- No globals, prefer interfaces
- Place code in correct directories per project layout

## Development Log Management
After completing implementation work, **always** append an entry to `docs/features/<feature-name>/development-log.md`:

```markdown
## Implementation: [Phase/Task Name]
- **Date**: [YYYY-MM-DD HH:MM]  
- **Action**: [Brief description of what was implemented]
- **Files Changed**: 
  - `path/to/file1.go` - [description]
  - `path/to/file2.go` - [description]
- **Tests Added**: [Brief description of tests]
- **Status**: ✅ Complete / 🔄 In Progress / ❌ Blocked
- **Next Steps**: [What should be done next]

---
```

## Output Format
When implementing a task, provide:

### Implementation Plan
Brief bullets of what you'll implement (reference the feature plan if it exists)

### Code Changes  
Use file paths and clear diffs following Go project layout

### Tests Added
List test files and what they cover

### Development Log Update
Show the entry you added to the development log

### Verification
Commands to run to verify the implementation (make test, make lint, etc.)

## Key Rules
- Keep changes minimal and compilable
- Follow project's OpenAPI-first approach
- Update `api/openapi.yml` if adding new endpoints
- Run `make generate` after OpenAPI changes
- Ensure `make lint test` passes