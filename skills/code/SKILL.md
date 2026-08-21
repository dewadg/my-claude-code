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

2. **List impacted projects**

   From description, or determine yourself. Goal: know each impacted project's tech stack (Go, React, JavaScript, infra) — that selects the subagent.

3. **Plan the work**

   Track with TaskCreate, TaskGet, TaskList, TaskUpdate. Single tracking system for this skill — no parallel todo list.

   Decide changes before writing code, record as tasks in dependency order (shared contracts before consumers; backend before frontend calling it). Every behavioral change gets test task alongside when project supports one.

4. **Delegate, then review once for the whole change**

   Loop implementation tasks in dependency order.

   a. **ALWAYS delegate project-level investigation, coding, testing to subagent.** Never edit inline.

   b. **No review until every implementing subagent finished.** Independent tasks run parallel; each subagent pauses when done. Review is batch-level, never per-subagent or per-project — separate reviews burn duplicate rounds and miss cross-boundary mismatches (frontend calling backend contract).

   c. **One review pass: single code-review subagent, single invocation, full combined diff across every touched project.** Reviewer gets the whole working-tree diff of all touched repos in one briefing, not one diff per agent.

   d. Review returns findings: split by project, **resume the same implementing subagents** (SendMessage with name or agent ID, findings verbatim, each its slice only), not fresh spawns — resumed agents apply fixes from context already held. Fixes for different projects run parallel; same-project findings go to that project's agent in one message.

   e. All fixes in: re-review the full combined diff once, same way. Bound at two rounds total — findings survive the second pass, stop, bring to user.

   f. **If something genuinely unclear**: use **AskUserQuestion tool** to clarify, then continue.

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

Give each subagent: goal, files or package to work in, applicable project conventions, what "done" means (build passes, tests pass). Point at source files, not restated spec. Scope to what asked — no opportunistic refactors.

Standing conventions every subagent follows:

- **Tests**: testing skill exists for language (example: Go unit-test skill): use it. Otherwise match existing test style in that project.
- **Comments**: comments, test names, story descriptions explain code behavior and intent. No review tags, ticket numbers, task ids.
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