---
name: breakdown
description: >
  Break a big block of requirements — a PRD, a pasted requirements doc, a vague or oversized
  ticket, an epic, a Figma flow — into tickets human engineers can pick up (backend, frontend,
  mobile, and the other seams: migrations, jobs, infra, analytics, admin), each sized in human
  working time (t-shirt), framed by business impact (via the business-impact skill), prioritized,
  and sequenced. Also grooms an
  existing backlog. Orchestrates: maps the codebase with code explorers, cuts tickets per seam,
  writes each body with the ticket-description skill, and creates them via glab/gh only after
  confirmation. Use whenever the user wants requirements or a ticket turned into multiple tickets —
  e.g. "break this down into tickets", "split #42", "this ticket is too big", "slice this PRD into
  BE/FE/mobile tickets", "estimate this feature", "which of these should we do first", "groom the
  backlog for milestone Q4", or /breakdown. Not for a single ticket's description with no breakdown
  (use ticket-description), technical impact analysis with no tickets out (use /analyze), or
  building the work (use code).
argument-hint: "[requirements file/URL, ticket id/URL, or pasted text] [--groom] [--write]"
allowed-tools: Read, Grep, Glob, Bash, Write, WebFetch, Artifact, Task, SendMessage, AskUserQuestion, TaskCreate, TaskGet, TaskList, TaskUpdate, Skill
---

# Breakdown

Turn a big block of requirements into tickets a **human** engineer can pick up on Monday without a
meeting: one seam, one deliverable, sized in human effort, justified by business value, ordered.

You play the senior product manager next to an engineering lead who is strong technically and
weaker on business framing. Carry the "why" and the ordering, and explain it in plain language.

**Input**: `$ARGUMENTS` — a requirements file path or URL, a ticket/epic id or URL (GitLab `#42` /
`group/project#42`, GitHub `#42`), or pasted text. May name projects.
`--groom` — the target is a backlog (label, milestone, epic) to groom; see `references/grooming.md`.
`--write` — create the tickets in the tracker after the user approves them.

## Reference files — read when you reach the step

| File | Read at |
|---|---|
| `references/seams.md` | steps 4–5 |
| `references/estimation.md` | step 6 |
| `references/prioritization.md` | step 7 |
| `references/grooming.md` | `--groom` only |
| `references/tracker-cli.md` | steps 1 and 9, whenever touching the tracker |

## Workflow

Track steps with TaskCreate/TaskUpdate — a big block has many moving parts.

### 1. Ingest

- `$ARGUMENTS` empty → ask with AskUserQuestion: "What should I break down? (file, URL, ticket, or
  paste it)".
- File → Read. URL → WebFetch (or the tracker CLI if it's a ticket URL). Ticket/epic → fetch with
  `glab`/`gh` including comments, linked issues, and parent (comments often hold decisions the
  description never caught up with). Figma link → keep it for the ticket bodies; fetch only if the
  text doesn't describe the screens. Claude Design canvas (a claude.ai Design artifact, by link or
  name) → invoke the **design-shot skill** to list its screens (`--list`); the screen inventory
  feeds step 5's FE/mobile tickets.
- Use the **note-search skill** for prior design docs/RCAs on the same area.

### 2. Decompose into requirements

Before any ticket, turn the block into a numbered list of **requirements** `R1…Rn`: one
user-visible or system outcome each, quoting the source line it came from. Then list:

- **Ambiguities** — anything the block leaves open (empty states, permissions, existing behaviour
  changing, platform parity web vs mobile).
- **Implied work** the block doesn't say but needs (migration, notification, admin toggle).
- **Out of scope** — what a reader might assume is in but isn't.

Every ticket later traces to an `R#`; that is how nothing falls through.

### 3. Frame the business impact — checkpoint 1

Invoke the **business-impact skill** with the source and the `R#` list. It returns the framing
block (problem & who, value type, metric, cost of delay, smallest valuable slice, confidence) and a
**Core / Supporting / Speculative** tag per `R#` — step 7 scores Value from those tags, and the
smallest valuable slice becomes the Must bucket.

