# AI Agent Escalation Criteria

Simple guidelines for when AI agents should hand off tasks to maintain effective collaboration and quality outcomes.

## Overview

With our simplified 4-agent system, escalation happens when an agent encounters work outside their primary focus area. Each agent should clearly identify handoff opportunities and recommend the appropriate agent.

## Simple Escalation Rules

```
Task Request
     ↓
Is this my primary focus area?
     ↓ YES                    ↓ NO
Do I have the right mode?    Recommend appropriate agent
     ↓ YES        ↓ NO       and explain why
Proceed      Ask for plan
             or edit mode
```

## Agent-Specific Handoffs

### Documentation Agent Handoffs
- **Code implementation needed** → Development Agent
- **Complex planning required** → Planning Agent  
- **Code quality assessment** → Review Agent

### Planning Agent Handoffs
- **Feature requirements unclear** → Documentation Agent
- **Ready to implement plan** → Development Agent
- **Code analysis needed** → Review Agent

### Review Agent Handoffs
- **Issues need fixing** → Development Agent
- **Requirements clarification** → Documentation Agent
- **Planning architectural changes** → Planning Agent

### Development Agent Handoffs  
- **Need feature specification** → Documentation Agent
- **Need implementation plan** → Planning Agent
- **Ready for code review** → Review Agent

## Mode Selection Guidance

### When to Request Plan Mode
- Complex or unfamiliar tasks
- High-risk changes
- Need to understand approach first
- Want to see full scope before committing

### When to Use Edit Mode  
- Clear, well-defined tasks
- Following an existing plan
- Quick fixes or updates
- Confident in the approach

## Standard Handoff Message

```
"This task requires [specific capability] which is outside my focus area. 

The [appropriate agent] would be better suited because they specialize in [relevant area].

Would you like me to recommend starting with that agent in [plan/edit] mode?"
```

## Natural Handoff Points

### Documentation → Planning  
When feature specifications are complete and ready for technical breakdown.

### Planning → Development
When implementation plans are complete and ready for execution.

### Development → Review
When implementation is complete and ready for quality assessment.

### Review → Development  
When review identifies issues that need fixing.

## Quick Reference

| **If you need...** | **Use this agent** |
|---------------------|---------------------|
| Feature specifications | Documentation Agent |
| User stories | Documentation Agent |
| README updates | Documentation Agent |
| AI prompt changes | Documentation Agent |
| Implementation planning | Planning Agent |
| Task breakdown | Planning Agent |
| Risk assessment | Planning Agent |
| Code review | Review Agent |
| Quality analysis | Review Agent |
| Standards compliance | Review Agent |
| Feature implementation | Development Agent |  
| Bug fixes | Development Agent |
| Code changes | Development Agent |

Remember: Start with Plan Mode for complex tasks, use Edit Mode for straightforward execution.