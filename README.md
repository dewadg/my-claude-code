# my-claude-code

My collection of Claude Code subagents, skills, and MCP server config.

Nothing here is an application — it's markdown that Claude Code itself consumes. Copy what you want
into `~/.claude/`.

## Install

```bash
cp agents/*.md ~/.claude/agents/
cp -R skills/* ~/.claude/skills/
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
| `compliance-auditor` | GDPR / CCPA obligations, data practices, privacy gaps |
| `db-admin` | Query performance, HA architecture, disaster recovery |
| `docker-eng` | Designing, building, running, optimizing containers |
| `golang-eng` | Go services, concurrency, cloud-native systems |
| `js-eng` | Modern JavaScript — browser, Node.js, Bun, full-stack |
| `qa-eng` | QA strategy, test planning, quality metrics |
| `react-eng` | React 18+ performance, state management, component architecture |
| `security-auditor` | Security audits, vulnerability analysis, compliance gaps |
| `ui-tester` | Exhaustive UI/UX testing from documented user flows |
| `vue-eng` | Vue 3 Composition API, reactivity, Nuxt 3 |

These agents are from [awesome-claude-code-subagents](https://github.com/VoltAgent/awesome-claude-code-subagents) and modified for my personal use.

## Skills

### Orchestrators

These three share one shape: plan the work as tasks, delegate every project-level unit to a
subagent, compile the result into a note. They route to agents by capability rather than by name, so
adding an agent to `agents/` makes it usable without touching a skill.

| Skill | Use for | Not for |
|---|---|---|
| `analyze` | Scoping a feature or change before building it | Diagnosing a bug; writing the code |
| `investigate` | Finding the root cause of something broken | Scoping a new feature; writing the fix |
| `code` | Writing or changing code across one or more repos | Diagnosing an unexplained bug; scoping |

### Notes

| Skill | Use for |
|---|---|
| `note-write` | Persist knowledge as a dated, tagged markdown note |
| `note-search` | Retrieve past notes by project, topic, ticket, or date |

Notes live in `.notes/` (or wherever the project's `CLAUDE.md` says), named
`{YYYY}-{MM}-{DD}[-{TICKET}]-slug.md`, with `tags` / `projects` / `created` / `updated` / `related` /
`description` frontmatter. The orchestrator skills read and write this format, so the two note skills
must stay in sync.

### Testing

| Skill | Use for |
|---|---|
| `golang-unit-test` | Go table-driven tests with before/after hooks |
| `http-api-test` | Drive HTTP endpoints with curl and verify responses |

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
- **chrome-devtools** — browser driving for UI testing

Agents opt into servers via their `mcpServers:` frontmatter, so a server name here must match the key
in `.mcp.json`.

## License

MIT — see [LICENSE](LICENSE).
