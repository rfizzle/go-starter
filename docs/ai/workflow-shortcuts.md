# Workflow Shortcuts

Quick workflows for common development tasks using the 4-agent system with plan/edit modes.

## Quick Reference

### 🚀 Express Workflows (Edit Mode)

**Simple Bug Fix**
```
Development Agent (edit) → Fix issue directly
```

**Update Documentation**  
```
Documentation Agent (edit) → Make changes directly
```

**Quick Code Review**
```
Review Agent (edit) → Analyze and provide feedback
```

### 🎯 Guided Workflows (Plan → Edit)

**New Feature Development**
```
1. Documentation Agent (plan) → Create feature spec
2. Planning Agent (plan) → Break down implementation  
3. Development Agent (plan) → Design approach
4. Development Agent (edit) → Implement code
5. Review Agent (edit) → Validate work
```

**Major Refactoring**
```
1. Review Agent (plan) → Analyze current state
2. Planning Agent (plan) → Create refactoring strategy  
3. Development Agent (plan) → Design new structure
4. Development Agent (edit) → Execute changes
5. Review Agent (edit) → Validate improvements
```

## Common Task Patterns

### Feature Implementation

#### 📋 From User Request
```
User: "I need user authentication for the API"

Step 1a: Documentation Agent (plan mode)
→ "Plan the feature specification for user authentication"
→ Agent presents: Problem statement, user stories structure, acceptance criteria approach

Step 1b: User reviews plan and provides feedback
→ "Add mobile app considerations"
→ "Include performance requirements"

Step 1c: Documentation Agent (edit mode)
→ Creates actual feature specification in docs/features/

Step 2a: Planning Agent (plan mode)  
→ "Plan implementation for the authentication feature"
→ Agent presents: Package stages, task organization, risk areas

Step 2b: User reviews plan and provides feedback
→ "Break auth into smaller tasks"
→ "Add security testing phase"

Step 2c: Planning Agent (edit mode)
→ Creates implementation plan document in docs/plans/

Step 3a: Development Agent (plan mode)
→ "Plan the implementation approach for this plan"
→ Agent presents: Technical approach, phase strategy, implementation sequence

Step 3b: User reviews and approves approach

Step 3c: Development Agent (edit mode)
→ Implements the authentication feature

Step 4: Review Agent (edit mode)
→ Reviews the authentication implementation
```

#### ⚡ From Existing Spec
```
If you already have a feature specification:

Step 1a: Planning Agent (plan mode)
→ "Plan implementation from this feature spec"
→ Agent presents: Task breakdown strategy, package stages, delivery phases

Step 1b: User reviews and provides feedback

Step 1c: Planning Agent (edit mode)
→ Creates implementation plan document in docs/plans/

Step 2a: Development Agent (plan mode)
→ "Plan the implementation approach for this plan"
→ Agent presents: Technical strategy, phase approach

Step 2b: User reviews and approves

Step 2c: Development Agent (edit mode)
→ Implements the feature

Step 3: Review Agent (edit mode)
→ Reviews the implementation
```

### Bug Fixes

#### 🐛 Simple Fix
```
User: "Fix this typo in the login error message"

Development Agent (edit mode)
→ "Fix the typo in [file location]"
```

#### 🔍 Complex Issue
```
User: "Users are experiencing intermittent login failures"

Step 1: Review Agent (plan mode)
→ "Analyze the login system for potential issues"

Step 2: Development Agent (plan mode)  
→ "Design fix approach for the identified issues"

Step 3: Development Agent (edit mode)
→ "Implement the fixes"

Step 4: Review Agent (edit mode)
→ "Verify the fixes resolve the issues"
```

### Documentation Tasks

#### 📝 New Documentation
```
User: "Create API documentation for the new endpoints"

Step 1a: Documentation Agent (plan mode)
→ "Plan comprehensive API documentation for the endpoints"
→ Agent presents: Documentation structure, sections to include, code examples needed

Step 1b: User reviews plan and provides feedback
→ "Add authentication examples"
→ "Include error handling patterns"

Step 1c: Documentation Agent (edit mode)
→ Creates the API documentation with all requested elements
```

#### ✏️ Complex Documentation Updates
```
User: "Restructure the README to be more user-friendly"

Step 1a: Documentation Agent (plan mode)
→ "Plan README restructuring for better user experience"
→ Agent presents: New structure, content organization, sections to add/remove

Step 1b: User reviews plan and provides feedback

Step 1c: Documentation Agent (edit mode)
→ Implements the README restructuring
```

#### ⚡ Simple Documentation Updates
```
User: "Update the README with new installation steps"

Documentation Agent (edit mode)
→ Updates README with new installation instructions
(Simple updates can skip plan mode)
```

