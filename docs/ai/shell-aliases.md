# Shell Aliases for AI Workflow

Add these aliases to your shell configuration (`.bashrc`, `.zshrc`, etc.):

## Direct Agent Starters

```bash
# AI Planning session
alias ai-plan='claude -p "plan: "'

# AI Task session  
alias ai-task='claude -p "task: "'

# AI Review session
alias ai-review='claude -p "review: "'
```

## Usage Examples

```bash
# Start planning session
ai-plan "Add user authentication system"

# Start implementation session
ai-task "Implement Phase 1 from the auth plan"

# Start review session
ai-review "Check the authentication implementation"
```

## Project-Specific Scripts

The project also includes bash scripts in `scripts/`:

```bash
# Using project scripts
./scripts/ai-plan.bash "user authentication system"
./scripts/ai-task.bash "implement auth service"
./scripts/ai-review.bash

# Using Makefile
make ai-plan FEATURE="user authentication system"
make ai-task TASK="implement auth service"  
make ai-review
```

## Benefits

- **One command**: Directly starts Claude Code with agent role
- **Context loaded**: Hooks automatically load project context
- **Fast iteration**: No typing prefixes manually each time
- **Consistent workflow**: Same commands across team members