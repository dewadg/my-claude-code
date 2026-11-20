---
description: Investigate a reported issue across one or more projects — find the root cause and compile findings into an RCA note. Not for scoping a new feature (use /analyze) or implementing a fix (use the code skill).
argument-hint: [issue, symptom, or ticket to investigate]
allowed-tools: Read, Grep, Glob, Bash, Task, SendMessage, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
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

   The main thread maintains the note and launches every subagent itself — subagents cannot spawn
   subagents.

   Loop through tasks in dependency order:

   a. **Scale fan-out to scope — tokens are the budget.** Specialized subagents are expensive, so
      fan out only when the scope is genuinely broad: multiple projects, a large code surface, or
      several distinct hypotheses to chase in parallel. For a single project, a narrow trace, or a
      linear root-cause path you can follow in one pass, investigate **inline** with
      Read/Grep/Glob/Bash instead of spawning an agent. Prefer the smallest unit of work that still
      covers the scope; one well-scoped agent beats a roster for a narrow bug.

   b. **Continue until all tasks are done.** Each subagent gives its findings and its opinion; the
      main thread compiles them into the note.

      The brief is self-contained: a spawned agent starts with no memory of this conversation — goal,
      the files or paths to inspect, what "done" looks like, and what is out of scope. Point at paths
      and symbols; never paste file bodies or diffs into a brief.

      Report contract: findings with `file:line`, commands with exit codes, at most ~15 lines, no
      restated brief, no prose preamble. Every claim in your synthesis traces to the subagent that
      produced it; an unattributable claim does not go in the note. A thin or ambiguous report:
      resume that same subagent (SendMessage) with the specific gap in ~5 lines — never re-run its
      investigation yourself.

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
the read-only explorer agent:

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

**Default for code reading: the read-only explorer agent** (the one whose description is read-only exploration). It returns call-graph trees and file:line tables
that anchor the **Evidence**, **Root cause**, and **Blast radius** sections — trace entry → defect,
map a suspect function's calls, list callers of the buggy symbol. Routing rules:

- **Explorer flag.** On a stack-semantic gap (concurrency, framework lifecycle, ORM behavior) it
  flags `→ needs <specialist>`. Spawn that specialist for just that finding.
- **Smart-skip.** Symptom already points at a semantic cause (race, reactivity, ORM/query,
  framework lifecycle) → skip the explorer, spawn the specialist directly.
- **Pass the tree.** Escalating from a flag → hand the explorer's tree/table into the specialist
  spawn as context; it does only the semantic diagnosis, no re-trace.

Route the rest by capability: reproduction and E2E to the QA or UI testing agent, schema and query
performance to the database agent. Reach for a language engineering agent only when the trace needs
stack-specific depth an explorer cannot provide. If several fit, pick the most specific. If none
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
