# Domain package layout

Reference implementation: `assets/scaffold/internal/health`.

## Files

| File | Holds |
| --- | --- |
| `module.go` | `var Module = fx.Module("<domain>", ...)`; fx.In param structs; `RegisterRoutes` adapters. Composition root. |
| `routes.go` | `Register(g, Deps)` only — the route table, no fx types |
| `deps.go` | `Deps` structs the route table needs (consumer interfaces); fx.In structs for cross-domain deps |
| `handler.go` | transport: bind → call service → map error → respond. No persistence. |
| `service.go` / `service_<flow>.go` | domain logic. No router, no persistence imports. |
| `repository.go` / `repository_<flow>.go` | the only files that import the DB driver/ORM; map not-found to a domain sentinel |
| `schemas.go` | request/response DTOs, status constants |
| `errors.go` | `Err*` sentinels and types implementing `error` (not `mapError` — that stays with its handler) |
| `doc.go` | package comment |
| `service_interfaces.go` | interfaces the handler consumes + `var _ iface = (*Service)(nil)` |
| `store_interfaces.go` | interfaces the service consumes + `var _ iface = (*Repository)(nil)` |

Omit files with nothing in them. Repository is exposed to fx under its interface:
`fx.Provide(fx.Annotate(NewRepository, fx.As(new(orderStore))))`.

## File-name grammar

`{actor_}{layer}{_flow}.go`

- `actor` (optional, multi-actor domains): `admin`, `visitor`, `shared`
- `layer`: `handler` (never `handlers`), `service`, `repository`, `routes`, `module`, `schemas`,
  `errors`, `deps`, `doc`, `service_interfaces`, `store_interfaces`
- `flow`: anything after the layer word — `service_export.go`, `admin_handler_session.go`
- The package name as a prefix is dropped: `order_repository.go` → `repository.go`;
  any other leading token moves after the layer: `session_handler.go` → `handler_session.go`
- A file with no layer word is a **concern file** (noun: `cursor.go`, `pricing.go`). It must be
  consumed by ≥2 layers/families; if only one consumes it, rename it into that layer or fold it in.
  A concern file importing the DB driver belongs in the repository layer; importing the router,
  in the handler layer.

## Layer direction

`repository` < `service` < `handler` < (`routes`, `module`). A symbol may be consumed only by a
strictly higher layer (or by `errors.go`/`schemas.go`/`doc.go`/`*_interfaces.go`/`deps.go`, which
anyone may use). Shared same-layer helpers across actors go in `shared_<layer>.go`. A helper used
by one flow file lives in that flow file; used by several flows of one family, in the base file.

## Exemptions

`infra`, `seeders`, `httpserver`, `pkg/*` are not domains: no layering rules. A grown sub-flow with
its own handler→service→repository chain becomes a subpackage (`order/refund/`), not a deeper flat
layout.

## Testing

Service tests mock the store interfaces; handler tests mock the service interface; no test touches
a real database. Use the `golang-unit-test` skill for the test shape.

## Audit coverage

`scripts/audit.sh` FAILs on: missing `module.go`/`fx.Module("<domain>")`, `handlers` file names,
layer word not leading, persistence imports outside repository/module/deps, router imports in
service/repository/concern files, `Err*` sentinels outside `errors.go`. WARNs on: module not
registered in `cmd/`, cross-domain imports, context-taking methods with no span.
