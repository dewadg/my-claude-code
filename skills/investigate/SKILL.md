---
name: investigate
description: >
  Investigate an issue across one or more projects, breaking the investigation into tasks delegated
  across specialized subagents and compiling the findings into a root-cause note.
  Use when the user reports something broken, wrong, or unexplained and wants the cause found —
  e.g. "investigate TICKET-123", "why is X returning empty", "users report Y is failing", "find the
  root cause of Z", "this regressed after the last release".
  Not for scoping a new feature (analyze it instead) or for writing the fix itself.
allowed-tools: Read, Grep, Glob, Bash, Task, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---

# Investigate

Break a complex investigation into tasks delegated across subagents. Output is a note.

**Input**: the description of the issue. May be a ticket number, and may name the related projects.

## Steps

1. **If no input provided, ask what issue the user hit**

   Use the **AskUserQuestion tool** (open-ended, no preset options) to ask:
   > "What issue are you currently facing?"

   **IMPORTANT**: Do NOT proceed without understanding the symptom, and which environment it was
   seen in.

2. **Create the investigation note**

   **IMPORTANT**: Use the **note-write skill** and the **note-search skill** to write notes.
   Search the existing notes first — a past investigation may already cover this area.

   If a ticket number is given and an issue-tracker MCP is connected (Jira, Linear, GitHub issues),
   fetch the ticket for context.

   This creates the note that will hold the investigation results.

3. **List impacted projects**

   Take them from the description, or determine them yourself. The goal is to know each impacted
   project's tech stack (Go, React, JavaScript, infra), because that is what selects the subagent.

4. **Investigate**

   Use TaskCreate, TaskGet, TaskList, and TaskUpdate to track progress per project and task.

   Launch **Agent(context-manager)** to maintain the note and to orchestrate the technical
   subagents in the next step.

   Loop through tasks in dependency order:

   a. **Scale fan-out to scope — tokens are the budget.** Specialized subagents are expensive, so
      fan out only when the scope is genuinely broad: multiple projects, a large code surface, or
      several distinct hypotheses to chase in parallel. For a single project, a narrow trace, or a
      linear root-cause path you can follow in one pass, investigate **inline** with
      Read/Grep/Glob/Bash instead of spawning an agent. Prefer the smallest unit of work that still
      covers the scope; one well-scoped agent beats a roster for a narrow bug.

   b. **Continue until all tasks are done.**
      Each subagent gives its findings and its opinion; context-manager compiles them.

   c. **If a finding requires user input** (unclear context), use the **AskUserQuestion tool** to
      clarify, then continue.

   Follow the evidence, not the first plausible story. A cause is only confirmed once the mechanism
   explains every part of the reported symptom. Record the theories you ruled out and why — that is
   what stops the same ground being re-covered later.

## Choosing subagents

Do not rely on a fixed roster — pick from the agent types available to the `Agent` tool in this
session, so newly added agents are used without this skill changing. If you need to enumerate them,
run `ls .claude/agents/ ~/.claude/agents/`.

**Default for code reading: spawn `code-diver`.** It returns call-graph trees and file:line tables
that anchor the **Evidence**, **Root cause**, and **Blast radius** sections — trace entry → defect,
map a suspect function's calls, list callers of the buggy symbol. Routing rules:

- **code-diver flag.** On a stack-semantic gap (concurrency, framework lifecycle, ORM behavior) it
  flags `→ needs <specialist>`. Spawn that specialist for just that finding.
- **Smart-skip.** Symptom already points at a semantic cause (race, reactivity, ORM/query,
  framework lifecycle) → skip code-diver, spawn the specialist directly.
- **Pass the tree.** Escalating from a flag → hand code-diver's tree/table into the specialist
  spawn as context; it does only the semantic diagnosis, no re-trace.

Route the rest by capability: reproduction and E2E to the QA or UI testing agent, schema and query
performance to the database agent. Reach for a language engineering agent only when the trace needs
stack-specific depth code-diver cannot provide. If several fit, pick the most specific. If none
fits, use `general-purpose`.

## Output

After all tasks are done, report back to the user in chat:
- The TLDR (symptom, root cause, blast radius, fix verdict)
- Whether a workaround exists to unblock users now
- Any open question that blocks the fix
- The path to the investigation note

## Note guidelines

The note is a root-cause document. Someone who was not in the investigation should be able to read
it, reproduce the bug, agree with the diagnosis, and ship the fix. Scale the depth to the issue, but
never drop the core sections.

### Core sections (always)

- **Frontmatter** — per the note-write skill (tags incl. ticket number, projects, related)
- **TLDR** — symptom, root cause, blast radius, and fix verdict in five lines or fewer
- **Symptom and impact** — what users see, who is affected, since when, which environments, severity
- **Reproduction** — exact steps to reproduce in a development environment, with the prerequisites
  (data, account, feature flag), plus observed vs. expected behaviour
- **Evidence** — the logs, data, network traces, and code paths that support the diagnosis, anchored
  with `file:line` references
- **Flow diagram** — a mermaid diagram (sequence or flowchart) when the failure crosses components
  or services: the request path from entry to the defect, async/message flow, or the failure
  mechanism step by step. Skip for single-function bugs — the prose is enough.
- **Ruled out** — theories investigated and discarded, each with the reason
- **Root cause** — the single defect: where it lives (`file:line`) and the mechanism by which it
  produces the symptom
- **Regression origin** — the commit, PR, or release that introduced it, and when (if the behaviour
  ever worked)
- **Blast radius** — other call sites, environments, or data affected by the same defect
- **Workaround** — how to unblock users right now (IF POSSIBLE), and its risks
- **Fix options** — table: `option | scope | risk | effort`, with the recommended one and why
- **Verification** — how to prove the fix works, and the regression test to add so it stays fixed
- **Rollout** — hotfix vs. normal release, deploy order across repos, feature flag, any data cleanup
  or backfill the fix requires
- **Open questions** — each with an owner
- **References** — ticket, related notes, dashboards, logs

### Conditional sections (only when the issue triggers them)

Data corruption and backfill · security or privacy exposure · multi-tenancy · performance and load ·
customer communication.

### Cross-cutting checks

Call these out explicitly whenever the issue touches them:

- **Shipped vs. source**: confirm the defect exists on the version actually deployed to the affected
  environment, not only on the development branch
- **Shared contracts**: does the fix require a shared package or event schema bump, and which
  consumers must follow?
- **Deploy order**: which service must ship before which client

## Guardrails

- If context is critically unclear, ask the user — but prefer reasonable decisions to keep momentum.
- Verify each artifact file exists after writing, before moving to the next.
- Do not present a theory as the root cause until the mechanism is confirmed against the code or the
  data. Say plainly when a cause is still unconfirmed.
