---
name: note-search
description: >
  Search `.notes/` (including `archive/`) by tag, topic, ticket, project or date, and read only a
  candidate's frontmatter before opening the body. Use for "what did we decide about…", "find the
  RCA/design doc on…", "did we hit this before?", or when another skill needs prior context. Not
  for writing notes (note-write).
allowed-tools: Read, Grep, Bash
---

Retrieve relevant notes.

## Input
As the user instructs OR can be by filtered projects, topics, ticket numbers.

## Output
List of relevant note files based on the keywords.

## Note Format

Defined by the **note-write skill** — don't restate it. Fields to filter on: `tags` (topics, ticket
numbers), `projects`, `created`, `updated`, `description`; `related` links sibling files.

### Location
The project's `CLAUDE.md`, else `.notes/` in the workspace.

**Always search recursively (`-r`), never a flat `.notes/*.md` glob** — `.notes/archive/` holds
still-valid prior notes and a flat glob skips it.

### Searching for note files with `grep`

**IMPORTANT**: Respect the frontmatter fields.

```
grep -rl "tags: .*keywords" .notes/            # topic or ticket number
grep -rl "projects: .*user-management-svc" .notes/
grep -rl "created: 2025-04-30" .notes/         # also: updated:
```

### Reporting archived notes

Mark any hit under `archive/` as archived when listing results, and rank current notes above
archived ones. If a current and an archived note cover the same topic, prefer the current one and
mention the archived note as prior context.
**IMPORTANT**: Refrain from reading the whole file without checking the frontmatter:
```
awk '/^---$/{c++; if(c==2){exit;} next} c==1' .notes/not-sure-the-correct-note.md
```
