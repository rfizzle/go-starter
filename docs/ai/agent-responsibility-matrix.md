# AI Agent Responsibility Matrix

This matrix defines clear boundaries and responsibilities for AI agents working with the Go Starter project, ensuring effective coordination and preventing overlapping work.

## Agent Overview

| Agent                     | Primary Focus           | Tools             | Can Write Code | Can Plan |
|---------------------------|-------------------------|-------------------|----------------|----------|
| **AI Documentation**      | AI-facing docs          | Read, Write, Bash | No             | No       |
| **General Documentation** | All other documentation | Read, Write, Bash | No             | No       |
| **Feature**               | Requirements gathering  | Read, Write, Bash | No             | No       |
| **Plan**                  | Implementation planning | Read, Write, Bash | No             | Yes      |
| **Review**                | Code quality analysis   | Read, Bash        | No             | No       |
| **General Purpose**       | Implementation & coding | All tools         | Yes            | Yes      |

## Detailed Responsibilities

### AI Documentation Agent (`docs/ai/prompts/ai.md`)

#### ✅ Primary Responsibilities
- **Scope**: `docs/ai/` directory only - NO EXCEPTIONS
- Create and maintain AI agent prompts
- Update AI documentation templates
- Maintain agent responsibility definitions
- Create context files for AI agents
- Document AI workflow patterns

#### ❌ Out of Scope
- Writing application code
- Making changes outside `docs/ai/` directory
- General documentation (README files, code comments, user guides)
- Feature requirements gathering
- Code implementation planning
- Code review activities

#### 🔄 Handoff Criteria
- **To Feature Agent**: When requirements need to be gathered
- **To Plan Agent**: When implementation planning is needed
- **To General Purpose**: When code needs to be written
- **To General Documentation**: When non-AI documentation needs to be created or updated

---

### General Documentation Agent (`docs/ai/prompts/docs.md`)

#### ✅ Primary Responsibilities
- **Scope**: All documentation EXCEPT `docs/ai/` directory
- Create and maintain README files throughout the project
- Write and update inline Go code documentation and comments
- Maintain user-facing documentation and guides
- Create architecture and development documentation (non-AI specific)
- Can READ from `docs/ai/` for coordination and reference

