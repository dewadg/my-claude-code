---
name: code
description: >
  Coding and programming work, broken into tasks and delegated to specialized subagents chosen by
  each project's tech stack. Works on a single repo or across several projects in a workspace.
  Use when the user wants code written or changed — e.g. "implement TICKET-123", "add endpoint X",
  "fix the bug in Y", "refactor Z", "wire the frontend to the new API".
  Not for diagnosing an unexplained bug (investigate the cause first) or for scoping a feature
  before building it.
allowed-tools: Read, Grep, Glob, Bash, Task, SendMessage, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---
# Code

Break coding instruction into tasks, orchestrate specialized subagents. Output: working, tested, reviewed code.

**Input**: change description. May be ticket number, may name projects.

## Steps

1. **If no input, ask what to work on**

   Use **AskUserQuestion tool** (open-ended, no preset options):
   > "What do you want to work on?"

   **IMPORTANT**: do NOT proceed without understanding what user wants built.

2. **Ground yourself before editing**

   - Ticket number given + issue-tracker MCP connected (Jira, Linear, GitHub issues): fetch ticket for context.
   - Note-search skill available: run it — past note may already carry analysis, root cause, decisions. Do not re-derive what written down.
   - Read `CLAUDE.md` / `AGENTS.md` of every project about to touch. Project conventions there override your defaults.

3. **List impacted projects**

   From description, or determine yourself. Goal: know each impacted project's tech stack (Go, React, JavaScript, infra) — that selects the subagent.

4. **Plan the work**

   Track with TaskCreate, TaskGet, TaskList, TaskUpdate. Single tracking system for this skill — no parallel todo list.

   Decide changes before writing code, record as tasks in dependency order (shared contracts before consumers; backend before frontend calling it). Every behavioural change gets test task alongside when project supports one.

5. **Delegate, then review**

   Loop tasks in dependency order.

   a. **ALWAYS delegate project-level investigation, coding, testing to subagent.** Never edit inline.

   b. **Continue until every task done.**

   c. **Every change reviewed by code-review subagent.** Phases strictly sequential: implementing subagent finishes task, pauses — never keep it running or start review while it works. Review returns findings: **resume same implementing subagent** (SendMessage with name or agent ID, findings verbatim), not fresh spawn — resumed agent applies fixes from context already held; fresh agent re-reads codebase to rebuild context, burns tokens for nothing. Then re-review. Bound at two rounds — findings survive second round, stop, bring to user.

   d. **If something genuinely unclear**: use **AskUserQuestion tool** to clarify, then continue.

## Choosing subagents

No fixed roster — pick from agent types available to `Agent` tool this session, so newly added agents work without skill change. Need enumeration: run `ls .claude/agents/ ~/.claude/agents/`.

Route by capability, not name:

- Service or backend: engineering agent for that language
- Web frontend: engineering agent for that framework
- Test strategy, E2E, browser verification: QA / UI testing agent
- Review of any produced diff: code review agent
- Containers, compose, images: container agent
- Schema, migrations, query performance: database agent
- Cross-cutting design concerns: architecture review agent

Several fit: pick most specific. None fits: `general-purpose`.

One subagent per project per task — two agents editing same project concurrently conflict. Independent projects run parallel.

## Briefing subagents

Give each subagent: goal, files or package to work in, applicable project conventions (that project's `CLAUDE.md` / `AGENTS.md`), what "done" means (build passes, tests pass). Point at source files, not restated spec. Scope to what asked — no opportunistic refactors.

Standing conventions every subagent follows:

- **Tests**: testing skill exists for language (example: Go unit-test skill): use it. Otherwise match existing test style in that project.
- **Comments**: comments, test names, story descriptions explain code behaviour and intent. No review tags, ticket numbers, task ids.
- **Generated code**: never hand-edit generated files (API clients, schema types, mocks, API docs). Change source of truth, re-run generator.

## Verifying

Before reporting done: change must actually build, tests actually pass. Run project build and test commands, read output — never take subagent's word.

UI work: verify running app, not only tests.

No commit, push, tag, or PR unless user asked.

## Output

All tasks done, report back in chat:
- Changes grouped by project, files touched
- Test and build results — quote failures if anything failed
- Anything undone, deferred, knowingly out of scope
- Follow-ups user must action (migration to run, config or env var to set, deploy order across repos)

## Guardrails

- Context critically unclear: ask user — but prefer reasonable decisions, keep momentum.
- Verify each artifact file exists after writing, before next.
- Report outcomes faithfully. Tests fail: say so with output. Never claim done on unverified work.