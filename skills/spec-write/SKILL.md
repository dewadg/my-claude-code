---
name: spec-write
description: >
  Write or search a lean single-file spec at specs/{slug}.md — owns the spec FORMAT: the frontmatter
  schema, the section template (context, goals/non-goals, flow, ADR-lite decisions, configurations,
  testing, open questions, dependency-wave todo), the lean principles, and the search recipes. Use
  whenever the user wants a spec written, formatted, or searched — e.g. "write a spec for", "format
  this as a spec", "find specs about", or just "/spec" (the command invokes this skill). This skill
  does the formatting and the file mechanics only — it does NOT fan out investigator subagents or
  drive the research process; for a change needing multi-project research, the /spec command
  orchestrates the investigators first, then calls this skill to format the result. Do NOT use for:
  implementing already-decided work (use /code), investigating an unexplained bug to find its cause
  (use /investigate), open-ended impact analysis with no commitment to a plan (use /analyze), or
  archiving a shipped or abandoned spec (use spec-archive).
allowed-tools: Read, Grep, Glob, Bash, Write, Edit, AskUserQuestion
---
# Spec

This skill owns the spec **file format** and mechanics: the frontmatter schema, the section template,
the lean principles, and how specs are searched. It does not own the research process — spawning
investigator subagents, tracing code flow, or committing to decisions is the `/spec` command's job —
and it does not own archiving (that is the `spec-archive` skill). When fired directly (not via
`/spec`), assemble the caller's input into the template, ask if intent is unclear, and write the file;
do light grounding reads only, never fan out agents.

Two jobs: **create** a spec (default, below) or **search** prior specs (see Search). Route on intent
— "find specs about Y", "what specs touch Z", "is there already a spec for W", or "/spec search"
jumps to Search. Archiving a done spec ("archive the X spec", "/spec archive") is the `spec-archive`
skill's job, not this one.

**Input**: description of what to build or change, plus any research/decisions the caller already
made. May be a ticket number, feature name, or prose.

## Create

1. **If no clear input, ask**

   Use **AskUserQuestion tool** (open-ended, no preset options):
   > "What change do you want to spec? Describe what you want to build or fix."

   Do NOT proceed without understanding intent. From the answer, derive kebab-case slug (e.g. "add user authentication" becomes `add-user-auth`).

2. **Light grounding (no fan-out)**

   - Search prior specs by topic or project (see Search) — a related or superseded spec may already
     hold decisions worth reusing. Cheap: reads frontmatter only.
   - Read `CLAUDE.md` / `AGENTS.md` of every project the change touches — conventions override
     defaults, and the tech stack tells you which owner agent to name in the Todo.
   - Slug already exists in `specs/`, read it — ask user whether to continue or overwrite before
     doing anything.

   If the change needs real research or multi-project investigation, stop and tell the user to run
   `/spec` — that command fans out investigators and hands the result back here for formatting.

3. **Write `specs/{slug}.md`**

   `mkdir -p specs/` if needed, then write the single file using the template below. Start with the
   frontmatter block: set `status` to `draft`, `created` and `updated` to today, and fill `tags` /
   `projects` / `ticket` / `description`. Fill every applicable section; omit ones that don't apply.
   Keep prose tight — bullets over paragraphs, one idea per line.

4. **Hand off**

   Summarize in chat: spec path, decisions recorded (one line each), wave count, any Open Questions
   needing the user. End with: "Review the spec, then run `/code` to implement."

## The spec.md template

Use this structure. Section order intentional — context and goals before flow, decisions before work
they shape, open questions parked near end, todo last as executable plan.

```markdown
---
status: draft                         # draft | in-progress | shipped | abandoned
created: {YYYY-MM-DD}                 # today
updated: {YYYY-MM-DD}                 # today; bump on any edit
tags: [{topic}, {area}]               # domains this change touches
projects: [{repo-or-package}]         # repos/packages affected
ticket: {TICKET-NUMBER}               # omit the whole line if there is none
related:                              # linked specs, notes, or source files
  - specs/{other-slug}.md
description: {one-line summary of the change — Search reads this, keep it honest}
---

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
stacks needing different agents). Every task names the single subagent that owns it — route by
capability, not a fixed roster (the `/spec` command lists the routing rules) — and the files/package
it lands in.

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

## Lean principles (apply throughout)

- **One file.** If spec feels too long, it's scoping too much — split the change, not the file.
- **Omit, don't pad.** Section that doesn't apply removed entirely, never filled with "N/A" or "TBD". Reader should trust what's there matters.
- **Decide, don't list.** Open Questions is short parking lot, not place to defer every choice.
- **Waves earn their tokens.** Parallel wave runs concurrently but re-pays each spawn's overhead — token cost, not time cost. Parallelize when it helps; just make sure each task in wave independent and big enough to justify that overhead. Otherwise push it down a wave, or merge small same-stack tasks into single spawn.
- **Every behavioral change gets a test task** in same or next wave, whenever project supports tests.
- **Point at source, don't restate.** Reference files and packages by path; don't paste code into spec.
- **Frontmatter stays honest.** `status` tracks the lifecycle (`draft` → `in-progress` when `/code` begins executing waves → `shipped` / `abandoned` on archive) and `updated` bumps on any edit. Search filters on these fields, so stale status or dates misroute lookups.

## Search

Look up prior specs by their frontmatter, not by reading full bodies. Returns a ranked list of matching specs with their frontmatter summary only — pull a full body into context only when a hit is worth acting on.

### Location
`specs/` for live specs; `specs/archives/` for shipped or abandoned ones. **Search both** — archived specs still carry decisions worth recalling. Recursive grep, never a flat `specs/*.md` glob (flat skips `archives/`).

### grep recipes — filter on frontmatter fields
```bash
# by topic / tag
grep -rl "tags: .*auth" specs/

# by affected project / repo
grep -rl "projects: .*chatto-server" specs/

# by lifecycle status: draft | in-progress | shipped | abandoned
grep -rl "status: shipped" specs/

# by ticket
grep -rl "ticket: AUTH-123" specs/

# by date (partial match works)
grep -rl "created: 2026-08" specs/
grep -rl "updated: 2026-08" specs/

# free-text in description
grep -rli "export" specs/
```

Combine filters by piping greps; widen with a bare `grep -rl ... specs/` then eyeball frontmatter on the hits.

### Read frontmatter only, never the whole body
**IMPORTANT**: before reading a candidate spec end-to-end, extract just its frontmatter to decide if it is relevant:
```bash
awk '/^---$/{c++; if(c==2){exit;} next} c==1' specs/{slug}.md
```
Prints the frontmatter block (status, tags, projects, description, dates) and exits at the closing `---` — the body never enters context. Read the full file only when frontmatter confirms the spec is the one you want. Same awk note-search uses on notes.

### Reporting results
List each hit as one line — slug, status, description — pulled from frontmatter. Mark any hit under `archives/` as archived, and rank live specs above archived ones. If a live and an archived spec cover the same topic, prefer the live one and mention the archived as prior context — it may hold the original decisions the live one refined.

## Guardrails

- Write exactly one file: `specs/{slug}.md`.
- Do not create branches, tags, commits, or tracker issues unless user asks — formatting/planning, not an outward-facing action.
- Do not spawn investigator subagents — that is the `/spec` command's role. If research is needed beyond light grounding, defer to `/spec`.
- If context critically unclear, use AskUserQuestion — but prefer reasonable committed decision over endless hedging; that's what Open Questions is for.
- Re-read existing spec from disk before editing it; user may have changed it since you last saw it.
- Report faithfully: couldn't verify something, say so in Open Questions rather than papering over it.