Show the user the requirements, the framing, and the blocking questions in one AskUserQuestion
batch. **Small block (≤ ~5 requirements, no blocking questions) → skip the stop and continue.**

### 4. Map the code

Identify the projects behind each seam (backend, web, mobile, infra) and read each project's
`CLAUDE.md` / `AGENTS.md`. Greenfield → skip mapping, mark sizes as greenfield.

**Scale fan-out to scope — tokens are the budget.** One small project → inventory inline with
Read/Grep/Glob. Several projects or a large surface → spawn **`code-diver`** once per project, in
parallel, with the seam brief from `references/seams.md` (endpoint inventory for BE, screen and
shared-component inventory for FE/mobile, etc.). Ask for `file:line` tables, not prose. Stop at
inventory — what exists, what's reusable, what's touched — not implementation design.

When code-diver flags `→ needs <specialist>`, spawn that specialist only for that finding, passing
the tree.

### 5. Cut tickets per seam

Per `references/seams.md`: BE per endpoint (+ groundwork tickets), FE/mobile per screen (shared or
large components separate), plus the cross-cutting seams that are real for this block. Each ticket:
one seam, one engineer, one mergeable deliverable, ≤ L, traced to `R#`s. Contract-first so FE/mobile
start in parallel. Spikes for genuine unknowns.

**Coverage check** — every `R#` lands in ≥ 1 ticket; list any that don't as gaps.

### 6. Estimate

Per `references/estimation.md`: t-shirt in **human** effort, with the size driver tied to code
evidence. Unknown that swings the size → `risk: high` + a spike, not an inflated size. If a size
hinges on code you didn't see, run the trace (step 4 rules) before guessing.

### 7. Prioritize & sequence

Per `references/prioritization.md`: hard constraints first, then WSJF-lite score, Must/Should/Could
buckets, per-seam lanes with critical path, cut line if capacity is known. Flag judgment calls as
the user's decision.

### 8. Present — checkpoint 2

Show the **plan**, not bodies yet:

1. Business framing (from step 3, updated)
2. Ticket table — `# | Title | Seam | Size | R# | Blocked by | Priority (bucket, score)`
3. Sequence lanes + critical path
4. Effort per lane, and the rough calendar estimate if team size is known
5. Gaps, open questions, assumptions, size risks

Let the user adjust (merge, split, resize, reorder) before writing bodies — bodies are the
expensive part, don't write 20 then rework them.

### 9. Write ticket bodies, then (optionally) the tracker

- Write each body with the **ticket-description skill** (story / bug / task shape). Its rules hold:
  never invent requirements, `> TODO: confirm …` for minor unknowns. Each body carries its `R#`s
  and `Parent: #<id>` when splitting a ticket, and only the slice of parent ACs it satisfies.
- FE/mobile tickets with a Claude Design source: invoke the **design-shot skill** per ticket for
  the screen (and the states its ACs name — error, loading, empty) and, where the ticket covers only
  part of a screen, a crop of that part. Each shot becomes a row of the ticket's **Design** section
  (ticket-description owns the table; design-shot supplies the rows, image embedded once uploaded).
- Save bodies to the scratchpad (one file per ticket) so they can be passed by path.
- `--write` → confirm first:
  > "Create N tickets in <project>, linked to #<parent>? (parent description unchanged)"
  On explicit yes, create via `references/tracker-cli.md`, link to the parent, report every URL.
  Design screenshots upload with it (GitLab, via design-shot's upload step) — mention the image
  count in the confirmation.
  Updating the parent (child checklist) needs its own yes.
- No `--write` → print bodies in fenced blocks and offer creation in one line.

## Guardrails

- Never write to the tracker without `--write` **and** an explicit yes; never overwrite an existing
  description without separate confirmation.
- Never invent requirements, metrics, deadlines, or sizes without a driver — trace to a source or
  mark unknown.
- Estimates are human effort; never shrink a size because an AI could write the code faster.
- Context critically unclear (no user, no goal) → ask; otherwise prefer reasonable decisions and
  list them as assumptions.
- Calibration: when the user reports actual vs estimated for shipped tickets, save the gap as a
  memory so later estimates in this project adjust.
