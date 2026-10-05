---
name: golang-project-structure
description: >
  Scaffold, extend, audit, or migrate a Go service onto one opinionated structure: cobra commands
  in cmd/{bin} (serve, migrate, seed), Uber fx for DI and lifecycle, caarlos0/env + godotenv parsed
  into internal/pkg/config before the graph, goose SQL migrations embedded from build/migrations,
  OpenTelemetry logs/metrics/traces in internal/infra + internal/pkg/observability, log/slog with
  canonical source/error keys, internal/{domain} packages with a fixed file-name grammar, a
  multi-stage distroless Dockerfile and a standard Makefile (dev, build, migrate, seed, test).
  ALWAYS use when the user asks to start, scaffold, or bootstrap a Go service/API/worker; add a
  domain, module, endpoint, migration, seeder, or subcommand to a Go service; check, audit, lint,
  or enforce Go project structure or logging conventions; or restructure/migrate an existing Go
  project onto this layout — even if they only say "spin up a Go API", "add an orders module",
  "where should this file go", or "is this repo following our structure". Not for writing unit
  tests (use golang-unit-test) or for non-Go services.
allowed-tools: Read, Write, Edit, Bash, Glob, Grep, AskUserQuestion
effort: medium
---

# Go project structure

One structure across every Go service, so any engineer (or agent) can open any repo and know where
things are. The skill ships a **runnable scaffold** (`assets/scaffold/`) and an **audit script**
(`scripts/audit.sh`) — the scaffold is the canonical form of every file, the audit is how
conformance is checked. When this document and the scaffold disagree, the scaffold wins.

Paths below are relative to this skill directory unless they start with a project path.

## What is enforced vs. yours to choose

| Enforced (audit FAILs) | Yours to choose |
| --- | --- |
| `go.uber.org/fx` for DI + lifecycle | HTTP router (scaffold uses echo; chi/gin/net/http fine) |
| `github.com/spf13/cobra` root in `cmd/{bin}/main.go` | DB driver / ORM (scaffold: `database/sql` + pgx; gorm/sqlc fine) |
| `github.com/caarlos0/env/v11` + `github.com/joho/godotenv` in `internal/pkg/config` | Whether a service has HTTP at all (worker: drop `httpserver`) |
| `github.com/pressly/goose/v3`, SQL files embedded from `build/migrations` | Seeder contents, domain names |
| OTel via `internal/pkg/observability` (`Setup*`, `Start`/`End`) wired in `internal/infra` | Extra `internal/pkg/*` helpers |
| `log/slog` only, `logging.KeySource` / `logging.KeyError` | |
| Domain file-name grammar + layer import rules | |
| Multi-stage Dockerfile, version via ldflags, distroless nonroot | |
| Makefile targets `dev serve build migrate test vet` (+ `seed migrate-new migrate-down migrate-status lint docker-build`) | |

## The layout

```
cmd/{bin}/
  main.go        cobra root; bootstrap slog; config.LoadDotenv(); bare binary = serve
  serve.go       config.Parse → AssertRequiredEnv → fx.New(fx.Supply(cfg), modules...).Run()
  migrate.go     goose provider over migrations.FS: migrate | migrate down | status | create <name>
  seed.go        idempotent seeders, outside fx
  db.go          withDB: short-lived connection for one-off commands
build/migrations/
  embed.go       //go:embed *.sql  → migrations.FS
  00001_init.sql goose sequential: NNNNN_snake_case.sql, -- +goose Up / Down
internal/
  pkg/config/        raw string struct (env tags) → typed Config; Parse, AssertRequiredEnv, LoadDotenv
  pkg/logging/       KeySource, KeyError, ... ; request-id context helpers
  pkg/observability/ SetupLogs, SetupMetrics, SetupTracing; Start/End/Detach span helpers; slog handlers
  pkg/buildutil/     Name, Version (ldflags)
  pkg/httpapi/       transport helpers shared by handlers (Data, FailWith)
  infra/             fx.Module("infra"): base logger, logger, meter/tracer providers, DB
  httpserver/        fx.Module("httpserver"): router, name-tagged groups, middleware, lifecycle
  seeders/           seeders.Run(ctx, db) — exempt from layering
  {domain}/          fx.Module("{domain}") — see references/domain-layout.md
Dockerfile  .dockerignore  Makefile  .env.example  .gitignore
```

