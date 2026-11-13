# Seams

A seam is a boundary where work splits cleanly between engineers: different codebase, different
skill, or a dependency one side can mock. Tickets never cross a seam.

## Primary seams and granularity

### Backend — one ticket per endpoint

- Unit: one endpoint (`METHOD /path`), or one async consumer/job/cron.
- Title names it: `[BE] POST /orders/{id}/refund — issue a partial refund`.
- **Contract in the ticket**: request/response shape, status + error codes, auth/role. FE/mobile
  tickets reference it and build against a mock.
- **Groundwork gets its own ticket** that endpoint tickets are `blocked by`: schema/migration, a new
  domain model/aggregate, a third-party client, a shared permission/role change.
- Merge trivial siblings when each alone is XS (e.g. plain list + get on one resource) — say so.
- Changing an existing endpoint: one ticket per endpoint changed, note backward compatibility
  (mobile clients in the wild can't upgrade instantly).

### Frontend (web) — one ticket per screen

- Unit: one screen/route/page, wiring data + all its states (loading, empty, error, permission).
- Title: `[FE] Order detail — refund entry point`.
- **Component gets its own ticket only when** it's shared by ≥ 2 screens in this breakdown, or is
  ≥ M on its own (complex form, data grid, editor). Otherwise it folds into its screen ticket —
  a lone component has no QA-able outcome.
- Screen ticket is `blocked by` its component tickets; references the BE contract.
- Modal/drawer with its own flow (multi-step, own API call) counts as a screen.

### Mobile — one ticket per screen

- Same rules as FE.
- **Codebase split**: cross-platform (React Native/Flutter) → one ticket per screen. Separate
  native iOS + Android codebases → one ticket per screen **per platform** (`[iOS]`, `[Android]`).
- Mobile-only work worth separate tickets when present: push notification wiring, deep links,
  offline/sync, permissions prompts, app release + store submission (with review lead time).
- Backward compatibility: old app versions keep calling old endpoints — flag in the BE ticket.

## Cross-cutting seams — cut only when real

| Seam | Ticket when | Title prefix |
|---|---|---|
| Data | migration, backfill, data fix on existing rows | `[BE]` or `[Data]` |
| Async | new queue consumer, scheduled job, webhook receiver | `[BE]` |
| Integration | new third-party API/SDK, credentials, sandbox setup | `[BE]` |
| Infra | new env vars/secrets, queue/bucket/topic, service, CI change | `[Infra]` |
| Flag | feature-flagged rollout (creation + cleanup ticket later) | seam of the code |
| Analytics | tracking events for the metric in business framing | seam that emits them |
| Admin | back-office toggle/config/view for ops | `[FE]`/`[BE]` |
| Spike | genuine unknown blocks sizing (feasibility, third-party, perf) | `[Spike]` |

Not engineering tickets — list as **pre-conditions** instead: design not ready, copy not final,
legal sign-off, third-party contract. They block tickets; they aren't tickets.

## Slicing rules

- One ticket = one seam, one engineer, one mergeable deliverable, size ≤ L. XL → split.
- **Contract-first** — BE contract defined in its ticket lets FE/mobile run in parallel.
- **Thin slice early** — one happy-path end-to-end slice (BE + one client) first, then breadth.
- **Spike first** for unknowns — time-boxed (≤ 2 days), output a decision, not production code.
- Every ticket traces to `R#`s; every `R#` lands in ≥ 1 ticket.
- Mark `blocked by` explicitly; anything not blocked runs in parallel.

## Code-diver briefs per seam

Spawn one per project; ask for `file:line` tables, inventory only.

- **Backend**: "List existing routes/endpoints under <domain> (`METHOD path → handler file:line`),
  the models/tables they touch, any third-party clients involved, existing async consumers/jobs in
  this area."
- **Frontend / mobile**: "List existing screens/routes in <area> (`route → file:line`), reusable
  components already doing part of <capability>, and the API calls those screens make."
- **Infra** (if relevant): "Where are env vars, queues, buckets declared for <service>?"

Reuse beats new: an existing endpoint needing one field, or a component needing one prop, is an
XS/S change on the existing ticket, not a new one.
