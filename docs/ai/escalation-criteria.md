# AI Agent Escalation Criteria

Detailed criteria and procedures for when AI agents should escalate, delegate, or hand off tasks to maintain effective collaboration and high-quality outcomes.

## Overview

AI agents should escalate when they encounter situations that fall outside their scope, exceed their capabilities, or require different expertise. This document provides specific, actionable criteria for recognizing and handling these situations.

## Escalation Decision Tree

```
Task Request
     ↓
Does agent have appropriate tools? ──NO──→ Escalate to General Purpose
     ↓ YES
Is task within agent scope? ──NO──→ Escalate to appropriate agent
     ↓ YES
Can quality standards be met? ──NO──→ Escalate with blockers identified
     ↓ YES
Proceed with task
```

## Immediate Escalation Triggers

### Tool Limitations

#### When to escalate: Agent lacks required tools for task completion

**Examples:**
- Feature Agent asked to modify code (needs Write/Edit tools)
- Review Agent asked to fix issues (needs Write/Edit tools) 
- Plan Agent asked to implement features (needs all implementation tools)

**Escalation Response:**
```
"I cannot complete this task because it requires [specific tools] that I don't have access to. This task should be handled by [appropriate agent] who has [required tools]. 

The [appropriate agent] can: [list specific capabilities needed]

Would you like me to hand off this task?"
```

### Scope Boundary Violations

#### When to escalate: Task falls outside defined agent scope

**Examples:**
- AI Documentation Agent asked to write application code
- Feature Agent asked to create implementation plans
- Plan Agent asked to gather user requirements

**Escalation Response:**
```
"This task falls outside my defined scope of [agent scope]. I am designed to work specifically on [agent responsibilities].

This task would be better handled by [appropriate agent] because [reason].

I can [offer related services within scope] if that would be helpful."
```

### Quality Standard Violations

#### When to escalate: Agent cannot meet project quality standards

**Examples:**
- Insufficient context to create proper documentation
- Missing dependencies that prevent proper analysis
- Technical complexity exceeding agent's design parameters

**Escalation Response:**
```
"I cannot complete this task to the required quality standards because [specific reason]. 

To proceed, I would need:
- [Missing information/context]
- [Required dependencies]
- [Additional clarification]

Alternatively, [appropriate agent] might be better suited for this task because [reason]."
```

## Planned Handoff Criteria

### Phase Completion Handoffs

#### Feature Agent → Plan Agent
**Trigger:** Feature document is complete and approved

**Handoff Criteria:**
- [ ] All acceptance criteria defined
- [ ] Technical requirements documented
- [ ] Success metrics established
- [ ] Stakeholder sign-off obtained
- [ ] Dependencies identified

**Handoff Message:**
```
"Feature requirements are complete. This feature is ready for implementation planning.

Key details for planning:
- Feature scope: [summary]
- Critical requirements: [list]
- Constraints: [list]
- Success criteria: [list]

Handing off to Plan Agent for technical breakdown and task definition."
```

#### Plan Agent → General Purpose Agent
**Trigger:** Implementation plan is complete and validated

**Handoff Criteria:**
- [ ] Tasks are specific and actionable (4-8 hours each)
- [ ] Dependencies clearly identified
- [ ] Risk assessment complete
- [ ] Technical approach validated
- [ ] Quality standards defined

**Handoff Message:**
```
"Implementation plan is complete and ready for execution.

Plan summary:
- Total effort: [story points/hours]
- Key phases: [list]
- Critical path: [tasks]
- Risks identified: [list with mitigation]

All tasks are defined with specific acceptance criteria. Ready for implementation."
```

#### Implementation → Review Agent
**Trigger:** Code changes are complete and ready for review

**Handoff Criteria:**
- [ ] All planned functionality implemented
- [ ] Tests written and passing
- [ ] Code formatted and linted
- [ ] Documentation updated
- [ ] Self-review completed

**Handoff Message:**
```
"Implementation is complete and ready for review.

Changes summary:
- Features implemented: [list]
- Files modified: [list]
- Test coverage: [percentage]
- Known issues: [list if any]

Please review for code quality, adherence to standards, and overall implementation approach."
```

### Cross-Domain Handoffs

#### Any Agent → AI Documentation Agent
**Trigger:** AI documentation needs updates based on agent work

**Examples:**
- New agent patterns discovered
- Template improvements identified
- Workflow optimizations found
- Agent scope clarifications needed

**Handoff Message:**
```
"I've identified improvements needed for AI documentation:

Updates needed:
- [Template/prompt changes required]
- [New patterns discovered]
- [Workflow improvements]

Context for updates:
- [Specific examples]
- [Lessons learned]
- [Recommended changes]

Handing off to AI Documentation Agent for implementation."
```

## Escalation Scenarios and Responses

### Scenario 1: Insufficient Context

**Situation:** Agent receives task without sufficient background information

**Decision Criteria:**
- Can agent make reasonable assumptions?
- Are the assumptions likely to be correct?
- Would incorrect assumptions cause significant rework?

**Response Options:**

**Minor Context Gaps:**
```
"I need clarification on [specific point] to ensure I complete this correctly. 

Based on the information provided, I'm assuming [assumption]. Is this correct?

I can proceed with this assumption or wait for clarification."
```