`internal/pkg/*` never imports a domain. A domain never imports another domain for construction:
it declares the interface it consumes (`deps.go` / `*_interfaces.go`) and fx wires the concrete type.

## Rules (each exists because of a failure it prevents)

1. **Config is parsed in the subcommand, before fx exists**, then `fx.Supply(cfg)`. A bad env fails
   with one plain line, not a stack trace from inside a constructor. `raw` holds every var as a
   string; `Parse` converts. Safely-defaultable values **warn and fall back**; values that can't be
   defaulted **error** in `Parse` (malformed) or `AssertRequiredEnv` (missing). Only `config` reads
   the environment. Every var `raw` reads is listed in `.env.example`.
2. **`serve.go`'s `fx.New` list is the only module registration point.** Order is for reading; fx
   resolves by type. Real ordering is expressed as a dependency, never list position.
3. **Migrations are a separate one-off command, never run at boot.** `migrate` (then `seed`) gates
   deploy; `serve` assumes the schema is current. Same image runs all three subcommands.
4. **Domains never receive the router.** They take a route group by name tag
   (`` Group *echo.Group `name:"public"` ``) so they can't bypass group middleware.
   `RegisterRoutes(fx.In params)` is a thin adapter over plain `Register(group, Deps)`.
5. **Constructors are pure; lifecycle is a separate `fx.Invoke`.** Bind the socket synchronously in
   `OnStart` (port in use fails startup), serve on a goroutine.
6. **Whoever opens a resource registers its `OnStop`** right there. Telemetry flush failures are
   logged and return `nil` (a returned error turns SIGTERM into a non-zero exit); DB close returns
   its error.
7. **fx is lazy.** A side effect nothing consumes (runtime metrics) is an `fx.Invoke`, or it never runs.
8. **Telemetry off ≠ telemetry absent.** Disabled pipelines return no-op providers so call sites
   never branch. `OTEL_ENABLED` is the master switch; per-signal switches are ANDed in config.
9. **Every handler/service/repository method that takes a context opens a span** named
   `<package>.<Struct>.<Method>` with `ctx, span := tracing.Start(ctx, name)` +
   `defer tracing.End(span, &err)` on a named `err` result (import alias `tracing`).
10. **Logging** — see `references/logging.md`. Short form: `l := slog.With(logging.KeySource,
    "<package>.<Struct>.<Method>")` (same string as the span), `*Context` variants, lowercase
    message phrases, `logging.KeyError` never `"err"`, request/trace ids are injected never passed,
    log a failure once at the boundary that handles it.

Deeper detail: `references/domain-layout.md` (file grammar, placement, interfaces),
`references/observability.md` (pipelines, shutdown order, Docker/Makefile notes),
`references/logging.md` (slog conventions).

## Mode A — scaffold a new service

1. **Get the Go module path first; never guess.** Offer candidates from `git remote -v`, an
   existing `go.mod`, and the directory name via AskUserQuestion. Binary name defaults to the last
   path segment.
2. Run the scaffolder (copies `assets/scaffold`, rewrites module/binary/service name, pins the
   local Go version, `go mod tidy`, gofmt check, `go build`, `go vet`, creates `.env`):
   ```bash
   <skill-dir>/scripts/scaffold.sh <dest-dir> <module-path> [binary-name]
   ```
   It refuses a non-empty destination — use Mode B/C there.
3. If the user picked a different router/driver, swap bodies now; file names and boundaries stay.
   No HTTP (worker)? Delete `internal/httpserver`, `internal/pkg/httpapi`, domain `routes.go`, and
   the `RegisterRoutes` invoke.
