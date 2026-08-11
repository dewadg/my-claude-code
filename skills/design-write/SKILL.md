---
name: design-write
description: >
  Compile feature-analysis findings into a design-doc note — owns the design-doc note FORMAT: the
  section template, conditional sections, cross-cutting checks, and writing principles. Built on the
  note-write skill's frontmatter + filename + location conventions — a design-doc note IS a note.
  Use whenever the user wants a design doc, analysis note, or impact assessment written, formatted,
  or compiled — e.g. "write the design doc", "compile the analysis into a note", "format this as an
  impact assessment", or when the /analyze command hands off findings. This skill formats and writes
  the note only — it does NOT fan out investigator subagents or drive the analysis; for a change
  needing multi-project research, the /analyze command orchestrates the investigators first, then
  calls this skill to compile the result. Do NOT use for: a root-cause analysis (use rca-write),
  writing a spec (use spec-write), or generic notes (use note-write).
allowed-tools: Read, Grep, Glob, Write, Edit, AskUserQuestion
---
# Design-doc note

This skill owns the design-doc note **format** and the file mechanics: the section template, the
conditional sections, the cross-cutting checks, and the writing principles. It does not own the
analysis process — spawning investigator subagents, tracing code flow, or committing to a
recommendation is the `/analyze` command's job. When fired directly (not via `/analyze`), assemble
the caller's findings into the template, ask if intent is unclear, and write the note; do light
grounding reads only, never fan out agents.

A design-doc note IS a note: the **frontmatter, the filename, and the location follow the note-write
skill**. This skill owns only the design-doc-specific **body** section template on top of those
conventions. Do not redefine frontmatter or filename format here — note-search greps those fields
uniformly across every note, design doc or not, so the shared note-schema contract stays intact.

One job: **create** the design-doc note (below).

**Input**: the feature/change description plus any findings the caller already has (ticket number,
current-state traces, impact map, proposed approach). May name related projects.

## Create

1. **If no clear input, ask**

   Use **AskUserQuestion tool** (open-ended, no preset options):
   > "What change do you want analyzed? Describe the feature or change, the projects it touches, and
   > any findings so far."

   Do NOT proceed without understanding intent. From the answer, derive the descriptive slug (and
   ticket number if given) for the filename.

2. **Search prior notes**

   Use the **note-search skill** to look for existing notes — a past analysis may already cover this
   area, and related notes belong in the new note's frontmatter `related` field.

3. **Write the design-doc note**

   Write the note file per the note-write skill's location and filename conventions
   (`.notes/{YYYY}-{MM}-{DD}[-{TICKET}]-descriptive-slug.md`, lowercase), with the note frontmatter
   (per note-write) followed by the full section skeleton below. Fill every applicable core section;
   omit conditional ones that do not apply. The skeleton itself — every core section header present
   — is how the format is enforced at creation; the analysis then fills each section in place.

4. **Hand off**

   Summarize in chat: note path, recommended approach, and any open questions needing the user. End
   with: "Review the design doc; the remaining sections fill in as the analysis progresses."

## The design-doc note template

The note is a design doc a reviewer should be able to approve or reject without reading the code.
Scale the depth to the change, but never drop the core sections.

### Core sections (always)

- **Frontmatter** — per the note-write skill (tags incl. ticket number, projects, related)
- **TLDR** — what, why, and the verdict (size + risk) in five lines or fewer
- **Context** — ticket acceptance criteria, current behaviour, why now
- **Scope** — in scope / out of scope / assumptions
- **Current state** — per project, how it works today, anchored with `file:line` references and
  the call-graph tree where the flow matters
- **Impact map** — table: `project | layer | files | new vs. modify | effort`
- **Proposed change** — per project: API contract, data model and migrations, events, config and
  env vars
- **Flow diagram** — a call-graph tree for code call paths; a mermaid diagram (sequence,
  flowchart, or C4) when the change has cross-service moving parts worth visualizing: request/data
  flow across services, component interaction, deploy order, or state transitions. Skip for trivial
  single-file changes — the prose is enough.
- **Alternatives considered** — the options, their tradeoffs, and why the recommended one wins
- **Risks and drawbacks** — table: `risk | impact | mitigation`
- **Feature flag and rollout** — is a flag needed (name, default, kill switch); deploy order across
  repos and services
- **Test strategy** — unit / integration / E2E coverage plus manual QA scenarios
- **Delivery plan** — work breakdown in dependency order, PR split per repo, sizing
- **Open questions** — each with an owner
- **References** — ticket, designs, related notes

### Conditional sections (only when the change triggers them)

Data backfill and migration · backward compatibility and versioning · multi-tenancy · auth and
permission model · i18n · performance and load · security and privacy.

### Cross-cutting checks

Call these out explicitly whenever the change touches them:

- **Generated artifacts**: does an API spec, client, schema type, or mock need regenerating?
- **Shared contracts**: does a shared package or event schema need a version bump, and which
  consumers must follow?
- **Deploy order**: which service must ship before which client, and what deployment config follows

## Code flow in the note

The **Current state** and **Flow diagram** sections render code paths as call-graph trees (per the
**call-graph skill** — indented arrow tree, one node per call hop, each carrying `file:line`). The
`/analyze` command owns WHEN to invoke call-graph during the trace; this skill only requires that,
once traced, a code path lands in those sections as a tree, not prose.

## Guardrails

- Write exactly one note file.
- Do not spawn investigator subagents — that is `/analyze`'s role. If real analysis is needed beyond
  light grounding, defer to `/analyze`.
- Do not present a recommendation as decided until the alternatives and their tradeoffs are on the
  page. Say plainly in Open Questions when a choice is still open — report faithfully rather than
  papering over.
- Re-read the note from disk before editing it; the analysis may have changed it since you last saw
  it.
