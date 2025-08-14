# AI Development Workflow

This project uses Claude Code with hooks for streamlined development.

## Quick Start

### 1. Plan a Feature
```bash
claude
> plan: Add user authentication system
```

The `plan:` prefix automatically loads project context and creates a structured implementation plan.

### 2. Implement Tasks  
```bash
claude
> task: Implement the authentication service from the plan
```

The `task:` prefix automatically loads project context and implementation guidelines.

### 3. Review Code Changes  
```bash
claude
> review: Check the recent authentication changes
```

The `review:` prefix automatically loads code review guidelines.

## How It Works

Claude Code hooks automatically inject context:
- `plan:` → Loads project context + planning guidelines for breaking down features
- `task:` or `implement:` → Loads project context + task implementation guidelines
- `review:` → Loads project context + code review guidelines  

## Feature Tracking

Each feature automatically gets its own directory with structured documentation:

```
docs/features/<feature-name>/
├── plan.md              # Implementation plan with phases
└── development-log.md   # Progress tracking with timestamps
```

### Auto-Generated Documentation
- **Planning**: `plan:` commands create detailed implementation plans
- **Progress Tracking**: `task:` commands automatically update development logs  
- **Context Loading**: Existing plans and logs are automatically loaded for relevant commands

## Files Structure

- **Project context**: `docs/project-context.md` - Consolidated project info
- **Planning guidance**: `docs/ai/prompts/plan-agent.md` - Feature planning and breakdown
- **Task guidance**: `docs/ai/prompts/task-agent.md` - Implementation best practices  
- **Review guidance**: `docs/ai/prompts/review-agent.md` - Code review checklist
- **Hooks config**: `.claude/hooks.json` - Automatic context loading
- **Feature docs**: `docs/features/*/` - Per-feature plans and progress logs

## Benefits

- **Structured planning**: Every feature gets a detailed implementation plan
- **Progress tracking**: Automatic development logs with timestamps and file changes
- **Context awareness**: Existing plans and logs automatically inform new work
- **No setup overhead**: Just use prefixes - documentation is created automatically
- **Team visibility**: Clear audit trail of planning decisions and implementation progress
- **Native tooling**: Uses Claude Code's built-in capabilities