4. **Prove it runs** — point `DATABASE_URL` in `.env` at a Postgres, then:
   ```bash
   make migrate && make seed && make migrate-status
   make build && ./bin/<bin> serve &   # then:
   curl -s localhost:3000/api/health   # {"status":"ok","version":"dev","checks":{"database":"ok"}}
   kill -INT %1                        # exit 0, "server stopping" logged
   <skill-dir>/scripts/audit.sh <dest-dir>   # 0 FAIL, 0 WARN
   ```
   Report what you observed, not what should happen.

## Mode B — add to an existing project

Read one existing domain first; the repo's own conventions outrank the scaffold where they differ
in body (not in the enforced rules).

- **Domain:** copy `assets/scaffold/internal/health` as the shape — `module.go`, `routes.go`,
  `deps.go`, `handler.go`, `service.go`, `repository.go`, `schemas.go`, `errors.go`, `doc.go`,
  `service_interfaces.go`, `store_interfaces.go`. Omit files with nothing in them. Add
  `{domain}.Module` to `serve.go`'s `fx.New`.
- **Migration:** `make migrate-new name=<snake_case>` → edit `build/migrations/NNNNN_<name>.sql`
  (keep `-- +goose Up` and `-- +goose Down`) → `make migrate`.
- **Seeder:** add a `Seeder{Name, Run}` to `seeders.All()`; it must be idempotent
  (`ON CONFLICT DO NOTHING`, existence checks) → `make seed`.
- **Env var:** field on `raw` (string, `env:` + `envDefault:` tags) → typed field on `Config`
  → conversion in `Parse` (warn+default or error) → `AssertRequiredEnv` if mandatory → line in
  `.env.example`.
- **Subcommand:** `cmd/{bin}/<name>.go` with `newXxxCmd()`, registered in `main.go`. One-off
  commands use `withDB`, never `fx.New`.

Finish with `go build ./... && go vet ./...` and `scripts/audit.sh <project>`.

## Mode C — audit / migrate an existing project

1. Run `scripts/audit.sh <project>` (read-only). `FAIL` = breaks an enforced rule; `WARN` = drift
   worth fixing. Show the user the grouped result before changing anything.
2. Map current files → target locations and show the mapping. A file that doesn't fit usually
   does two jobs (handler that queries the DB) — split it, don't just move it.
3. Bottom up, each slice ending green (`go build ./... && go vet ./...` + existing tests):
   `internal/pkg` (config, logging, observability) → `internal/infra` → transport → `cmd` →
   domains one at a time. Introduce fx last if the project has none. `git mv` so history follows.
4. Behavior stays identical — note improvements for a follow-up instead of smuggling them in.
5. Re-run the audit until 0 FAIL. Large project: convert one domain end-to-end and let the user
   review before the rest.

## Gotchas (hit while building this skill)

- **Check `docker context show` before running containers.** A remote default context silently
  runs your throwaway Postgres on another host (seen: `port is already allocated` on a port free
  locally, then `connection refused` on localhost). Use `DOCKER_CONTEXT=default` (zsh won't
  word-split `DK="docker --context default"`).
- **buildx `docker-container` builders discard the image** without `--load`
  ("No output specified with docker-container driver"); the Makefile's `docker-build` passes it.
- **OTel SDK export errors** go through OTel's global logger and surface as source-less INFO
  lines; `SetupTracing` installs `otel.SetErrorHandler` to route them to slog at WARN.
- **`goose create` needs no DB** — `migrate create` calls `goose.Create(nil, dir, name, "sql")`
  with `goose.SetSequential(true)`; it writes relative to the CWD, so run it from the module root.
- **Inside a container, reach host Postgres** via `host.containers.internal` (podman) /
  `host.docker.internal` (Docker Desktop).
- **`fx` logs every provide/invoke as `[Fx]` lines on stderr.** Expected; filter with
  `grep -v '^\[Fx\]'` when reading output.
