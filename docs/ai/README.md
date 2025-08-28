# AI Agent Documentation

Simplified AI agent system for Go development with plan/edit mode support optimized for Opus planning and Sonnet execution.

## 🎯 Quick Start

**Need to choose an agent?** → See [Agent Selection Guide](agent-selection-guide.md)  
**Want workflow shortcuts?** → See [Workflow Shortcuts](workflow-shortcuts.md)  
**Need to hand off tasks?** → See [Escalation Criteria](escalation-criteria.md)

## 📁 Directory Structure

```
ai/
├── README.md                     # This overview
├── agent-selection-guide.md      # Quick agent & mode selection  
├── workflow-shortcuts.md         # Common task patterns
├── agent-responsibility-matrix.md # Detailed agent boundaries
├── escalation-criteria.md        # When to switch agents
├── prompts/                      # 4 core AI agents
│   ├── documentation.md          # All documentation tasks
│   ├── planning.md               # Implementation planning  
│   ├── review.md                 # Code quality analysis
│   └── development.md            # Feature implementation & fixes
└── templates/                    # Document templates
    ├── feature.md                # Feature specifications
    ├── plan.md                   # Implementation plans
    ├── review.md                 # Code review reports
    └── prompt.md                 # AI agent prompts
```

## 4-Agent System

| Agent | Focus | Plan Mode (Opus) | Edit Mode (Sonnet) |
|-------|-------|------------------|-------------------|
| **[Documentation](prompts/documentation.md)** | All docs & specs | Design content strategy | Write/update docs |
| **[Planning](prompts/planning.md)** | Task breakdown | Analyze & strategize | Create implementation plans |  
| **[Review](prompts/review.md)** | Code quality | Plan review approach | Execute code analysis |
| **[Development](prompts/development.md)** | Implementation | Design approach | Write/fix code |

## Plan/Edit Mode Workflow

### Optimal Usage Pattern
```
1. Agent (Plan Mode with Opus) → Comprehensive analysis & strategy
2. User Review & Approval → Validate approach
3. Same Agent (Edit Mode with Sonnet) → Focused execution
```

### Mode Selection
- **Plan Mode**: Complex tasks, unfamiliar territory, high-risk changes
- **Edit Mode**: Clear tasks, following plans, quick fixes

## Quick Agent Selection

```
📝 Need documentation? → Documentation Agent
📋 Need planning? → Planning Agent
🔍 Need code review? → Review Agent  
💻 Need code changes? → Development Agent
```

## Common Workflows

### 🚀 Full Feature Development
```
1. Documentation Agent (plan) → Feature specification
2. Planning Agent (plan) → Implementation breakdown  
3. Development Agent (plan) → Technical approach
4. Development Agent (edit) → Code implementation
5. Review Agent (edit) → Quality validation
```

### ⚡ Quick Bug Fix
```
1. Review Agent (plan) → Analyze issue
2. Development Agent (edit) → Implement fix
```

## Key Benefits

### Simplified Structure
- **4 focused agents** instead of 7 overlapping ones
- **Clear boundaries** between documentation, planning, review, and development
- **Reduced complexity** in agent selection

### Plan/Edit Mode Optimization
- **Plan Mode (Opus)**: Deep analysis, strategic thinking, comprehensive planning
- **Edit Mode (Sonnet)**: Efficient execution, focused implementation
- **User approval gates** between planning and execution

### Better Workflow Support
- **Natural handoff points** between agents
- **Package-stage implementation** with commit boundaries
- **Workflow shortcuts** for common patterns

## Getting Help

- **Agent Selection**: Use the [Agent Selection Guide](agent-selection-guide.md)
- **Common Tasks**: Check [Workflow Shortcuts](workflow-shortcuts.md)  
- **Handoffs**: Reference [Escalation Criteria](escalation-criteria.md)
- **Detailed Boundaries**: See [Agent Responsibility Matrix](agent-responsibility-matrix.md)

## Templates

Available in the `templates/` directory:
- **Feature specifications** - User stories and acceptance criteria
- **Implementation plans** - Task breakdown with package stages
- **Review reports** - Code quality analysis and recommendations
- **AI prompts** - Template for creating new agents

---

*Simplified AI agent system optimized for manual triggering with Opus planning and Sonnet execution.*