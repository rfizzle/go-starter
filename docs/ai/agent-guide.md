# AI Agent Guide

Comprehensive guide for selecting and working with AI agents in Go development projects.

## Quick Selection

```
📝 Need documentation? → Documentation Agent
📋 Need planning? → Planning Agent
🔍 Need code review? → Review Agent  
💻 Need code changes? → Development Agent
```

## 4-Agent System Overview

| Agent                                         | Focus                | Best For                | Plan Mode | Edit Mode |
|-----------------------------------------------|----------------------|-------------------------|-----------|-----------|
| **[Documentation](prompts/documentation.md)** | All docs             | Specs, READMEs, prompts | ✓         | ✓         |
| **[Planning](prompts/planning.md)**           | Implementation plans | Task breakdown, risks   | ✓         | ✓         |
| **[Review](prompts/review.md)**               | Code quality         | Analysis, standards     | ✓         | ✓         |
| **[Development](prompts/development.md)**     | Code changes         | Implementation, fixes   | ✓         | ✓         |


## Agent Responsibilities

### Documentation Agent (`prompts/documentation.md`)

#### ✅ Primary Responsibilities
- **Scope**: All documentation across the project
- Create feature specifications in `docs/features/`
- Manage AI prompts and templates in `docs/ai/`
- Maintain README files and general documentation
- Write inline code documentation and user guides
- Support both plan mode (design) and edit mode (execution)

#### ❌ Out of Scope
- Writing application code
- Creating implementation plans
- Code review activities
- Code implementation or bug fixes

#### 🔄 Handoff Criteria
- **To Planning Agent**: When feature docs are complete and need implementation planning
- **To Development Agent**: When documentation needs code examples or implementation
- **To Review Agent**: When documentation quality needs assessment

### Planning Agent (`prompts/planning.md`)

#### ✅ Primary Responsibilities
- **Scope**: Implementation planning in `docs/plans/` directory
- Transform feature documents into actionable development plans
- Break down features into package-based implementation stages
- Define Go-specific technical approaches and testing strategies
- Support both plan mode (analysis) and edit mode (planning)
- Assess risks and dependencies

#### ❌ Out of Scope
- Writing application code
- Implementing planned features
- Gathering feature requirements
- Code review activities

#### 🔄 Handoff Criteria
- **To Development Agent**: When implementation plan is complete and ready for execution
- **To Documentation Agent**: When requirements are incomplete or need clarification
- **To Review Agent**: When existing code analysis is needed for planning

### Review Agent (`prompts/review.md`)

#### ✅ Primary Responsibilities
- **Scope**: Code analysis and review document creation in `docs/reviews/`
- Analyze code changes for quality, security, and standards compliance
- Support both plan mode (analysis strategy) and edit mode (review execution)
- Create structured review documents with findings and recommendations
- Support iterative review cycles with status tracking
- Validate test value, risk coverage, and Go conventions

#### ❌ Out of Scope
- Writing or modifying application code
- Creating implementation plans
- Gathering feature requirements
- Making code changes based on review feedback

#### 🔄 Handoff Criteria
- **To Development Agent**: When review identifies issues requiring fixes
- **To Planning Agent**: When architectural changes require planning
- **To Documentation Agent**: When documentation updates are needed

### Development Agent (`prompts/development.md`)

#### ✅ Primary Responsibilities
- **Scope**: All code implementation and modifications
- Execute feature implementation following plans from `docs/plans/`
- Fix issues identified in reviews from `docs/reviews/`
- Support both plan mode (design approach) and edit mode (implementation)
- Handle 3-phase feature implementation or direct bug fixes
- Maintain package boundaries and commit points

#### ❌ Out of Scope
- Creating feature requirements or specifications
- Performing code reviews or creating review documents
- Creating implementation plans (uses existing plans)

#### 🔄 Handoff Criteria
- **From Planning Agent**: When implementation plan is ready for execution
- **From Review Agent**: When review contains specific issues to fix
- **To Review Agent**: When implementation is complete and ready for review

## Task Type Mapping

### Documentation Tasks
- **Feature specifications** → Documentation Agent
- **User stories and acceptance criteria** → Documentation Agent  
- **AI prompts and templates** → Documentation Agent
- **README files** → Documentation Agent
- **Code documentation** → Documentation Agent
- **User guides** → Documentation Agent

### Planning Tasks
- **Implementation plans** → Planning Agent
- **Technical task breakdown** → Planning Agent
- **Risk assessment** → Planning Agent
- **Architecture decisions** → Planning Agent
- **Testing strategy** → Planning Agent
- **Package stage organization** → Planning Agent

### Review Tasks
- **Code quality analysis** → Review Agent
- **Security assessment** → Review Agent
- **Performance analysis** → Review Agent
- **Test-value and risk analysis** → Review Agent
- **Review document creation** → Review Agent
- **Standards compliance** → Review Agent

### Development Tasks
- **Feature implementation** → Development Agent
- **Bug fixes** → Development Agent
- **Interface design and implementation** → Development Agent
- **Test creation and implementation** → Development Agent
- **Review-based fixes** → Development Agent
- **Refactoring tasks** → Development Agent

## Mode Selection

### 🎯 **Default: Always Start with Plan Mode**
Plan Mode should be your default approach for any non-trivial task.

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

## Agent Selection Decision Tree

```
What type of work needs to be done?

📝 Writing/updating documentation?
   → Documentation Agent

📋 Planning implementation tasks?
   → Planning Agent  

🔍 Reviewing/analyzing code?
   → Review Agent

💻 Writing/fixing code?
   → Development Agent
```

### Mode Selection Decision Tree

```
Do you want to see the approach first?
   → Use Plan Mode (Opus)
   
Ready to execute immediately?
   → Use Edit Mode (Sonnet)
   
Unsure of the complexity?
   → Start with Plan Mode
```

