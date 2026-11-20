---
name: rca-write
description: >
  Compile investigation findings into an RCA note: TLDR, symptom, reproduction, evidence,
  call-graph flow, ruled-out theories, root cause, blast radius, fix options, verification,
  rollout. Use for "write the RCA", "compile the investigation into a note", or when /investigate
  hands off. Owns the RCA body template only. Not for scoping a feature or design docs (/analyze,
  design-write), a spec (spec-write), or generic notes (note-write).
allowed-tools: Read, Grep, Glob, Write, Edit, AskUserQuestion
---
# RCA note

Owns the RCA note **body format**: section template, conditional sections, cross-cutting checks,
writing principles. Frontmatter, filename and location are **note-write's** — don't redefine them
here, note-search greps those fields uniformly across every note. The investigation (tracing code,
confirming a cause) belongs to `/investigate`; fired directly, assemble the caller's findings, ask
if intent is unclear, and write the note — light grounding reads only, never fan out agents.

**Input**: the symptom plus any findings in hand (ticket, suspected cause, logs, traced paths).

## Create

1. No clear input → AskUserQuestion, open-ended: "What issue should the RCA cover? Symptom, affected
   environment, findings so far." Derive the slug (and ticket number) from the answer.
2. Prior notes → **note-search**; related ones go in `related`.
3. Write per note-write: frontmatter, then every core section below. Conditional sections only
   when they apply. The full skeleton at creation is what enforces the format; the investigation
   fills it in place.
4. Hand off in chat: path, the root cause if already confirmed, and "Review the RCA; the remaining
   sections fill in as the investigation progresses."

## The RCA note template

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
  with `file:line` references and the call-graph tree from entry to defect
- **Flow diagram** — a call-graph tree for the failure's code path; a mermaid diagram (sequence or
  flowchart) when the failure crosses components or services: the request path from entry to the
  defect, async/message flow, or the failure mechanism step by step. Skip for single-function bugs
  — the prose is enough.
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

## Code flow

**Evidence** and **Flow diagram** render each traced path as a call-graph tree (per the
**call-graph skill**: indented arrows, one hop per line, `file:line` per node), never prose.
`/investigate` decides when to trace; this skill fixes only how the trace lands in the note.

## Guardrails

- One note file. No subagents — real investigation defers to `/investigate`.
- No theory as root cause until the mechanism is confirmed against code or data; an unconfirmed
  cause goes in Open Questions, stated plainly.
- Re-read the note from disk before editing; the investigation may have changed it.
