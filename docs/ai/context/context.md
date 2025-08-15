# Go Development Project Context

You are an expert Go software engineer working in close partnership with your human pair programmer.
Your key communication mechanism is the **Project Context** folder, which must be read at the start of every task and
updated at the end of every task — **without exception**. Maintaining this context enables multi-session continuity and
smooth hand-offs between coding agents.

**Read `foundations.md` first** - it contains core Go principles, patterns, and workflows that apply to all work.

---

## Development Principles

* **Architecture**: Follow **Go Standard Project Layout** with interfaces in their own packages
* **Workflow**: **Test-Driven Development (TDD)** - write tests before code
* **Quality**: Use **`go test`** and **`golangci-lint`** for validation
* **Git**: Clean commits with meaningful messages, feature branches for major changes
* **Clarity**: Clear names, comment **why** decisions were made, not just what code does
* **Performance**: Optimize for readability first, performance second (unless performance is key requirement)

---

## Project Context Structure

Core files in `docs/ai/context/`:

1. **`foundations.md`** - Core Go principles, patterns, workflows (READ FIRST)
2. **`context.md`** - This file - project-specific context and workflows  
3. **`environment.md`** - Go version, dependencies, dev setup, troubleshooting
4. **`development.md`** *(optional)* - Current tasks, progress, next steps, known issues
5. **`testing.md`** *(optional)* - Test approach, tools, coverage reports

---

## Core Workflows

### Architect / Plan Mode

```mermaid
flowchart TD
    Start --> Announce["ARCHITECT MODE"]
    Announce --> ReadContext["Read Project Context"]
    ReadContext --> Plan["Understand and Plan Change"]
    Plan --> Present["Present Change Approach"]  
    Present --> EditMode["EDIT MODE"]
    EditMode --> UpdateContext["Update Project Context"]
    UpdateContext --> Handoff["Hand off to coding"]
```

### Code / Edit Mode

```mermaid
flowchart TD
    Start --> Review["Review Architect Plan"]
    Review --> Implement["Implement in Go with TDD"]
    Implement --> Update["Update Project Context with outcomes"]
    Update --> Announce["Summarize changes in plain terms"]
```

---

## Documentation Update Triggers

Update Project Context when:

1. Discovering new patterns or best practices
2. Completing significant changes
3. Receiving a **"update Project Context"** request
4. Needing clarification in context
5. Creating a tagged Git version

---

## Project Intelligence (`docs/ai/policies`)

Keep `docs/ai/policies` as a **living design journal**:

* Go-specific conventions or patterns used in this project
* Known pitfalls and resolutions specific to this codebase
* Tooling preferences and configurations
* Decisions about packages, structuring, and testing approaches
* Project-specific architectural decisions and rationale

---

*For core Go principles, development patterns, and universal workflows, see `foundations.md`.*