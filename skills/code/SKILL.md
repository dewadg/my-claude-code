---
name: code
description: >
  Write or change code in one repo or several: break the work into tasks, delegate to the
  project's own stack agents, then verify against its gates; journals multi-task runs. Use for
  "implement TICKET-123", "add endpoint X", "fix the bug in Y", "refactor Z", "wire the frontend to
  the new API". Not for diagnosing an unexplained bug (investigate its cause first) or scoping a
  feature before building it.
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, Task, SendMessage, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---
# Code

Break coding instruction into tasks, orchestrate specialized subagents. Output: working code,
verified against the project's own gates, and — for larger runs — a session journal.

**Input**: change description. May be ticket number, may name projects.

The main thread is the orchestrator: it plans, briefs, verifies returned reports, and writes the
journal. It never writes product code, runs builds or tests, or drives a browser — subagents cannot
spawn subagents, so it launches every one itself.

## Steps

1. **If no input, ask what to work on**

   Use **AskUserQuestion tool** (open-ended, no preset options):
   > "What do you want to work on?"

   **IMPORTANT**: do NOT proceed without understanding what user wants built.

2. **List impacted projects**

   From description, or determine yourself. Goal: know each impacted project's tech stack (Go,
   React, JavaScript, infra) — that selects the subagent. Read each project's `CLAUDE.md` /
   `AGENTS.md`, README and CI config: they define commands, gates, branch convention. An
   undiscoverable convention: ask once, record it as a decision `D<n>`.

   **Pre-flight recon.** When the change spans more than one project, or has more than two tasks,
   launch one read-only explorer subagent per project, in parallel. Its single deliverable is a
   recon digest at `<session dir>/recon-{project}.md` (see **Session directory**): the entry points
   the change touches, the project's conventions, the narrow and full test/build commands,
   cross-project contract touch points, and a `file:line` index. Every later brief cites that file
   and does not restate it. A single-project, small change skips recon.

3. **Plan the work**

   Track with TaskCreate, TaskGet, TaskList, TaskUpdate — single tracking system, no parallel todo
   list. More than two tasks or more than one project: also open the session journal (see
   **Session journal**); briefs cite its section anchors — no separate plan file.

   Decide changes before writing code, record as tasks in dependency order (shared contracts before
   consumers; backend before frontend calling it). Every behavioral change gets a test task
   alongside when the project supports one.

4. **Delegate, then verify and review once for the whole change**

   Loop implementation tasks in dependency order.

   a. **ALWAYS delegate project-level investigation, coding, testing, and browser verification to a
      subagent** with the `Agent` tool. Never edit, build, test, or drive a browser inline. Keep the
      returned agent ID for each task. Implementing subagents write their full record to
      `<session dir>/evidence-{project}-{agent}.md`. Every brief states that file changes go through
      the Edit tool — `sed`/`awk` only for a uniform multi-file rewrite, never a Python script or
      `python3` heredoc patching files.

   b. **Concurrency.** One subagent per project per task — two agents in the same project collide on
      the test database, dev-server port, build output, lockfile, or migration state. Parallel is for
      independent projects and for read-only work (any number of explorers). Serialize anything that
      mutates a shared resource inside one project: builds, full suites, dev servers, migrations,
      generated artifacts.

   c. **Order the work, then close each unit.** What must pass before a unit is done is the
      project's own call — its `CLAUDE.md`/`AGENTS.md`, CI config, `Makefile` or scripts, existing
      practice. Unless the project defines its own gates, the default is: the implementing subagent
      self-checks narrow (build, lint, unit tests for what it touched), then the project build and
      suite run **once** per project, by resuming that project's last implementing subagent after
      the others finish; then one batch review; then behavioral verification last. Builds and suites
      belong to the implementing subagent, never the QA agent.

   d. **Batch every review.** Independent tasks run parallel; each subagent pauses when done. When the
      project defines a review convention, follow it; otherwise review is batch-level, never
      per-subagent or per-project — separate reviews burn duplicate rounds and miss cross-boundary
      mismatches (frontend calling backend contract).

   e. **One review pass: a single review-agent subagent, single invocation, scope not payload.** The
      brief carries each repo root, the output of `git diff --stat`, the acceptance criteria
      reference, an exclusion list, and an instruction to run `git diff` itself. NEVER paste diff
      text into a brief. Exclusion list, unless the excluded file is itself the change's source of
      truth: lockfiles, generated code (API clients, schema types, mocks, API docs), vendored
      directories, minified bundles, snapshots, `.notes/`. If the combined diff exceeds roughly 1500
      changed lines, review it in directory slices sequentially inside that one reviewer session,
      reporting once at the end.

   f. Review returns findings: split by project, **resume the same implementing subagents**
      (SendMessage with the saved agent ID, findings verbatim, each its slice only), not fresh
      spawns — resumed agents apply fixes from context already held. Fixes for different projects run
      parallel; same-project findings go to that project's agent in one message. Each fix self-checks
      narrow — no full-suite rerun for review fixes unless a fix crosses package boundaries, and then
      the project's last implementing agent reruns the full suite once.

   g. All fixes in: re-review the full combined diff once, same way; when the project defines no
      review convention, bound it at two rounds — findings surviving the second pass stop and come to
      the user.

   h. **Behavioral verification, where the project calls for it, once, last.** One brief to one QA
      subagent covering every touched project: verification on the running app against the
      acceptance criteria. QA defects loop back exactly like review findings — resume the project's
      implementing agent with the defect verbatim, it fixes and narrow re-checks, then resume the same
      QA agent with the failed scenario IDs only — never the whole pass, never a fresh full brief.
      Defects surviving one fix round: stop, bring to user.

   i. **If something genuinely unclear**: use **AskUserQuestion tool** to clarify, then continue.
      Reversible, in-scope choices: decide, record `D<n>`, keep momentum.

