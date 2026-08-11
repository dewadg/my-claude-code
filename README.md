# my-claude-code

My collection of Claude Code subagents, skills, commands, and MCP server config.

Nothing here is an application — it's markdown that Claude Code itself consumes. Copy what you want
into `~/.claude/`.

## Install

```bash
cp agents/*.md ~/.claude/agents/
cp -R skills/* ~/.claude/skills/
cp commands/*.md ~/.claude/commands/
```

`.mcp.json` is a project-scoped MCP config; copy it into a project (or merge into `~/.claude.json`)
and export the env vars it references: `CONTEXT7_API_KEY`, `GITLAB_PAT`.

## Agents

Subagents Claude delegates to. Each file is frontmatter (name, description, tools, model, allowed
skills, MCP servers) plus a system prompt.

| Agent | Use for |
|---|---|
| `api-documenter` | API docs, OpenAPI specs, doc portals, generated code examples |
| `architect-reviewer` | System design decisions, architectural patterns, technology choices |
| `code-reviewer` | Code quality, security vulnerabilities, best practices |
| `code-diver` | Read-only codebase exploration — trace call flow (call-graph tree), locate symbols (file:line table); spawned by `analyze`/`investigate`/`spec` for code reading, escalates to a specialist on a stack-semantic gap |
| `compliance-auditor` | GDPR / CCPA obligations, data practices, privacy gaps |
| `db-admin` | Query performance, HA architecture, disaster recovery |
| `docker-eng` | Designing, building, running, optimizing containers |
| `golang-eng` | Go services, concurrency, cloud-native systems |
| `js-eng` | Modern JavaScript — browser, Node.js, Bun, full-stack |
| `qa-eng` | Versatile QA — BE endpoints via http-request, FE/UI via chrome MCP, plus test strategy, automation, quality metrics |
| `react-eng` | React 18+ performance, state management, component architecture |
| `security-auditor` | Security audits, vulnerability analysis, compliance gaps |
| `vue-eng` | Vue 3 Composition API, reactivity, Nuxt 3 |

These agents are from [awesome-claude-code-subagents](https://github.com/VoltAgent/awesome-claude-code-subagents) and modified for my personal use.

## Commands

User-invoked slash commands. Run as `/analyze`, `/investigate`, `/spec`.

| Command | Use for | Not for |
|---|---|---|
| `analyze` | Scoping a feature or change across projects before building it | Diagnosing a bug (`/investigate`); writing the code (`/code`) |
| `investigate` | Finding the root cause of something broken | Scoping a new feature (`/analyze`); writing the fix (`/code`) |
| `spec` | Spec out a change end-to-end into a build-ready `specs/{slug}.md` — orchestrates research + decisions, delegates the file format to the `spec-write` skill | Implementing decided work (`/code`); open-ended impact analysis (`/analyze`) |

All three delegate code reading to `code-diver` and, when explaining code flow, render it with the
`call-graph` skill (indented arrow tree, one node per call hop) rather than prose.

## Skills

### Planning

| Skill | Use for | Not for |
|---|---|---|
| `spec-write` | The spec file FORMAT — frontmatter schema, section template, lean principles, search recipes; auto-fires on spec-writing intent, also driven by the `/spec` command | Implementing decided work (`/code`); diagnosing a bug (`/investigate`); archiving a done spec (`spec-archive`) |
| `spec-archive` | Move a shipped or abandoned spec to `specs/archives/` after finalizing its status — triggered by "archive the spec", "/spec archive", or any shelving intent | Writing/formatting a spec (`spec-write`); implementing the todo (`/code`) |

### Orchestrators

`analyze` and `investigate` (commands, above) and `code` (skill, below) share one shape: plan the
work as tasks, delegate every project-level unit to a subagent, compile the result into a note. They
route to agents by capability rather than by name, so adding an agent to `agents/` makes it usable
without touching the command or skill. `spec` (command, above) shares the same delegation shape but
compiles into a spec file via the `spec-write` skill rather than a note. For code reading and tracing
they spawn `code-diver` by default, escalating to a language specialist only when `code-diver` flags a
stack-semantic gap (`→ needs <specialist>`) — `smart-skip` sends semantic-flavored requests straight
to the specialist, `pass-tree` hands `code-diver`'s output to it so the specialist does only the
semantic step.

### Code reading

| Skill | Use for |
|---|---|
| `call-graph` | Render how code flows as an indented call-graph tree — one node per call hop with `file:line`. Fires on explicit invoke or "explain how X works" |

| Skill | Use for | Not for |
|---|---|---|
| `code` | Writing or changing code across one or more repos | Diagnosing an unexplained bug (`/investigate`); scoping (`/analyze`) |

### Notes

| Skill | Use for |
|---|---|
| `note-write` | Persist knowledge as a dated, tagged markdown note |
| `note-search` | Retrieve past notes by project, topic, ticket, or date |

Notes live in `.notes/` (or wherever the project's `CLAUDE.md` says), named
`{YYYY}-{MM}-{DD}[-{TICKET}]-slug.md`, with `tags` / `projects` / `created` / `updated` / `related` /
`description` frontmatter. The orchestrator skills read and write this format, so the two note skills
must stay in sync.

### API & testing

| Skill | Use for |
|---|---|
| `golang-unit-test` | Go table-driven tests with before/after hooks |
| `http-request` | Drive HTTP endpoints with curl and verify responses |
| `test-scenario` | Given/When/Then test scenarios as a markdown table — happy, alternate, edge, error paths |
| `hoppscotch-import` | Generate a Hoppscotch collection JSON from an OpenAPI spec, route code, Postman, HAR, or plain list |

### Writing

| Skill | Use for |
|---|---|
| `ticket-description` | Ticket/issue description bodies in a fixed section structure, any tracker |
| `pr-description` | PR/MR description bodies from the branch diff, emoji-sectioned template |

Both produce description text only — they never create the ticket or open the PR.

### Release & ops

| Skill | Use for |
|---|---|
| `glab-release` | Cut a GitLab release: pick prev tag, build a conventional-commit changelog, tag |
| `telegram-notify` | Ping me on Telegram when a long-running job or watch condition resolves |

## MCP servers

`.mcp.json` wires up:

- **context7** — up-to-date library/framework docs
- **gitlab** — issues, merge requests, pipelines
- **goland** — IDE integration (local, `127.0.0.1:64342`)
- **chrome** — browser driving for UI testing

Agents opt into servers via their `mcpServers:` frontmatter, so a server name here must match the key
in `.mcp.json`.

## License

MIT — see [LICENSE](LICENSE).
