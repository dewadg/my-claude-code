---
name: design-write
description: >
  Compile analysis findings into a design-doc note: TLDR, scope, current state, impact map,
  proposed change, alternatives, risks, rollout, test strategy, delivery plan. Use for "write the
  design doc", "format this as an impact assessment", or when /analyze hands off. Owns the
  design-doc body template only. Not for RCA (rca-write), a spec (spec-write), or generic notes
  (note-write).
allowed-tools: Read, Grep, Glob, Write, Edit, AskUserQuestion
---
# Design-doc note

Owns the design-doc note **body format**: section template, conditional sections, cross-cutting
checks, writing principles. Frontmatter, filename and location are **note-write's** — don't
redefine them here, note-search greps those fields uniformly across every note. The analysis
(tracing code, committing to a recommendation) belongs to `/analyze`; fired directly, assemble the
caller's findings, ask if intent is unclear, and write the note — light grounding reads only, never
fan out agents.

**Input**: the change plus any findings in hand (ticket, current-state traces, impact map,
proposed approach).

## Create

1. No clear input → AskUserQuestion, open-ended: "What change do you want analyzed? The feature or
   change, the projects it touches, findings so far." Derive the slug (and ticket) from the answer.
2. Prior notes → **note-search**; related ones go in `related`.
3. Write per note-write: frontmatter, then every core section below. Conditional sections only
   when they apply. The full skeleton at creation is what enforces the format; the analysis fills
   it in place.
4. Hand off in chat: path, the recommended approach, open questions, and "Review the design doc;
   the remaining sections fill in as the analysis progresses."

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

## Code flow

**Current state** and **Flow diagram** render each traced path as a call-graph tree (per the
**call-graph skill**: indented arrows, one hop per line, `file:line` per node), never prose.
`/analyze` decides when to trace; this skill fixes only how the trace lands in the note.

## Guardrails

- One note file. No subagents — real analysis defers to `/analyze`.
- No recommendation presented as decided until the alternatives and their tradeoffs are on the
  page; an open choice goes in Open Questions, stated plainly.
- Re-read the note from disk before editing; the analysis may have changed it.