## Choosing subagents

No fixed roster — pick from agent types available to the `Agent` tool this session, so a newly
added agent works without a skill change. Need enumeration: `ls .claude/agents/ ~/.claude/agents/`
once per session; keep the result in the plan.

Route by capability, not name: backend → that language's engineering agent; web frontend → that
framework's engineering agent; test strategy, E2E, browser verification → QA / UI testing agent;
review of a produced diff → code review agent; tracing or symbol lookup → read-only explorer;
containers, compose, images → container agent; schema, migrations, query performance → database
agent; cross-cutting design → architecture review agent. Most specific fit wins; none fits →
`general-purpose`. Any diff to source, tests, build config or migrations, however small, is a
subagent's job.

## Briefing subagents

A spawned agent starts with no memory of this conversation. Give each: goal, files or package to
work in, the search scope to confine itself to (grep there rather than read wide), applicable project
conventions, what "done" means (the narrow build, lint, and unit commands for its packages pass, and
the behavior under test is covered), and its evidence file path. Point at source files, acceptance
criteria, and the recon digest — not restated spec. Scope to what asked — no opportunistic
refactors.

Every brief carries this hard report contract:

- **Files touched**: paths only.
- **Commands**: `cmd` → exit code, one line each.
- **Failures**: failing test/step IDs plus the assertion line, at most ~10 lines; the full log goes
  to a file path, never into the report.
- **Not verified**: one line each, with the reason.
- At most ~15 lines total; the full record goes to the evidence file.
- No pasted diffs, no file bodies, no restated brief or plan.

Over roughly 200 lines back means the brief was too broad — narrow the next brief rather than
accepting the dump.

Standing conventions every subagent follows:

- **Tests**: a testing skill exists for the language — `golang-unit-test` for Go — so use it.
  Otherwise match existing test style in that project.
- **Comments**: comments, test names, story descriptions explain code behavior and intent. No review
  tags, ticket numbers, task ids.
- **Generated code**: never hand-edit generated files (API clients, schema types, mocks, API docs).
  Change source of truth, re-run generator.

## Session directory

Everything a journaled run produces goes into one directory:

`.notes/code/{YYYY}-{MM}-{DD}-{slug}/`, or `{YYYY}-{MM}-{DD}-{TICKET}-{slug}/` when a ticket is
known — date prefix always, ticket segment before the slug. Notes root follows the project's
`CLAUDE.md` if it names one; the layout still applies.

- `{stem}-journal.md` — the run's journal, sharing the directory's stem.
- `recon-{project}.md` — one per project, from pre-flight recon.
- `evidence-{project}-{agent}.md` — one per subagent, one writer per file, never shared between
  parallel implementers. The ~15-line cap binds the message back, not the record.
- `review-round-{n}.md` — the findings of review round `n`.
- `assets/` — screenshots and other non-markdown evidence.