#### ❌ Out of Scope
- Writing application code
- Making changes within `docs/ai/` directory (AI Documentation Agent's territory)
- Feature requirements gathering
- Code implementation planning
- Code review activities

#### 🔄 Handoff Criteria
- **To Feature Agent**: When requirements need to be gathered
- **To Plan Agent**: When implementation planning is needed
- **To General Purpose**: When code needs to be written
- **To AI Documentation**: When AI-specific documentation needs updates

---

### Feature Agent (`docs/ai/prompts/feature.md`)

#### ✅ Primary Responsibilities
- **Scope**: `docs/features/` directory
- Gather and document feature requirements
- Create user stories and acceptance criteria
- Define success metrics and KPIs
- Document technical requirements
- Follow product owner best practices

#### ❌ Out of Scope
- Creating implementation plans
- Writing application code
- Code review activities
- AI documentation updates
- Technical architecture decisions

#### 🔄 Handoff Criteria
- **To Plan Agent**: When feature document is complete and ready for implementation planning
- **To AI Documentation**: When AI prompts need updates based on new features
- **To General Purpose**: When feature requires immediate prototyping

---

### Plan Agent (`docs/ai/prompts/plan.md`)

#### ✅ Primary Responsibilities
- **Scope**: `docs/plans/` directory and READ-ONLY analysis
- Transform feature documents into implementation plans
- Break down features into specific tasks
- Define Go-specific technical approaches
- Create comprehensive testing strategies
- Assess risks and mitigation strategies

#### ❌ Out of Scope
- Writing application code
- Implementing planned features
- Gathering feature requirements
- Code review activities
- Making code changes

#### 🔄 Handoff Criteria
- **To General Purpose**: When plan is complete and ready for implementation
- **To Feature Agent**: When requirements are incomplete or unclear
- **To Review Agent**: When existing code analysis is needed for planning

---

### Review Agent (`docs/ai/prompts/review.md`)

#### ✅ Primary Responsibilities
- **Scope**: READ-ONLY code analysis + review document creation
- **Output**: Write review results to `docs/reviews/YYYY-MM-DD-HHmm-[feature-name]-review.md`
- Analyze code changes and pull requests
- Create structured review documents with findings and required changes
- Support iterative review cycles (initial → implementation → confirmation)
- Provide constructive feedback on code quality
- Ensure adherence to Go standards and project patterns
- Validate test coverage and quality
- Check for security and performance issues
- Track review status through approval workflow

#### ❌ Out of Scope
- Writing or modifying application code
- Creating implementation plans
- Gathering feature requirements
- AI documentation creation
- Making code changes based on review feedback (delegate to Implementation Agent)

#### 🔄 Handoff Criteria
- **To Implementation Agent**: When review document contains required changes to be addressed
- **To General Purpose**: When complex code changes need to be implemented outside of review fixes
- **To Plan Agent**: When architectural changes require planning
- **To Feature Agent**: When requirements clarification is needed
- **From Implementation Agent**: When fixes are ready for re-review and confirmation

---

### Implementation Agent (`docs/ai/prompts/implementation.md`)

#### ✅ Primary Responsibilities
- **Scope**: Execute tasks from plans (`docs/plans/`) or reviews (`docs/reviews/`)
- Implement features according to implementation plans
- Address issues identified in review documents
- Write and modify application code following specified approaches
- Create and update tests as defined in plans or required by reviews
- Update progress in plan documents or review status
- Support both plan-based and review-based implementation workflows

#### ❌ Out of Scope
- Creating implementation plans (delegate to Plan Agent)
- Creating feature requirements (delegate to Feature Agent)
- Performing code reviews (delegate to Review Agent)
- Modifying plan structure or review findings

#### 🔄 Handoff Criteria
- **From Plan Agent**: When implementation plan is complete and ready for execution
- **From Review Agent**: When review document contains specific required changes
- **To Review Agent**: When implementation is complete and ready for review/re-review
- **To Plan Agent**: When plan modifications are needed during implementation

---

### General Purpose Agent

#### ✅ Primary Responsibilities
- **Scope**: Entire codebase for complex, multi-step tasks
- Handle complex development tasks requiring multiple agent capabilities
- Execute tasks requiring all tool types
- Research and implement features without formal plans

#### ❌ Out of Scope
- Creating feature requirements (delegate to Feature Agent)
- Formal implementation planning (delegate to Plan Agent) 
- Formal code reviews (delegate to Review Agent)
- Routine implementation tasks with existing plans (delegate to Implementation Agent)

#### 🔄 Handoff Criteria
- **To Feature Agent**: When requirements are unclear or missing
- **To Plan Agent**: When complex features need detailed planning
- **To Implementation Agent**: When formal plans or review documents exist for the task
- **To Review Agent**: When code review is requested
- **To AI Documentation**: When AI documentation needs updates

## Task Type Mapping

### Requirements & Documentation
- **User stories** → Feature Agent
- **Acceptance criteria** → Feature Agent
- **API specifications** → Feature Agent
- **AI prompts** → AI Documentation Agent
- **General documentation** → General Documentation Agent
- **Implementation plans** → Plan Agent
- **Review documents** → Review Agent

### Analysis & Review
- **Code quality analysis** → Review Agent
- **Architecture review** → Review Agent
- **Security assessment** → Review Agent
- **Performance analysis** → Review Agent
- **Test coverage analysis** → Review Agent
- **Review document creation** → Review Agent
- **Review-based fixes** → Implementation Agent

### Implementation & Development
- **Plan-based implementation** → Implementation Agent
- **Review-based fixes** → Implementation Agent
- **Complex multi-step development** → General Purpose Agent
- **Ad-hoc bug fixes** → General Purpose Agent
- **Refactoring without plans** → General Purpose Agent
- **Exploratory development** → General Purpose Agent

### Planning & Strategy
- **Technical task breakdown** → Plan Agent
- **Risk assessment** → Plan Agent
- **Architecture decisions** → Plan Agent
- **Testing strategy** → Plan Agent

## Collaboration Patterns

### Sequential Workflows

#### Feature Development Flow
```
1. Feature Agent → Creates requirements in docs/features/
2. Plan Agent → Creates implementation plan in docs/plans/
3. Implementation Agent → Implements according to plan
4. Review Agent → Creates review document in docs/reviews/
5. Implementation Agent → Addresses review findings (if any)
6. Review Agent → Confirms fixes and provides final approval
```

#### Documentation Updates Flow
```
1. AI Documentation → Updates prompts and templates
2. Feature Agent → Updates feature templates if needed
3. Plan Agent → Updates plan templates if needed
```

### Parallel Workflows

#### Multi-Domain Features
- Feature Agent: Gathers requirements
- Plan Agent: Creates technical plans
- Review Agent: Analyzes existing related code and creates review documents
- Implementation Agent: Executes plans and addresses review feedback
- AI Documentation: Updates relevant AI guidance

## Conflict Resolution

### Overlapping Responsibilities

#### When multiple agents could handle a task:
1. **Default to most specific agent** - Use the agent with the narrowest, most relevant scope
2. **Consider tool requirements** - Choose agent with appropriate tool access
3. **Check workflow phase** - Match agent to current development phase

#### Examples:
- **Writing tests from plans**: Implementation Agent (following plan specifications)
- **Writing tests for review fixes**: Implementation Agent (addressing review requirements)
- **Complex test architecture**: General Purpose Agent (multi-step, exploratory)
- **Planning test strategy**: Plan Agent (analysis and documentation)
- **Reviewing test quality**: Review Agent (analysis and feedback via review documents)

### Scope Boundary Disputes

#### Clear Decision Tree:
1. **Is there an existing plan or review document?** → Implementation Agent
2. **Is it complex multi-step development without formal docs?** → General Purpose
3. **Is it requirements/user story related?** → Feature Agent  
4. **Is it planning/architecture related?** → Plan Agent
5. **Is it code analysis/review related?** → Review Agent
6. **Is it AI documentation related?** → AI Documentation Agent
7. **Is it general documentation related?** → General Documentation Agent

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

---

*This matrix ensures clear boundaries and effective collaboration between AI agents while maintaining high quality standards and preventing work overlap.*