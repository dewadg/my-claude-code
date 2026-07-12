---
name: code
description: >
  Coding and programming work, broken into tasks and delegated to specialized subagents chosen by
  each project's tech stack. Works on a single repo or across several projects in a workspace.
  Use when the user wants code written or changed — e.g. "implement TICKET-123", "add endpoint X",
  "fix the bug in Y", "refactor Z", "wire the frontend to the new API".
  Not for diagnosing an unexplained bug (investigate the cause first) or for scoping a feature
  before building it.
allowed-tools: Read, Grep, Glob, Bash, Task, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---

# Code

Break a coding instruction into tasks and orchestrate specialized subagents. Output is working,
tested, reviewed code.

**Input**: the description of the change. May be a ticket number, and may name the projects.

## Steps

1. **If no input provided, ask what to work on**

   Use the **AskUserQuestion tool** (open-ended, no preset options) to ask:
   > "What do you want to work on?"

   **IMPORTANT**: Do NOT proceed without understanding what the user wants to build.

2. **Ground yourself before editing**

   - If a ticket number is given and an issue-tracker MCP is connected (Jira, Linear, GitHub
     issues), fetch the ticket for context.
   - If a note-search skill is available, run it — a past note may already carry the analysis, the
     root cause, or the decisions for this work. Do not re-derive what is already written down.
   - Read the `CLAUDE.md` / `AGENTS.md` of every project you are about to touch. Project
     conventions live there and they override your defaults.

3. **List impacted projects**

   Take them from the description, or determine them yourself. The goal is to know each impacted
   project's tech stack (Go, React, JavaScript, infra), because that is what selects the subagent.

4. **Plan the work**

   Use TaskCreate, TaskGet, TaskList, and TaskUpdate to track progress. This is the single tracking
   system for this skill — do not also keep a parallel todo list.

   Decide what changes are needed before writing any code, and record them as tasks in dependency
   order (shared contracts before consumers; backend before the frontend that calls it). Every
   behavioural change gets a test task alongside it whenever the project supports one.

5. **Delegate, then review**

   Loop through tasks in dependency order.

   a. **ALWAYS delegate project-level investigation, coding, and testing to a subagent.**
      Do not make the edits inline.

   b. **Continue until every task is done.**

   c. **Every change is reviewed by a code-review subagent.** If the review returns findings, route
      them back to the implementing subagent and re-review. Bound this at two rounds — if findings
      survive a second round, stop and bring them to the user rather than looping.

   d. **If something is genuinely unclear**, use the **AskUserQuestion tool** to clarify, then
      continue.

## Choosing subagents

Do not rely on a fixed roster — pick from the agent types available to the `Agent` tool in this
session, so newly added agents are used without this skill changing. If you need to enumerate them,
run `ls .claude/agents/ ~/.claude/agents/`.

Route by capability, not by name:

- A service or backend → the engineering agent for that language
- A web frontend → the engineering agent for that framework
- Test strategy, E2E, browser verification → the QA / UI testing agent
- Review of any produced diff → the code review agent
- Containers, compose, images → the container agent
- Schema, migrations, query performance → the database agent
- Cross-cutting design concerns → the architecture review agent

If several agents fit, pick the most specific one. If none fits, use `general-purpose`.

One subagent per project per task — two agents editing the same project concurrently will conflict.
Independent projects can run in parallel.

## Briefing subagents

Give each subagent: the goal, the files or package to work in, the project conventions that apply
(from that project's `CLAUDE.md` / `AGENTS.md`), and what "done" means (build passes, tests pass).
Point them at source files rather than restating the spec inline. Scope them to what was asked — no
opportunistic refactors.

Standing conventions every subagent must follow:

- **Tests**: if a testing skill exists for the language (for example a Go unit-test skill), use it.
  Otherwise match the existing test style in whatever project they are in.
- **Comments**: comments, test names, and story descriptions explain the code's behaviour and
  intent. No review tags, no ticket numbers, no task ids.
- **Generated code**: never hand-edit generated files (API clients, schema types, mocks, API docs).
  Change the source of truth and re-run the generator.

## Verifying

Before reporting done, the change must actually build and its tests must actually pass. Run the
project's build and test commands and read the output — do not take a subagent's word for it.

For UI work, verify the running app, not only the tests.

Do not commit, push, tag, or open a PR unless the user asked for it.

## Output

After all tasks are done, report back to the user in chat:
- The changes, grouped by project, with the files touched
- Test and build results — quote the failures if anything failed
- Anything left undone, deferred, or knowingly out of scope
- Follow-ups the user needs to action (a migration to run, a config or env var to set, a deploy
  order across repos)

## Guardrails

- If context is critically unclear, ask the user — but prefer reasonable decisions to keep momentum.
- Verify each artifact file exists after writing, before moving to the next.
- Report outcomes faithfully. If tests fail, say so with the output. Never claim done on unverified
  work.
