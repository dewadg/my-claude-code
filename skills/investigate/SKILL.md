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

   a. **ALWAYS launch specialized subagents for project-level codebase investigation.**
      Do not investigate inline.

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

Route by capability: a service or backend goes to the engineering agent for its language, a web
frontend to the agent for its framework, reproduction and E2E to the QA or UI testing agent, schema
and query performance to the database agent. If several fit, pick the most specific. If none fits,
use `general-purpose`.

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