**Major Context Gaps:**
```
"I don't have sufficient context to complete this task to quality standards. 

I need:
- [Specific information required]
- [Background context needed]
- [Clarification on requirements]

Without this information, I risk creating work that doesn't meet your needs. Could you provide these details?"
```

### Scenario 2: Technical Complexity

**Situation:** Task requires technical knowledge or capabilities beyond agent design

**Decision Criteria:**
- Does task require deep technical expertise?
- Are there established patterns to follow?
- Would agent need to make architectural decisions?

**Response for High Complexity:**
```
"This task involves [technical complexity] that exceeds my design parameters. 

Specific challenges:
- [Technical aspects beyond scope]
- [Architectural decisions required]
- [Expertise areas needed]

This would be better handled by [appropriate agent/specialist] who can:
- [Relevant capabilities]
- [Technical expertise areas]
- [Decision-making authority]
```

### Scenario 3: Conflicting Requirements

**Situation:** Agent receives conflicting or contradictory instructions

**Response:**
```
"I've identified conflicting requirements that need resolution:

Conflict 1: [Requirement A] vs [Requirement B]
Conflict 2: [Requirement C] vs [Requirement D]

To proceed, I need clarification on:
- [Priority/preference decisions]
- [Trade-off considerations]
- [Stakeholder preferences]

Which approach should I take, or should this be escalated for stakeholder input?"
```

### Scenario 4: Dependency Blockers

**Situation:** Agent cannot proceed due to external dependencies

**Response:**
```
"I cannot complete this task due to the following blockers:

Blockers:
- [Dependency 1]: [Impact and required resolution]
- [Dependency 2]: [Impact and required resolution]

Options:
1. Wait for blocker resolution: [timeline if known]
2. Partial completion: [what can be done now]
3. Alternative approach: [if available]
4. Escalate to resolve dependencies: [who can help]

How would you like to proceed?"
```

## Quality Gates for Escalation

### Before Escalating

**Verify these criteria are met:**
- [ ] Escalation reason is clearly identified
- [ ] Agent has attempted reasonable solutions within scope
- [ ] Specific needs for escalation are documented
- [ ] Appropriate receiving agent is identified
- [ ] Context for handoff is prepared
- [ ] User expectations are set appropriately

### During Handoff

**Provide complete context:**
- [ ] Original task and requirements
- [ ] Work completed so far
- [ ] Specific reason for handoff
- [ ] Remaining work needed
- [ ] Constraints and considerations
- [ ] Quality standards and success criteria

### After Escalation

**Follow-up responsibilities:**
- [ ] Confirm receiving agent understands requirements
- [ ] Provide clarification if requested
- [ ] Document lessons learned for future tasks
- [ ] Update agent documentation if patterns identified

## Common Anti-Patterns to Avoid

### ❌ Don't Do These

**Attempting work outside scope:**
```
"I'll try to implement this even though I don't have the right tools..."
```
**Instead:** Immediately escalate with clear reasoning

**Making assumptions without validation:**
```
"I'm assuming you want X, so I'll proceed with that..."
```
**Instead:** Ask for clarification on important assumptions

**Providing partial work without explanation:**
```
"Here's what I could do, but it's incomplete..."
```
**Instead:** Explain limitations and recommend appropriate agent

**Escalating without context:**
```
"I can't do this. Someone else should handle it."
```
**Instead:** Provide specific reasons and recommendations

## Best Practices for Effective Escalation

### ✅ Do These

1. **Be Specific:** Clearly explain what you can't do and why
2. **Offer Alternatives:** Suggest appropriate agents or approaches
3. **Preserve Context:** Provide complete background for handoff
4. **Set Expectations:** Clarify what the receiving agent can accomplish
5. **Document Patterns:** Note recurring escalation triggers for improvement

### Example of Excellent Escalation

```
"I cannot complete this feature implementation because it requires writing and modifying application code, which is outside my scope as a Plan Agent. I can only analyze and create plans.

What I've accomplished:
- Analyzed the feature requirements
- Created detailed task breakdown
- Identified technical approach
- Assessed risks and mitigation strategies

What's needed next:
- Implementation of the planned tasks
- Code writing and testing
- Integration with existing systems

Recommendation: This should be handed off to a General Purpose Agent who has access to Write, Edit, and all implementation tools needed to execute the plan I've created.

The implementation plan is complete and ready for execution. Would you like me to hand this off now?"
```

## Escalation Metrics and Improvement

### Tracking Escalation Effectiveness

**Success Indicators:**
- Escalations resolved by appropriate agents
- No repeated escalations for same task types
- User satisfaction with handoff process
- Reduced time to task completion

**Improvement Triggers:**
- Frequent escalations for same task types
- User confusion about agent capabilities
- Quality issues after handoffs
- Inefficient handoff processes

### Continuous Improvement

**Regular Review Questions:**
1. Are agent scopes clearly defined and understood?
2. Are escalation criteria being followed consistently?
3. Are handoffs providing sufficient context?
4. Are there patterns indicating scope adjustments needed?
5. Is the escalation process efficient for users?

---

*These criteria ensure AI agents work effectively within their designed parameters while providing seamless handoffs when different expertise is needed.*