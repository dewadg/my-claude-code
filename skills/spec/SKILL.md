---
name: spec
description: >
  Write a lean single-file spec (specs/{slug}.md) that plans a change end-to-end BEFORE any code is
  written — context, goals/non-goals, flow diagram, decisions (ADR-lite), configurations, testing,
  and a dependency-wave todo where every task names the specialized subagent that should
  own it. Use whenever the user wants to spec out, plan, or design a feature, change, or fix before
  implementing it — e.g. "spec out the auth refactor", "plan the new export endpoint", "let's design
  X before we build it", "write a spec for", or just "/spec". Reach for it whenever planning,
  scoping-to-a-decision, or "what would it take to build" intent shows up, even if the word "spec"
  is never used. Also archives a shipped or abandoned spec to specs/archives/{YYYY-MM-DD}-{slug}.md
  — use when the user says "archive the spec", "archive <slug>", "/spec archive", or wants to shelve
  a done spec. Do NOT use for: implementing already-decided work (use /code), investigating an
  unexplained bug to find its cause (use /investigate), or open-ended impact analysis with no
  commitment to a plan (use /analyze).
allowed-tools: Read, Grep, Glob, Bash, Write, Edit, AskUserQuestion, Agent, Skill, WebFetch, WebSearch
---
# Spec

Turn change idea into one build-ready file: `specs/{slug}.md`. Spec commits to decisions + parallelizable plan, so next step just `/code` executing todo. Work order: research, decisions, then writing.

Two jobs: **create** a spec (default, steps below) or **archive** a done one (see Archive). Route on intent — "archive the X spec" jumps straight to Archive.

**Input**: description of what to build or change. May be ticket number, feature name, or prose.

## Steps

1. **If no clear input, ask**

   Use **AskUserQuestion tool** (open-ended, no preset options):
   > "What change do you want to spec? Describe what you want to build or fix."

   Do NOT proceed without understanding intent. From answer, derive kebab-case slug (e.g. "add user authentication" becomes `add-user-auth`).

2. **Ground yourself**

   - Note-search skill available, run it — prior note may already hold analysis or decisions. Don't re-derive what's written down.
   - Read `CLAUDE.md` / `AGENTS.md` of every project the change touches. Project conventions override your defaults, tell you tech stack (which picks subagents later).
   - Slug already exists in `specs/`, read it — ask user whether to continue or overwrite before doing anything.

3. **Research — inline by default, fan out when it pays**

   Default to single inline pass: read key source, existing patterns, contracts the change touches. Keeps token cost low for common case.

   Fan out parallel investigators only when change genuinely large or spans multiple projects — e.g. touches many files, crosses package/repo boundaries, or has unknowns several focused reads resolve faster in parallel. Spawn one `Agent` per project or per subsystem (investigator/Explore-style), each returning short structured summary, then synthesize. If you fan out, say so briefly so user knows the cost.

   Draft **Testing** section inline (pre-test requirements + scenario table, see template). No subagent needed.

4. **Commit to decisions**

   Spec that lists options without picking is analysis, not plan. For every meaningful fork, state choice, reason, alternatives rejected (see Decisions template below). Fork genuinely unresolved, park in Open Questions — but keep that list short.

5. **Write `specs/{slug}.md`**

   `mkdir -p specs/` if needed, then write single file using template below. Fill every applicable section; omit ones that don't apply. Keep prose tight — bullets over paragraphs, one idea per line.

6. **Hand off**

   Summarize in chat: spec path, decisions made (one line each), wave count, any Open Questions needing user. Tell `/code` to honor the checkbox contract under Todo: check off each task `[x]` in the spec as its subagent completes it. End with: "Review the spec, then run `/code` to implement."

## The spec.md template

Use this structure. Section order intentional — context and goals before flow, decisions before work they shape, open questions parked near end, todo last as executable plan.

