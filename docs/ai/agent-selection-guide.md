# Agent Selection Guide

Quick reference for choosing the right AI agent and mode for your task.

## 4-Agent System Overview

| Agent | Focus | Best For | Plan Mode | Edit Mode |
|-------|--------|----------|-----------|-----------|
| **Documentation** | All docs | Specs, READMEs, prompts | ✓ | ✓ |
| **Planning** | Implementation plans | Task breakdown, risks | ✓ | ✓ |  
| **Review** | Code quality | Analysis, standards | ✓ | ✓ |
| **Development** | Code changes | Implementation, fixes | ✓ | ✓ |

## Quick Agent Selection

### 🤔 "I need to..."

**Write or update documentation**
→ Documentation Agent

**Break down a feature into tasks** 
→ Planning Agent

**Analyze code quality or standards**
→ Review Agent

**Implement features or fix bugs**
→ Development Agent

### 🎯 Task Type → Agent

| Task | Agent |
|------|-------|
| Create feature specification | Documentation |
| Write user stories | Documentation |
| Update README files | Documentation |
| Create/update AI prompts | Documentation |
| Plan implementation approach | Planning |
| Break down tasks by package | Planning |
| Assess risks and dependencies | Planning |
| Review code changes | Review |
| Check standards compliance | Review |
| Analyze test coverage | Review |
| Implement new features | Development |
| Fix bugs or issues | Development |
| Add tests | Development |
| Refactor code | Development |

## Mode Selection Guide

### 🎯 **Default: Always Start with Plan Mode**
**Plan Mode First** should be your default approach for any non-trivial task.

### Why Plan Mode First?
- **See before you leap**: Understand what will be done before execution
- **Feedback opportunity**: Refine approach based on your input
- **Prevent rework**: Avoid misunderstood requirements
- **Strategic thinking**: Leverage Opus for comprehensive analysis
- **Quality assurance**: Better outcomes through deliberate planning

### When to Use Plan Mode (Opus)
- **Default choice** for most tasks (when in doubt, use plan mode)
- **All new features** (every step: docs → planning → implementation)
- **Complex modifications** (refactoring, architecture changes)
- **Unfamiliar territory** (new patterns, domains, technologies)  
- **High-risk changes** (security, performance, breaking changes)
- **Multi-step processes** (anything requiring multiple actions)

### When Edit Mode is Appropriate (Sonnet)
- **Following approved plans** (after plan mode was used)
- **Trivial mechanical changes** (typo fixes, version bumps)
- **User explicitly requests edit mode** ("just fix this quickly")
- **Emergency hot fixes** (but plan mode still preferred)

### Examples by Complexity

#### Plan Mode Scenarios (Most Tasks)
```
📝 "Create a feature specification for user authentication"
→ Documentation Agent (plan mode) - present content strategy first

📋 "Create implementation plan for the authentication feature"  
→ Planning Agent (plan mode) - show task breakdown strategy first

💻 "Implement user authentication based on this plan"
→ Development Agent (plan mode) - present technical approach first

🔍 "Review this authentication implementation"
→ Review Agent (plan mode) - present review strategy for complex features
```

#### Edit Mode Scenarios (Rare - Simple Tasks Only)
```
✅ "Fix this typo: 'athentication' should be 'authentication'"
→ Documentation Agent (edit mode) - mechanical fix

✅ "Bump version from 1.0.0 to 1.0.1 in package.json"
→ Development Agent (edit mode) - simple version update

✅ "Implement the authentication feature" (when plan already approved)
→ Development Agent (edit mode) - executing previously planned work
```

### The Plan → Review → Edit Cycle
```
Your Request: "I need user authentication"
     ↓
Documentation Agent (plan): "Here's my content strategy..."
     ↓ (you review and provide feedback)
Documentation Agent (edit): Creates feature spec
     ↓
Planning Agent (plan): "Here's my implementation strategy..."
     ↓ (you review and provide feedback)  
Planning Agent (edit): Creates implementation plan
     ↓
Development Agent (plan): "Here's my technical approach..."
     ↓ (you review and approve)
Development Agent (edit): Implements the feature
```

