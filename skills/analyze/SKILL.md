---
name: analyze
description: >
  Analyze existing projects for a new feature development, breaking the analysis into tasks
  delegated across specialized subagents and compiling the results into a note.
  Use when the user wants to scope, assess the impact of, or plan a new feature or change before
  building it — e.g. "analyze TICKET-123", "what would it take to build X", "which projects are
  impacted by Y", "scope this feature before we implement it".
  Not for diagnosing an existing bug (investigate the cause instead) or for writing the code.
allowed-tools: Read, Grep, Glob, Bash, Task, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---

# Analyze

Break a complex analysis into tasks delegated across subagents. Output is a note.

**Input**: the description of the feature or change to be developed. May be a ticket number, and
may name the related projects.

## Steps

1. **If no input provided, ask what the user wants**

   Use the **AskUserQuestion tool** (open-ended, no preset options) to ask:
   > "What do you want to analyze?"

   **IMPORTANT**: Do NOT proceed without understanding what the user wants to build.

2. **Create the analysis note**

   **IMPORTANT**: Use the **note-write skill** and the **note-search skill** to write notes.
   Search the existing notes first — a past note may already cover this area.

   If a ticket number is given and an issue-tracker MCP is connected (Jira, Linear, GitHub issues),
   fetch the ticket for context.

   This creates the note that will hold the analysis results.

3. **List impacted projects**

   Take them from the description, or determine them yourself. The goal is to know each impacted
   project's tech stack (Go, React, JavaScript, infra), because that is what selects the subagent.

   Read the `CLAUDE.md` / `AGENTS.md` of every project in scope — their conventions and constraints
   belong in the analysis.

4. **Analyze**

   Use TaskCreate, TaskGet, TaskList, and TaskUpdate to track progress per project and task.

   Launch **Agent(context-manager)** to maintain the note and to orchestrate the technical
   subagents in the next step.

   Loop through tasks in dependency order:

   a. **Scale fan-out to scope — tokens are the budget.** Specialized subagents are expensive, so
      fan out only when the scope is genuinely broad: multiple projects, a large surface, or several
      distinct concerns that benefit from parallel investigation. For a single project, a small
      surface, or a linear analysis you can hold in one pass, investigate **inline** with
      Read/Grep/Glob/Bash instead of spawning an agent. Prefer the smallest unit of work that still
      covers the scope; one well-scoped agent beats a roster for a narrow task.

   b. **Continue until all tasks are done.**
      Each subagent gives its findings and its opinion; context-manager compiles them.

   c. **If a finding requires user input** (unclear context), use the **AskUserQuestion tool** to
      clarify, then continue.

## Choosing subagents

Do not rely on a fixed roster — pick from the agent types available to the `Agent` tool in this
session, so newly added agents are used without this skill changing. If you need to enumerate them,
run `ls .claude/agents/ ~/.claude/agents/`.

**Default for code reading: spawn `code-diver`.** It returns call-graph trees and file:line tables
that anchor the **Current state** and **Impact map** sections — trace how a feature works today, map
a struct's methods, find call sites, scope touched files. Routing rules:

- **code-diver flag.** On a stack-semantic gap (concurrency, framework lifecycle, ORM behavior) it
  flags `→ needs <specialist>`. Spawn that specialist for just that finding.
- **Smart-skip.** Request already semantic (thread-safety, reactivity, ORM/query, framework
  internals) → skip code-diver, spawn the specialist directly.
- **Pass the tree.** Escalating from a flag → hand code-diver's tree/table into the specialist
  spawn as context; it does only the semantic step, no re-trace.

Route the rest by capability: schema and migrations to the database agent, cross-cutting design to
the architecture review agent, test strategy to the QA agent. Reach for a language engineering agent
only when the analysis needs stack-specific depth code-diver cannot provide (concurrency-safety
review, framework internals). If several fit, pick the most specific. If none fits, use
`general-purpose`.

## Output

After all tasks are done, report back to the user in chat:
- The TLDR (what changes, in which projects, rough size and risk)
- The recommended approach and why it beat the alternatives
- Any open question that blocks implementation
- The path to the analysis note

## Note guidelines

The note is a design doc a reviewer should be able to approve or reject without reading the code.
Scale the depth to the change, but never drop the core sections.

### Core sections (always)

- **Frontmatter** — per the note-write skill (tags incl. ticket number, projects, related)
- **TLDR** — what, why, and the verdict (size + risk) in five lines or fewer
- **Context** — ticket acceptance criteria, current behaviour, why now
- **Scope** — in scope / out of scope / assumptions
- **Current state** — per project, how it works today, anchored with `file:line` references
- **Impact map** — table: `project | layer | files | new vs. modify | effort`
- **Proposed change** — per project: API contract, data model and migrations, events, config and
  env vars
- **Flow diagram** — a mermaid diagram (sequence, flowchart, or C4) when the change has moving
  parts worth visualizing: request/data flow across services, component interaction, deploy order,
  or state transitions. Skip for trivial single-file changes — the prose is enough.
- **Alternatives considered** — the options, their tradeoffs, and why the recommended one wins
- **Risks and drawbacks** — table: `risk | impact | mitigation`
- **Feature flag and rollout** — is a flag needed (name, default, kill switch); deploy order across
  repos and services
- **Test strategy** — unit / integration / E2E coverage plus manual QA scenarios
- **Delivery plan** — work breakdown in dependency order, PR split per repo, sizing
- **Open questions** — each with an owner
- **References** — ticket, designs, related notes

### Conditional sections (only when the change triggers them)

Data backfill and migration · backward compatibility and versioning · multi-tenancy · auth and
permission model · i18n · performance and load · security and privacy.

### Cross-cutting checks

Call these out explicitly whenever the change touches them:

- **Generated artifacts**: does an API spec, client, schema type, or mock need regenerating?
- **Shared contracts**: does a shared package or event schema need a version bump, and which
  consumers must follow?
- **Deploy order**: which service must ship before which client, and what deployment config follows

## Guardrails

- If context is critically unclear, ask the user — but prefer reasonable decisions to keep momentum.
- Verify each artifact file exists after writing, before moving to the next.
