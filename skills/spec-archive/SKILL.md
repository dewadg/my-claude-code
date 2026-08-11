---
name: spec-archive
description: >
  Archive a shipped or abandoned spec — move it from the live specs/ set into
  specs/archives/{YYYY-MM-DD}-{slug}.md after finalizing its frontmatter status. Owns the whole
  archive process: identify the target, verify the todo list reflects reality, date-stamp + set status
  (shipped | abandoned), move (not copy), confirm. Use whenever the user wants to shelve a done spec —
  e.g. "archive the spec", "archive <slug>", "/spec archive", "this spec shipped, shelve it", or any
  shelving intent. Do NOT use for: writing or formatting a spec (use spec-write), searching prior specs
  (use spec-write's Search job), or implementing the spec's todo (use /code).
allowed-tools: Read, Grep, Glob, Bash, Edit, AskUserQuestion
---
# Spec archive

Move a shipped or abandoned spec out of the live `specs/` set into `specs/archives/`. Archiving is
finalization, not deletion — the spec stays readable (and searchable by spec-write's Search job, which
reads `specs/archives/` too), it just leaves the active set.

**Input**: the slug to archive, or a request naming the spec. If no slug is given, locate it.

## Steps

1. **Identify the target.**

   Slug named in the request; if none named, `ls specs/*.md` and ask which (AskUserQuestion). If
   `specs/{slug}.md` doesn't exist, say so and stop — don't guess, don't search archives for it
   (already-archived specs are not re-archived).

2. **Check the todo list.**

   Re-read the spec from disk and scan its Todo section. The archive must reflect reality, not a false
   "done".

   - Every box `[x]` → shipped clean. Proceed.
   - Unchecked `[ ]` items remain → surface them and ask (AskUserQuestion): shipped-but-incomplete, or
     abandoned?
     - **Shipped-incomplete**: record which tasks were dropped under a new `## Incomplete` heading
       before moving.
     - **Abandoned**: record that fact under the `## Incomplete` heading.

   Never archive with `[ ]` items unaccounted for.

3. **Date-stamp + finalize status.**

   Set frontmatter `status` to `shipped` (clean or shipped-incomplete) or `abandoned`, and set
   `updated` to today's `YYYY-MM-DD`. Edit the spec on disk before the move so the archive reflects
   reality.

4. **Move it** (not copy):

   ```bash
   mkdir -p specs/archives
   mv specs/{slug}.md specs/archives/{YYYY-MM-DD}-{slug}.md
   ```

5. **Confirm** the destination path to the user.

## Guardrails

- Move exactly one live spec per archive — `specs/{slug}.md` → `specs/archives/{YYYY-MM-DD}-{slug}.md`.
  Use `mv`, never `cp`; archiving removes the spec from `specs/`.
- If `specs/archives/{YYYY-MM-DD}-{slug}.md` already exists, ask overwrite vs skip before replacing.
- Re-read the spec from disk before editing it; the user may have changed it since you last saw it.
- Do not create branches, tags, commits, or tracker issues unless the user asks.
- Only `status` and `updated` change in frontmatter — leave the rest of the schema (created, tags,
  projects, ticket, related, description) untouched so spec-write's Search still routes to it.