## Workflow Patterns

### ✅ Recommended: Full Feature Development with Plan Mode
```
1a. Documentation Agent (plan) → Present feature spec strategy
1b. User reviews and provides feedback  
1c. Documentation Agent (edit) → Create feature specification

2a. Planning Agent (plan) → Present implementation strategy
2b. User reviews and provides feedback
2c. Planning Agent (edit) → Create implementation plan

3a. Development Agent (plan) → Present technical approach  
3b. User reviews and approves
3c. Development Agent (edit) → Execute implementation

4. Review Agent (edit) → Analyze completed work
```

### ✅ Bug Fix with Analysis
```
1a. Review Agent (plan) → Present issue analysis approach
1b. User reviews approach
1c. Review Agent (edit) → Analyze the issue and create review doc

2a. Development Agent (plan) → Present fix strategy
2b. User reviews and approves  
2c. Development Agent (edit) → Implement the fix
```

### ⚡ Simple Tasks (Rare)
```
1. Documentation Agent (edit) → Simple typo fix
2. Development Agent (edit) → Version bump
(Only when task is genuinely trivial)
```

## Decision Tree

```
Start Here: What do you want to accomplish?

📝 Documentation/Writing Task
   ├─ Is it a simple typo/mechanical fix? → Documentation Agent (Edit Mode)  
   └─ Everything else? → Documentation Agent (Plan Mode)

📋 Planning/Architecture Task  
   ├─ Following existing template exactly? → Planning Agent (Edit Mode)
   └─ Everything else? → Planning Agent (Plan Mode)

🔍 Code Review/Analysis Task
   ├─ Simple style/format check? → Review Agent (Edit Mode)
   └─ Everything else? → Review Agent (Plan Mode)

💻 Code Implementation Task
   ├─ Trivial change (typo, version bump)? → Development Agent (Edit Mode)
   └─ Everything else? → Development Agent (Plan Mode)

🤔 When in Doubt?
   → Always start with Plan Mode
```

## Pro Tips

### 🎯 **Default Strategy: Plan Mode First**
- **Prevent rework**: Get alignment before execution
- **Better outcomes**: Strategic thinking leads to higher quality  
- **User control**: You see and approve the approach first
- **Feedback integration**: Plans can be refined based on your input

### Maximize Opus (Plan Mode)
- **Strategic analysis**: Deep thinking about approach and risks
- **Comprehensive planning**: Full understanding before execution
- **Requirements clarification**: Surface assumptions and dependencies
- **Quality architecture**: Better design through deliberate planning

### Maximize Sonnet (Edit Mode)
- **Efficient execution**: Fast implementation of approved plans
- **Consistent patterns**: Following established approaches
- **Focused implementation**: Direct execution without decision overhead

### The Plan → Review → Edit Advantage
```
Plan Mode (Opus): "Here's what I propose to do..."
User Review: "Good, but also consider X and Y"  
Edit Mode (Sonnet): Executes refined approach efficiently
```

### When Plan Mode Might Be Overkill
- Single character typos in existing text
- Mechanical version number updates
- Following previously approved detailed plans
- **But even then**: Plan mode is safer if unsure

## Common Mistakes to Avoid

❌ **Jumping straight to Edit Mode** - Start with Plan Mode unless trivial
❌ **Using Development Agent for planning** - Use Planning Agent instead  
❌ **Using Planning Agent for implementation** - Use Development Agent instead  
❌ **Skipping user review of plans** - Always review before proceeding to edit
❌ **Using wrong agent scope** - Match task type to agent specialization  

✅ **Default to Plan Mode first**  
✅ **Review and refine plans before execution**
✅ **Match agent to task type**  
✅ **Follow natural handoff points**  
✅ **Trust agent recommendations for mode and handoffs**