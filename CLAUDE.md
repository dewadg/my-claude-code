# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

A personal collection of Claude Code **configuration artifacts** — subagent definitions, skills, and
an MCP server config. It is not an application: there is nothing to build, lint, or test, and no
package manager. The "code" here is markdown that Claude Code itself consumes.

Artifacts are used by copying them into `~/.claude/agents/` and `~/.claude/skills/` (they are plain
copies, not symlinks, and there is no install script). `~/.claude/` may also hold agents/skills that
are not tracked here — this repo is a curated subset, not a mirror.

## Layout

- `agents/*.md` — one subagent per file. Frontmatter + system prompt.
- `skills/<name>/SKILL.md` — one skill per directory. Supporting scripts live alongside
  (e.g. `skills/telegram-notify/scripts/send_telegram.sh`).
- `.mcp.json` — MCP servers (`context7`, `gitlab`, `goland`, `chrome`). Secrets are
  `${ENV_VAR}` references, never literals.

## Agent frontmatter

```yaml
---
name: golang-eng              # must equal the filename stem
description: "Use this agent when ..."   # third person, trigger-oriented; this is what routes work here
tools: Read, Write, Edit, Bash, Glob, Grep
model: sonnet                 # opus | sonnet | haiku | inherit
effort: medium                # low | medium | high
color: blue
skills:                       # skills the agent may invoke
  - golang-unit-test
  - note-write
  - note-search
mcpServers:                   # names MUST exist in .mcp.json
  - goland
---
```

Body convention across agents: a role paragraph, a "When invoked" list, then phased workflow
sections. Follow the shape of an existing agent when adding a new one.

## Skill frontmatter

```yaml
---
name: code
description: >                # what it does + explicit "use when" examples + explicit "not for" cases
  ...
allowed-tools: Read, Grep, Glob, Bash, Task, ...   # optional; may be scoped, e.g. Bash(curl *) Bash(jq *)
effort: medium                # optional
---
```

Descriptions carry the routing burden — they are how Claude decides to fire the skill. Write them
with concrete trigger phrases and, where a sibling skill overlaps, an explicit exclusion (see
`analyze` vs `investigate` vs `code`, which disambiguate each other in their own descriptions).

## Cross-cutting contracts

These span multiple files; changing one means changing the others.

**Notes.** `note-write` and `note-search` define a shared note format that `analyze`, `investigate`,
and `code` all depend on: files at `.notes/` (or wherever `CLAUDE.md` says), named
`{YYYY}-{MM}-{DD}[-{TICKET}]-descriptive-slug.md`, with frontmatter `tags`, `projects`, `created`,
`updated`, `related`, `description`. `note-search` greps those exact fields — if the frontmatter
schema changes in `note-write`, `note-search`'s grep recipes must change with it.

**Specs.** The `spec` skill writes single-file specs at `specs/{slug}.md` (live) and archives shipped
or abandoned ones to `specs/archives/{YYYY-MM-DD}-{slug}.md`. Each spec carries frontmatter `status`
(`draft` | `in-progress` | `shipped` | `abandoned`), `created`, `updated`, `tags`, `projects`,
`ticket`, `related`, `description` — `tags`/`projects`/`created`/`updated`/`related`/`description`
match the note schema deliberately. The skill's **Search** job greps these exact fields and
awk-extracts the frontmatter block only, so look-back never reads a spec body. If the spec
frontmatter schema changes, those grep recipes and the template's frontmatter block must change
with it.

**Orchestrator skills.** `analyze`, `investigate`, and `code` share one architecture: plan with
TaskCreate/TaskUpdate, delegate every project-level unit of work to a subagent, compile into a note.
They deliberately do **not** hardcode an agent roster — they route by capability ("a service backend
→ the engineering agent for that language") so that adding an agent to `agents/` makes it usable
without editing any skill. Preserve that indirection; do not introduce agent-name lists into skills.

**Skill/agent references.** An agent's `skills:` entries and a skill's name must match real
directories under `skills/`; `mcpServers:` entries must match keys in `.mcp.json`. Renaming a skill
or server means grepping the whole repo for the old name.

## Commits

Conventional Commits, scoped by area: `feat(agents):`, `refactor(agents):`, `feat(skills):`, `docs:`.
