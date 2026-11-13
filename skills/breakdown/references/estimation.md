# Estimation — t-shirt, human time

Sizes are **effort for a human engineer** — mid-level, familiar with the codebase — including tests,
code review round-trips, and QA fixes. **Not** AI-assisted time. **Not** calendar time.

## Scale

| Size | Effort (days) | Sum as | Typical shape |
|---|---|---|---|
| XS | ≤ ½ | 0.5 | copy/config change, add a field end-to-end on one seam |
| S | ~1 | 1 | one simple endpoint or one screen state, well-trodden pattern |
| M | 2–3 | 2.5 | endpoint with business logic + validation + tests; screen with list/form and all states |
| L | ~5 | 5 | new third-party integration, migration with backfill, complex multi-step flow |
| XL | > 5 | — | **not a ticket — split it** |

## Typical sizes per seam (starting point, adjust with drivers)

| Work | Usual |
|---|---|
| BE read endpoint on existing model | S |
| BE write endpoint with validation + business rules | M |
| BE endpoint calling a new third-party | L (or spike + M) |
| BE migration, no backfill / with backfill on large table | S / L |
| BE async consumer or cron job | M |
| FE/mobile simple screen (static, one call) | S |
| FE/mobile screen with list/form + loading/empty/error states | M |
| FE/mobile multi-step flow or complex form | L |
| Shared component, non-trivial (grid, editor, upload) | M |
| Mobile push/deep link wiring (first time in app) | M |
| Analytics events for one flow | XS–S |

## Size drivers — name one per ticket

Each ticket states its driver in one phrase, tied to code evidence when there is some
(`reuses OrderTable — web/src/components/OrderTable.tsx:12`):

- **New vs extend** — new pattern in this codebase costs more than copying an existing one
- **Unknown third-party** — docs quality, sandbox availability, auth model
- **Data** — migration on a large/hot table, backfill, data correctness risk
- **States** — number of UI states and permissions to handle
- **Cross-team** — needs another team's review, API, or deploy
- **Design readiness** — no final design → add risk, don't add size
- **Sensitive** — payments, auth, PII → extra review and testing

## Uncertainty

- Size swings by two steps depending on an unknown → don't pick the middle. Give the likely size,
  mark `risk: high`, and add a `[Spike]` ticket that's blocked-by-nothing and goes first.
- Size hinges on code nobody inspected → trace it (code-diver) before estimating, or mark
  `risk: high — needs trace: <what>`.
- Greenfield project → add one step to anything that establishes a pattern (first endpoint, first
  screen) — later siblings reuse it.

## Totals

- **Effort per lane** = sum of `Sum as` values per seam (BE, FE, mobile, …). Label it **effort**.
- **Rough calendar** (only if the user gives team size per lane):
  `lane days ÷ (engineers on lane × 0.7 focus factor)`, and the overall duration is the **critical
  path** through `blocked by`, not the sum. State the assumptions next to the number. Add fixed lead
  times that aren't effort (app store review ~1–3 days, third-party onboarding).

## Calibration

If the user reports actuals for shipped tickets, compare against the estimate and save the pattern
as a memory (e.g. "FE screens in this repo run one size larger than estimated"). Apply it next time
and say you did.
