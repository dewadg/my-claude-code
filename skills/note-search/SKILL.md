---
name: note-search
description: Retrieve relevant notes. ALWAYS use when the user needs to retrieve persisted knowledge from notes.
allowed-tools: Read, Grep, Bash
---

Retrieve relevant notes.

## Input
As the user instructs OR can be by filtered projects, topics, ticket numbers.

## Output
List of relevant note files based on the keywords.

## Note Format

Each notes have following frontmatter that can be used to filter:
```yaml
---
tags: [go, http] # topics, ticket numbers
projects: [user-management-svc, sso-frontend] # related projects
created: 2026-04-30 # when written
updated: 2026-05-02 # last edited
related: # linked notes/files
  - go-context.md
  - blabla.json
description: What this is about.
---
```

### Location
As defined on CLAUDE.md OR to `.notes/` relative to the workspace by default.

The note dir may contain an `archive/` subdirectory holding previous/outdated notes.
**Always search it too** — archived notes are still valid results.

### Searching for note files with `grep`

**IMPORTANT**: Respect the frontmatter fields.

**IMPORTANT**: Search recursively (`-r`), never with a flat `.notes/*.md` glob — a flat glob
skips `.notes/archive/`.

Search by topics, ticket number with `tags`:
```
grep -rl "tags: .*keywords" .notes/
```

Search with `projects` field:
```
grep -rl "tags: .*user-management-svc" .notes/
```

Search by date:
```
grep -rl "created: 2025-04-30" .notes/
grep -rl "updated: 2025-04-30" .notes/
```

### Reporting archived notes

Mark any hit under `archive/` as archived when listing results, and rank current notes above
archived ones. If a current and an archived note cover the same topic, prefer the current one and
mention the archived note as prior context.
**IMPORTANT**: Refrain from reading the whole file without checking the frontmatter:
```
awk '/^---$/{c++; if(c==2){exit;} next} c==1' .notes/not-sure-the-correct-note.md
```