`related:` entries are sibling-relative. Archive a finished session by one `mv` to
`.notes/code/archive/{same dir name}/` — the stem moves with it, so links stay valid. `.notes/` never
enters a commit and is outside review scope.

## Session journal

Written for runs with more than two tasks or more than one project; the main thread is its only
writer, transcribing subagents' report rows and evidence paths. The review agent and QA agent report
in chat; their findings are transcribed here. Content is append-only — rows accrue, `State` cells and
`updated:` change — and the file is rewritten whole at each boundary (Write), not patched cell by
cell. Write at natural boundaries: plan set, unit closed, run closed. Read it at resume boundaries
and before the final report.

Ids are numbered within the session: `D<n>` decisions, `R<round>.F<n>` review findings (`n`
sequential within the round), `Q<n>` QA defects (sequential across every QA pass, in report order —
never restarted per pass).

```markdown
---
tags: [code]
projects: [<project>]
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
related: [recon-<project>.md, evidence-<project>-<agent>.md]
description: <change> — session journal
---
# <change title>
<intent, two lines> · status: shipped | partial | blocked
## Status
| Task | Project | Owner | State | Evidence |
|---|---|---|---|---|
| T1 | <project> | <agent> | done | evidence-<project>-<agent>.md |
## Decisions
| # | Decision | Rejected | Why |
|---|---|---|---|
| D1 | <choice> | <alt> | <why> |
## Coordination log
| When | Agent ID | Agent | Project | Task | Brief |
|---|---|---|---|---|---|
| plan | <id> | <agent> | <project> | T1 | <anchor> |
## Findings ledger
| Round | ID | Tier | Location | Owner | State |
|---|---|---|---|---|---|
| 1 | R1.F1 | 🔴 bug | <path>:<line> | <agent> | fixed |
## Outcome
<what shipped, commits, what was deferred, user follow-ups>
```

A PR body composes from `Outcome` and the findings ledger.

## Context discipline

Each tool call is a turn and every turn re-reads the whole context, so output size and turn count,
not report length, dominate cost.

- Cap what comes back: reports ~15 lines, the full record in the evidence file. Cite `path:line` and
  section anchors instead of restating.
- Hand back paths, not payloads — screenshots, HTML and full logs stay on disk.
- Resume a saved agent (SendMessage) instead of respawning: it already holds its context.
- Search before reading: Grep/Glob to locate, then Read with `offset`/`limit`; never read a whole
  file over ~500 lines.
- Escalate cheaply: `git diff --stat` before `git diff`, `--name-only` before per-file diffs,
  `head`/`tail` on logs, `-q` on noisy commands.
- Batch — one grep over the tree beats five sequential greps, one wider Read beats three.
- Never echo tool output into the reply or re-read a file you just wrote.
- Inspection across many files (parsing a corpus, diffing two trees, querying a database) is work,
  not reading: brief an explorer with the question.

## Verifying

Verify the returned work by reading; never reproduce it.

- Every claim traces to the subagent that produced it; an unattributable claim did not happen.
- Every criterion in the brief maps to a reported result; commands carry exit codes and quoted
  output — "should pass" is not a result.
- A missing, thin or contradicted item is a gap: resume that same subagent with the specific gap in
  ~5 lines, or report it — never close it by running the build, tests, or browser yourself.
- A project's suite run, where it has one, is attributed to the implementing subagent that ran it.
- No commit, push, tag, or PR unless the user asked. When asked: stage by explicit path (never
  `git add -A` / `git add .`), `git diff --cached --stat` matches planned scope, `.notes/` never
  enters a commit, bodies come from reports — never an invented result.

## Output

All tasks done, report back in chat, bad news first:
- Outcome: one line — done, or blocked on what
- Changes grouped by project, files touched, the agent behind each
- Verification per task: evidence present, missing, or contradicted — quote failures, exit codes
- Decisions taken for the user (`D<n>`) that were not obvious
- Anything undone, deferred, knowingly out of scope
- Follow-ups user must action (migration to run, config or env var to set, rollout order across
  repos), and the journal path when one was written

## Guardrails

- Context critically unclear: ask user — but prefer reasonable decisions, keep momentum.
- Verify each artifact file exists after writing, before next.
- Report outcomes faithfully. Tests fail: say so with output. Never claim done on unverified work.
