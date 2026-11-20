---
name: note-write
description: >
  Write or update a note in the shared format: frontmatter (`tags`, `projects`, `created`,
  `updated`, `related`, `description`), filename `{YYYY}-{MM}-{DD}[-{TICKET}]-slug.md`, `.notes/`
  location, one directory per run, archived in one move. Use for "write a note", "journal this",
  "record this decision". Not for searching notes (note-search).
allowed-tools: Read, Write, Grep
---

Write efficient notes for future references.

## When invoked

- The user asks to persist something: "write a note about…", "journal this", "save what we
  learned", "record this decision", "note this for next time".
- Another skill hands off a finding to keep — an estimate gap, a resolved question, a decision the
  next session should not re-derive.
- A multi-task or multi-project session needs a durable record: the note becomes the session's
  journal, not a chat transcript.
- The user asks to update an existing note (bump `updated`, amend the body, fix `related:`).

Not for reading notes back (note-search) or for a format another skill already owns — an RCA or a
design doc has its own section template even though it lives in the same note format.

## Input
The context of current conversation OR as the user instructs.

## Output
A markdown file in a designated location.

## Note Format

The note markdown file should have a frontmatter as follow:
```yaml
---
tags: [go, http] # topics, ticket numbers
projects: [user-management-svc, sso-frontend] # related projects
created: 2026-04-30 # when written
updated: 2026-05-02 # last edited
related: # linked notes/files
  - go-context.md
  - blabla.json
description: What this is about in short.
---
```

### Location
As defined in the project's `CLAUDE.md`, or to `.notes/` relative to the workspace by default. A note
may live in a subdirectory of `.notes/` (e.g. a session directory); keep `related:` entries
sibling-relative so moving the directory preserves them.

### Writing
- Write note filename in format `{YYYY}-{MM}-{DD}-descriptive-slug.md` OR if ticket number is supplied `{YYYY}-{MM}-{DD}-{TICKET}-descriptive-slug.md`
- Note filenames should be all lowercase
- Note can have supporting files and should be referred on frontmatter's `related`

### Session directory

A session or run that produces more than one note gets **one directory**, named with the same stem
pattern as a single note:

```
.notes/<session-root>/{YYYY}-{MM}-{DD}[-{TICKET}]-{slug}/
```

Every note and artifact for that run lives inside it and **shares the directory stem**, with a
suffix naming its role — `{stem}-journal.md`, `{stem}-report.md`, `evidence-{project}-{agent}.md` —
so a directory is recognisable from any one file in it. Keep `related:` entries sibling-relative
(`-journal.md`, `evidence-{project}-{agent}.md`), never workspace-relative, so the directory survives
being moved or renamed.

`<session-root>` is `.notes/<area>/` where the owning skill or the project's `CLAUDE.md` says so,
otherwise `.notes/` itself — e.g. `.notes/code/…` for the code skill's session directory.

Archive the whole run in **one** move, not file by file:

```bash
mv .notes/<session-root>/{stem} .notes/<session-root>/archive/
```

The stem moves with it, so sibling-relative `related:` entries stay valid and note-search still
finds it under `archive/`.

### Note body

- A `#` title naming the subject, then short sections — Context, What we found / decided, Evidence
  (paths and `file:line`, not pasted bodies), Follow-ups.
- Write for the next session that has no memory of this one: what was true, what was ruled out, and
  what to do next. No narration of the conversation.
- Point at evidence by path; never paste logs or file bodies.

