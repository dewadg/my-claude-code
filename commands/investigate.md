---
description: Investigate a reported issue across one or more projects — break it into delegated tasks, find the root cause, compile findings into a note.
argument-hint: [issue, symptom, or ticket to investigate]
allowed-tools: Read, Grep, Glob, Bash, Task, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---

# Investigate

Break a complex investigation into tasks delegated across subagents. Output is a root-cause note.

**Input**: `$ARGUMENTS` — the description of the issue. May be a ticket number, and may name the
related projects.

## Steps

1. **If `$ARGUMENTS` is empty, ask what issue the user hit**

   Use the **AskUserQuestion tool** (open-ended, no preset options) to ask:
   > "What issue are you currently facing?"

   **IMPORTANT**: Do NOT proceed without understanding the symptom, and which environment it was
   seen in.

2. **Create the investigation note**

   **IMPORTANT**: Use the **note-search skill** first — a past investigation may already cover this
   area.

   Then invoke the **rca-write skill** to create the RCA note. rca-write owns and enforces the RCA
   note format (section template, conditional sections, cross-cutting checks) and builds on the
   note-write skill's frontmatter and filename conventions. This creates the note skeleton — every
   core section header in place — that will hold the investigation results.

   If a ticket number is given and an issue-tracker MCP is connected (Jira, Linear, GitHub issues),
   fetch the ticket for context.

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

## Explaining code flow — use the call-graph skill

Whenever you explain how code flows — the path from entry point to the defect, a suspect function's
calls, who calls a symbol, a request's journey through layers — render it with the **call-graph
skill**, not prose. Invoke the skill via the **Skill tool** and emit its indented arrow tree (one
node per call hop, each carrying `file:line`). This applies whether you trace inline or hand off to
`code-diver`:

- The note's **Evidence** section anchors on these trees/tables — the code path from entry to defect
  is the tree.
- The note's **Flow diagram** uses the call-graph tree for the failure's code path. (A mermaid
  sequence/flowchart is still right when the failure crosses components or services — async/message
  flow, request path across services — pick by what is being shown, but a code call path is always
  the tree.)
- The chat TLDR references the same tree when it needs to show the mechanism.

Do not paraphrase a traced path as prose when the call-graph tree can carry it. The tree is the
format for code flow.

## Choosing subagents

Do not rely on a fixed roster — pick from the agent types available to the `Agent` tool in this
session, so newly added agents are used without this command changing. If you need to enumerate
them, run `ls .claude/agents/ ~/.claude/agents/`.

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

## Guardrails

- If context is critically unclear, ask the user — but prefer reasonable decisions to keep momentum.
- Verify each artifact file exists after writing, before moving to the next.
- Do not present a theory as the root cause until the mechanism is confirmed against the code or the
  data. Say plainly when a cause is still unconfirmed.