```markdown
# {Title}

> One-line summary of the change. Link to a ticket/issue only if one already exists.

## Context
Why this change, and the current state. Ground the reader in 2–4 sentences — the problem or
opportunity, not a code dump.

## Goals
- Outcome bullets. What "done" looks like, verifiable where possible.

## Non-goals
- Explicit scope fences — what we are NOT doing, and why it stays out.

## Flow
Mermaid diagram (sequence or flowchart) of the change. Only when the change has a non-trivial flow
across components; otherwise omit the whole section. Example:

\`\`\`mermaid
sequenceDiagram
  participant U as User
  participant S as Server
  U->>S: POST /export
  S-->>U: 202 Accepted (job id)
  S->>S: build file (async)
  S-->>U: SSE progress + result
\`\`\`

## Decisions
ADR-lite. One block per meaningful fork. State the choice first, then why, then what you rejected.

### D1: {decision title}
- **Choice**: what we picked.
- **Why**: the reason it wins here.
- **Alternatives**: what else was considered, rejected in one line each.

## Configurations
Only if the change introduces or updates config, env vars, feature flags, schema, or migrations.
Bullet form — name, default, purpose. Otherwise omit.

## Testing
Only if applicable. Two parts — pre-test requirements (what must hold before tests run), then the
scenario. Write both inline, no subagent.

### Pre-test requirements
- Setup specific to this change that tests depend on: seed data, running services, env vars, applied
  migrations, fixtures. Omit the subsection if there is nothing beyond the project's normal test setup.

### Scenario
Compact Given/When/Then table:

| Given | When | Then |
|-------|------|------|
| visitor has an open conversation | they send a message | server appends a user message and streams an assistant reply |
| conversation is closed (`closedAt` set) | visitor tries to send | composer is locked, message rejected |

## Open Questions
Checkboxes. Keep the section — it records what was unclear and how it resolved. Check `[x]` once
answered and record the answer inline after the item.

- [ ] {unresolved item needing a human decision or follow-up — note what blocks on it}
- [x] {answered item — answer recorded here}

## Todo
A dependency DAG, layered into waves. The steering concern is token cost: every subagent spawn pays a
fixed overhead — its system prompt, context load, and file reads — so parallelism trades tokens for
wall-clock, not the reverse. We parallelize freely when it helps; we just make sure each spawn earns
its overhead.

Size each task to one subagent spawn: the cohesive chunk one agent owns end-to-end, scoped to one
package or project with a clear "done". Group related edits one agent would handle in one spawn into
a single task with sequential sub-bullets — don't split them into parallel spawns that each re-pay
the overhead for trivia. Reserve a parallel wave for tasks that are genuinely independent and each
big enough that the overhead is small relative to the work (typically different projects or different
stacks needing different agents). Every task names the single subagent that owns it (route by
capability — see "Picking subagents" below) and the files/package it lands in.

**Checkbox contract.** The Todo list is the source of truth for progress. Every task starts as `[ ]`.
The instant a task's owning subagent reports its work complete, flip that one checkbox to `[x]` in the
spec file (Edit, not a chat claim) before the next task starts — never leave a finished task unchecked,
never pre-check a task still in flight. `/code` owns this update as it executes each wave; re-read the
spec from disk before each edit so concurrent checks don't clobber. A task is "done" only when the
subagent finished the work AND any test task paired with it is green — not when it merely starts. Before
the spec is archived (Archive step), every box must read `[x]`; any `[ ]` means the change is not done.

Size rule: one subagent spawn per task. Too large (needs multiple spawns or handoffs) → split into
separate tasks. Too small (overhead dwarfs the work, or one agent would do several in one go) →
merge into one.

### Wave 1 — {theme} (parallel: independent projects)
- [ ] [golang-eng] {task} — packages/chatto-server/internal/{pkg}/
  - related sub-step covered by the same spawn
- [ ] [vue-eng] {task} — packages/chatto-ui/src/{dir}/

### Wave 2 — {theme} (depends on Wave 1)
- [ ] [golang-eng] {task that consumes Wave 1 contracts}
- [ ] [qa-eng] write visitor-chat tests — packages/chatto-server/internal/conversation/

### Wave 3 — Verify & review
- [ ] [code-reviewer] review all waves
- [ ] [general-purpose] build + test green (`make build && make test`)
```

## Picking subagents

Do not hardcode a roster — right agent depends on stack, new agents appear over time. Discover what's available this session:

```bash
ls .claude/agents/ ~/.claude/agents/ 2>/dev/null
```

Route by capability, not by name (mirror `/code`'s rule):

- Service / backend: engineering agent for that language (`golang-eng`, `js-eng`, …)
- Web frontend: framework agent (`vue-eng`, `react-eng`, …)
- Test strategy, E2E, browser verification: `qa-eng` / `ui-tester`
- Review of any produced diff: `code-reviewer`
- Containers, compose, images: `docker-eng`
- Schema, migrations, query performance: `db-admin`
- Cross-cutting design / architecture: `architect-reviewer`
- None fits: `general-purpose`

One subagent per project per task — two agents editing same project concurrently conflict. Independent projects (or independent files in one project) can share a wave.

## Lean principles (apply throughout)

- **One file.** If spec feels too long, it's scoping too much — split the change, not the file.
- **Omit, don't pad.** Section that doesn't apply removed entirely, never filled with "N/A" or "TBD". Reader should trust what's there matters.
- **Decide, don't list.** Open Questions is short parking lot, not place to defer every choice.
- **Waves earn their tokens.** Parallel wave runs concurrently but re-pays each spawn's overhead — token cost, not time cost. Parallelize when it helps; just make sure each task in wave independent and big enough to justify that overhead. Otherwise push it down a wave, or merge small same-stack tasks into single spawn.
- **Every behavioral change gets a test task** in same or next wave, whenever project supports tests.
- **Point at source, don't restate.** Reference files and packages by path; don't paste code into spec.

## Archive

Move shipped or abandoned spec out of live set into `specs/archives/`. Trigger on "archive the spec", "archive `<slug>`", "/spec archive", or any shelving intent.

1. **Identify the target.** Slug named in request; if none named, `ls specs/*.md` and ask which (AskUserQuestion). If `specs/{slug}.md` doesn't exist, say so and stop — don't guess.
2. **Check the todo list.** Re-read the spec from disk and scan the Todo section. If every box is `[x]`, it shipped clean — proceed. If unchecked `[ ]` items remain, surface them and ask (AskUserQuestion): is this shipped-but-incomplete or abandoned? For shipped-incomplete, record which tasks were dropped under a new `## Incomplete` heading before moving; for abandoned, record that fact under the heading. Never archive with `[ ]` items unaccounted for — the archive must reflect reality, not a false "done".
3. **Date-stamp.** Use today's date from session context as `YYYY-MM-DD`.
4. **Move it** (not copy):
   ```bash
   mkdir -p specs/archives
   mv specs/{slug}.md specs/archives/{YYYY-MM-DD}-{slug}.md
   ```
5. **Confirm** the destination path.

If `specs/archives/{YYYY-MM-DD}-{slug}.md` already exists, ask overwrite vs skip before replacing.
Archiving removes the spec from `specs/` — it is no longer a live spec.

## Guardrails

- Write exactly one file: `specs/{slug}.md`.
- Do not create branches, tags, commits, or tracker issues unless user asks — planning not an outward-facing action.
- If context critically unclear, use AskUserQuestion — but prefer reasonable committed decision over endless hedging; that's what Open Questions is for.
- Re-read existing spec from disk before editing it; user may have changed it since you last saw it.
- Report faithfully: couldn't verify something during research, say so in Open Questions rather than papering over it.