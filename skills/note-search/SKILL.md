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

### Searching for note files with `grep`

**IMPORTANT**: Respect the frontmatter fields.

Search by topics, ticket number with `tags`:
```
grep -l "tags: .*keywords" .note/*.md
```

Search with `projects` field:
```
grep -l "tags: .*user-management-svc" .note/*.md
```

Search by date:
```
grep -l "created: 2025-04-30" .note/*.md
grep -l "updated: 2025-04-30" .note/*.md
```
**IMPORTANT**: Refrain from reading the whole file without checking the frontmatter:
```
awk '/^---$/{c++; if(c==2){exit;} next} c==1' .memory/not-sure-the-correct-note.md
```
