---
name: test-scenario
description: >
  Generate test scenarios in Given/When/Then (Gherkin-style) format and present them as a single
  markdown table with columns ID | Scenario | Given | When | Then. ALWAYS use when the user asks to
  write test cases, test scenarios, test plans, acceptance criteria, Given/When/Then, BDD scenarios,
  or "how should we test this feature" — and when scoping a feature BEFORE development (pre-dev,
  defines what to build) or verifying one AFTER (post-dev, defines what to check). Covers happy
  path, alternate paths, edge cases, and error/negative paths. Not for executing the tests (the
  the agent executing the tests owns that) or writing unit test code (the language engineering agent does).
allowed-tools: Read, Grep, Glob
---

# Test Scenario

Produce a complete Given/When/Then scenario set as one markdown table. The goal: a reviewer or
tester can take the table and know exactly what to verify. Pre-development it defines what to
build; post-development it defines what to check.

## When invoked

1. Read the feature/bug context — ticket, spec, acceptance criteria, or the code that implements it.
2. Identify the actor (user, system, API consumer) and the entry points (endpoint, screen, CLI).
3. Enumerate paths: happy, alternate, edge, error/negative.
4. Emit the table in the Output Format below.

## Coverage — every scenario set must include

- **Happy path** — the primary flow that must work.
- **Alternate paths** — valid inputs/variations that take a different successful route.
- **Edge cases** — boundary values, empty input, concurrency/timing, large payloads, unicode, off-by-one.
- **Error / negative paths** — invalid input, missing auth, wrong role, conflicts, not-found,
  server errors, idempotency violations.

If a category has no applicable scenario, say so explicitly rather than silently dropping it.
Silent gaps are the most common review failure — an empty category with a reason is more useful
than a category that quietly disappears.

## Output format

A single markdown table. Columns exactly:

| ID | Scenario | Given | When | Then |
|----|----------|-------|------|------|

- **ID** — `TC-<group>-<n>` (e.g. `TC-HAPPY-1`, `TC-EDGE-3`, `TC-ERR-2`). Groups: `HAPPY`, `ALT`,
  `EDGE`, `ERR`. Stable IDs so whichever agent executes the table can reference a row and report
  pass/fail per ID.
- **Scenario** — one-line name, intent-focused ("User logs in with valid credentials").
- **Given** — preconditions/state before the action (data, auth, feature flag, environment).
- **When** — the single action or trigger (request, click, event). One action per row; split
  compound "when" clauses into separate rows so each is independently verifiable.
- **Then** — the expected outcome: response, state change, side effect, emitted event. Include
  status code or error message where applicable. One assertion per Then where possible.

If a precondition is shared by many rows, state it once in a lead-in line above the table rather
than repeating it in every Given cell.

## Example

Lead-in: all rows assume a user account exists with a known password, on the staging environment.

| ID | Scenario | Given | When | Then |
|----|----------|-------|------|------|
| TC-HAPPY-1 | Login with valid credentials | user exists | POST /login valid body | 200, JWT returned, `last_login` updated |
| TC-ALT-1 | Login with username instead of email | user exists | POST /login using username field | 200, JWT returned |
| TC-EDGE-1 | Password at bcrypt max length | user password is exactly 72 chars | POST /login with 72-char password | 200, JWT returned |
| TC-EDGE-2 | Email at boundary unicode | email contains é + emoji | POST /login | 200, JWT returned |
| TC-ERR-1 | Wrong password | user exists | POST /login wrong password | 401, "invalid credentials", no JWT |
| TC-ERR-2 | Missing email field | — | POST /login with no email | 400, validation error on `email` |

## Rules

- Anchor each row to the real contract — read the endpoint/schema/component code so Given/Then
  reflect actual status codes, field names, and side effects, not guesses.
- Pre-development (no code yet): derive Given/Then from the spec/ticket; list assumptions openly.
- Post-development (code exists): derive Given/Then from the actual implementation; cite
  `file:line` in a note under the table for the key paths so execution can trace back.

After the table, list assumptions, open questions, and any path that needs clarification before
the table is executed.