### Common Scenarios
- **"Create a new feature"** → Documentation Agent (spec) → Planning Agent (tasks) → Development Agent (code)
- **"Fix this bug"** → Review Agent (analyze) → Development Agent (fix)
- **"Improve documentation"** → Documentation Agent (plan/edit)
- **"Review my changes"** → Review Agent (analyze/document)

## Workflow Patterns

### Sequential Workflows

#### Full Feature Development Flow
```
1. Documentation Agent → Creates feature specifications in docs/features/
2. Planning Agent → Creates implementation plan in docs/plans/
3. Development Agent (Plan Mode) → Analyzes approach and gets approval
4. Development Agent (Edit Mode) → Implements feature in phases
5. Review Agent → Creates review document in docs/reviews/
6. Development Agent → Addresses review findings (if any)
7. Review Agent → Confirms fixes and provides final approval
```

#### Quick Fix Flow
```
1. Review Agent → Identifies issues and creates review document
2. Development Agent (Plan Mode) → Analyzes fix approach
3. Development Agent (Edit Mode) → Implements fixes
4. Review Agent → Validates fixes
```

### Plan/Edit Mode Workflows

#### Opus Planning → Sonnet Execution
```
1. Agent (Plan Mode with Opus) → Detailed analysis and approach design
2. User Approval → Review and approve the plan
3. Same Agent (Edit Mode with Sonnet) → Execute the approved plan
```

This workflow maximizes the strengths of each model:
- **Opus**: Deep analysis, comprehensive planning, risk assessment
- **Sonnet**: Efficient execution, focused implementation

## Escalation and Handoffs

### Simple Escalation Rules

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

### Agent-Specific Handoffs

#### Documentation Agent Handoffs
- **Code implementation needed** → Development Agent
- **Complex planning required** → Planning Agent  
- **Code quality assessment** → Review Agent

#### Planning Agent Handoffs
- **Feature requirements unclear** → Documentation Agent
- **Ready to implement plan** → Development Agent
- **Code analysis needed** → Review Agent

#### Review Agent Handoffs
- **Issues need fixing** → Development Agent
- **Requirements clarification** → Documentation Agent
- **Planning architectural changes** → Planning Agent

#### Development Agent Handoffs  
- **Need feature specification** → Documentation Agent
- **Need implementation plan** → Planning Agent
- **Ready for code review** → Review Agent

### Standard Handoff Message

```
"This task requires [specific capability] which is outside my focus area. 

The [appropriate agent] would be better suited because they specialize in [relevant area].

Would you like me to recommend starting with that agent in [plan/edit] mode?"
```

### Natural Handoff Points

#### Documentation → Planning  
When feature specifications are complete and ready for technical breakdown.

#### Planning → Development
When implementation plans are complete and ready for execution.

#### Development → Review
When implementation is complete and ready for quality assessment.

#### Review → Development  
When review identifies issues that need fixing.

## Quality Gates

### Before Agent Handoff
- [ ] Current agent has completed their scope of work
- [ ] Work products meet quality standards for that agent type
- [ ] Clear context provided to receiving agent
- [ ] Dependencies and blockers identified
- [ ] Success criteria defined for next phase

### Agent Scope Validation
- [ ] Agent is working within defined scope boundaries
- [ ] Agent has appropriate tools for the task
- [ ] Task type matches agent capabilities
- [ ] No overlap with other active agents
- [ ] Handoff criteria clearly defined

## Escalation Triggers

### When to Change Agents

#### Immediate Escalation Required:
- Agent attempts work outside defined scope
- Agent lacks necessary tools for task completion
- Agent encounters blockers requiring different expertise
- Task type doesn't match agent capabilities
- Quality standards cannot be met by current agent

#### Planned Handoffs:
- Work phase completion (Feature → Plan → Implementation)
- Document creation complete and ready for next phase
- Analysis complete and actionable items identified
- Implementation ready for review and validation

## Best Practices

### For Users Working with Agents
1. **Start with the right agent** - Choose based on task type and current phase
2. **Provide clear context** - Include relevant background and constraints
3. **Define success criteria** - Specify what "done" looks like
4. **Respect scope boundaries** - Don't ask agents to work outside their defined areas
5. **Use handoff points** - Allow agents to complete their phase before switching

### For Agent Coordination
1. **Clear scope adherence** - Always work within defined boundaries
2. **Explicit handoff communication** - Clearly state when and why handing off
3. **Context preservation** - Provide complete context to receiving agent
4. **Quality validation** - Ensure work meets standards before handoff
5. **Documentation updates** - Keep agent documentation current with role changes

## Quick Reference

| **If you need...**      | **Use this agent**  | **Typical mode**         |
|-------------------------|---------------------|--------------------------|
| Feature specifications  | Documentation Agent | Plan → Edit              |
| User stories            | Documentation Agent | Plan → Edit              |
| README updates          | Documentation Agent | Edit (if simple)         |
| AI prompt changes       | Documentation Agent | Plan → Edit              |
| Implementation planning | Planning Agent      | Plan → Edit              |
| Task breakdown          | Planning Agent      | Plan → Edit              |
| Risk assessment         | Planning Agent      | Plan → Edit              |
| Code review             | Review Agent        | Plan → Edit              |
| Quality analysis        | Review Agent        | Plan → Edit              |
| Standards compliance    | Review Agent        | Edit                     |
| Feature implementation  | Development Agent   | Plan → Edit              |
| Bug fixes               | Development Agent   | Plan → Edit              |
| Code changes            | Development Agent   | Edit (if following plan) |


---

*This guide ensures clear boundaries and effective collaboration between AI agents while maintaining high quality standards and preventing work overlap.*
