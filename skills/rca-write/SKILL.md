---
name: rca-write
description: >
  Compile investigation findings into a root-cause analysis (RCA) note — owns the RCA note FORMAT:
  the section template, conditional sections, cross-cutting checks, and writing principles. Built on
  the note-write skill's frontmatter + filename + location conventions — an RCA note IS a note. Use
  whenever the user wants an RCA written, formatted, or compiled — e.g. "write the RCA", "compile
  the investigation into a note", "format this as a root-cause doc", or when the /investigate command
  hands off findings. This skill formats and writes the note only — it does NOT fan out investigator
  subagents or drive the investigation; for an issue needing multi-project research, the /investigate
  command orchestrates the investigators first, then calls this skill to compile the result. Do NOT
  use for: scoping a new feature (use /analyze), writing a spec (use spec-write), or generic notes
  (use note-write).
allowed-tools: Read, Grep, Glob, Write, Edit, AskUserQuestion
---
# RCA note

This skill owns the RCA note **format** and the file mechanics: the section template, the
conditional sections, the cross-cutting checks, and the writing principles. It does not own the
investigation process — spawning investigator subagents, tracing code flow, or confirming a root
cause is the `/investigate` command's job. When fired directly (not via `/investigate`), assemble
the caller's findings into the template, ask if intent is unclear, and write the note; do light
grounding reads only, never fan out agents.

An RCA note IS a note: the **frontmatter, the filename, and the location follow the note-write
skill**. This skill owns only the RCA-specific **body** section template on top of those
conventions. Do not redefine frontmatter or filename format here — note-search greps those fields
uniformly across every note, RCA or not, so the shared note-schema contract stays intact.

One job: **create** the RCA note (below).

**Input**: the issue/symptom description plus any findings the caller already has (ticket number,
suspected cause, logs, traced code paths). May name related projects.

## Create

1. **If no clear input, ask**

   Use **AskUserQuestion tool** (open-ended, no preset options):
   > "What issue should the RCA cover? Describe the symptom, the affected environment, and any
   > findings so far."

   Do NOT proceed without understanding the symptom. From the answer, derive the descriptive slug
   (and ticket number if given) for the filename.

2. **Search prior notes**

   Use the **note-search skill** to look for existing notes — a past investigation may already
   cover this area, and related notes belong in the new note's frontmatter `related` field.

3. **Write the RCA note**

   Write the note file per the note-write skill's location and filename conventions
   (`.notes/{YYYY}-{MM}-{DD}[-{TICKET}]-descriptive-slug.md`, lowercase), with the note frontmatter
   (per note-write) followed by the full section skeleton below. Fill every applicable core section;
   omit conditional ones that do not apply. The skeleton itself — every core section header present
   — is how the format is enforced at creation; the investigation then fills each section in place.

4. **Hand off**

   Summarize in chat: note path, and the confirmed root cause if one is already in hand. End with:
   "Review the RCA; the remaining sections fill in as the investigation progresses."

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

## Code flow in the note

The **Evidence** and **Flow diagram** sections render code paths as call-graph trees (per the
**call-graph skill** — indented arrow tree, one node per call hop, each carrying `file:line`). The
`/investigate` command owns WHEN to invoke call-graph during the trace; this skill only requires
that, once traced, a code path lands in those sections as a tree, not prose.

## Guardrails

- Write exactly one note file.
- Do not spawn investigator subagents — that is `/investigate`'s role. If real investigation is
  needed beyond light grounding, defer to `/investigate`.
- Do not present a theory as the root cause until the mechanism is confirmed against the code or the
  data. Say plainly in Open Questions when a cause is still unconfirmed — report faithfully rather
  than papering over.
- Re-read the note from disk before editing it; the investigation may have changed it since you last
  saw it.