### Code Reviews

#### 🔎 Standard Review (Simple)
```
User: "Review my changes for the user profile feature"

Review Agent (edit mode)
→ Reviews the user profile feature implementation directly
(Standard reviews can often skip plan mode)
```

#### 🏗️ Complex Architecture Review
```
User: "Review this major system redesign"

Step 1a: Review Agent (plan mode)
→ "Plan systematic review of the system redesign"
→ Agent presents: Review strategy, areas to analyze, quality criteria to check

Step 1b: User reviews plan and provides feedback
→ "Focus on security implications"
→ "Check performance impact"

Step 1c: Review Agent (edit mode)
→ Executes the comprehensive architectural review
```

#### 🔍 Security/Performance Review
```
User: "Review this code for security vulnerabilities"

Step 1a: Review Agent (plan mode)
→ "Plan security-focused review approach"
→ Agent presents: Security areas to check, vulnerability patterns to look for

Step 1b: User reviews and approves approach

Step 1c: Review Agent (edit mode)
→ Performs detailed security review
```

## Mode Selection Shortcuts

### 🎯 Default: Plan Mode First
**Start with Plan Mode unless the task is clearly simple**
- Present approach and get approval before executing
- Allows for feedback and refinement
- Prevents rework from misunderstood requirements
- Maximizes Opus strategic thinking capabilities

### Always Use Plan Mode For:
- **All new features** (documentation → planning → implementation)
- **Complex modifications** (architecture, refactoring, system changes)
- **Unfamiliar territory** (new domains, patterns, or technologies)
- **High-risk changes** (security, performance, breaking changes)
- **Multi-step processes** (anything requiring multiple actions)
- **When in doubt** (better to plan than assume)

### Edit Mode Only For:
- **Simple typo fixes** (single character or word changes)
- **Version number updates** (clear, mechanical changes)
- **Following approved plans** (when plan mode was already used)
- **Trivial updates** (adding a single line of documentation)
- **Configuration tweaks** (known, low-risk changes)

### Plan → Review → Edit Pattern:
```
Step a: Agent (Plan Mode with Opus) → Present strategy and approach
         ↓
Step b: User Review → Provide feedback and approval  
         ↓
Step c: Agent (Edit Mode with Sonnet) → Execute approved plan
```

### When to Skip Plan Mode:
- User explicitly requests edit mode: "just fix this typo"
- Following an existing detailed plan or specification
- Task is genuinely trivial (< 5 minutes, single file, no decisions)
- Emergency hot fixes (but plan mode still preferred)

## Agent Handoff Shortcuts

### Natural Handoff Chain
```
📝 Documentation → 📋 Planning → 💻 Development → 🔍 Review
```

### Quick Handoffs
```
"I need this reviewed" → Review Agent
"I need this implemented" → Development Agent  
"I need this planned" → Planning Agent
"I need this documented" → Documentation Agent
```

## Prompt Templates

### Plan Mode Requests
```
"Plan [task description]"
"Design approach for [requirement]"  
"Analyze [system/code] for [purpose]"
"Create strategy for [goal]"
```

### Edit Mode Requests
```  
"Implement [specific task]"
"Fix [specific issue]"
"Update [specific file/section]"
"Create [specific deliverable]"
```

### Mode Switching
```
"Switch to edit mode and implement this plan"
"Execute the approach we just designed"  
"Proceed with implementation"
```

## Time-Saving Tips

### 💡 Batch Similar Tasks
Group related tasks for the same agent:
```
Documentation Agent (plan):
- Plan README updates
- Plan API documentation  
- Plan user guide improvements
→ Then execute all in edit mode
```

### 🔄 Use Existing Artifacts
Reference previous work:
```
"Use the plan from docs/plans/2024-01-15-auth.md"
"Follow the pattern from the user-service implementation"
"Apply the review feedback from docs/reviews/2024-01-14.md"
```

### ⚡ Skip Planning for Routine Tasks
Use edit mode directly for:
- Dependency updates
- Version bumps  
- Configuration tweaks
- Documentation typos
- Simple bug fixes

### 🎯 Combine Modes Strategically
```
Complex Feature = Plan Mode → Edit Mode
Simple Task = Edit Mode Only
Unknown Complexity = Start Plan Mode
```

## Troubleshooting Workflows

### Agent Says "Wrong Mode"
→ Switch modes or pick different agent

### Agent Says "Out of Scope"  
→ Check agent selection guide, use recommended agent

### Stuck in Planning
→ Ask agent to "switch to edit mode and implement"

### Implementation Fails
→ Go back to plan mode, reassess approach

### Quality Issues
→ Use Review Agent to identify problems, then Development Agent to fix