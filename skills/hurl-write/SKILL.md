---
name: hurl-write
description: >
  Author Hurl files (.hurl) — plain-text HTTP test scenarios: chained requests with captures, status
  and body asserts, variables and secrets. ALWAYS use when the user wants an HTTP endpoint test
  scenario written down as a re-runnable file — e.g. "write a hurl test for the /users API", "turn
  these test scenarios into hurl", "add a hurl file covering the login flow", "create an API
  regression suite", "script the create → read → delete flow". Also use when converting a
  Given/When/Then table, OpenAPI spec, curl commands, or route code into .hurl files. This skill
  writes the file only — to execute a file or fire a one-off request, use hurl-run. Not for Go unit
  tests (use golang-unit-test) or browser UI flows.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash(hurl *), Bash(hurlfmt *)
---

How to write a `.hurl` file that is a **test scenario**, not a request log: every entry asserts what
it expects, data flows between entries via captures, and nothing environment-specific is hardcoded.
Syntax reference: https://hurl.dev/docs/hurl-file.html. Execution is owned by `hurl-run`.

## Where files go

1. Project already has `.hurl` files → put new ones next to them, same naming.
2. Otherwise → `tests/hurl/{feature}.hurl`, plus `tests/hurl/{env}.env` variables files
   (`local.env`, `staging.env`). Secrets never go in these files — see Variables.

One file per feature or flow (`users.hurl`, `auth.hurl`). Each file must run standalone from a clean
state: it logs in itself, creates the data it reads, and deletes what it created.

## File shape

A file is a sequence of **entries**. Each entry = request + (optional) response spec. Entries run in
order and share a cookie store and captured variables.

```hurl
# Users API — create/read/delete flow plus auth + validation error paths.
# Vars: base_url, password (secret)

# TC-HAPPY-1: health is up
GET {{base_url}}/health
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"
duration < 1000

# TC-HAPPY-2: login with valid credentials returns a token
POST {{base_url}}/login
{
  "username": "qa",
  "password": "{{password}}"
}
HTTP 200
[Captures]
token: jsonpath "$.token"
[Asserts]
jsonpath "$.token" isString
jsonpath "$.expires_in" > 0

# TC-ERR-1: login with wrong password is rejected
POST {{base_url}}/login
{
  "username": "qa",
  "password": "wrong"
}
HTTP 401
[Asserts]
jsonpath "$.error" == "invalid credentials"

# TC-HAPPY-3: create user returns 201 + Location
POST {{base_url}}/users
Authorization: Bearer {{token}}
{
  "name": "John",
  "email": "{{newUuid}}@test.com"
}
HTTP 201
[Captures]
user_id: jsonpath "$.id"
location: header "Location"
[Asserts]
header "Location" matches "^/users/u\\d+$"
jsonpath "$.email" endsWith "@test.com"
jsonpath "$.created_at" isIsoDate

# TC-HAPPY-4: created user is readable at its Location
GET {{base_url}}{{location}}
Authorization: Bearer {{token}}
HTTP 200
[Asserts]
jsonpath "$.id" == "{{user_id}}"
jsonpath "$.name" == "John"

# TC-ERR-2: missing email → 422 with field
POST {{base_url}}/users
Authorization: Bearer {{token}}
{
  "name": "NoEmail"
}
HTTP 422
[Asserts]
jsonpath "$.field" == "email"

# TC-ERR-3: no token → 401
GET {{base_url}}/users/{{user_id}}
HTTP 401

# TC-HAPPY-5: list contains the created user
GET {{base_url}}/users
[Query]
limit: 10
HTTP 200
[Asserts]
jsonpath "$.items" count >= 1
jsonpath "$.items[?(@.id == '{{user_id}}')].id" == "{{user_id}}"

# TC-HAPPY-6: delete then 404
DELETE {{base_url}}/users/{{user_id}}
Authorization: Bearer {{token}}
HTTP 204

GET {{base_url}}/users/{{user_id}}
Authorization: Bearer {{token}}
HTTP 404
```

Rules the example encodes:

- **Header comment** names the flow and lists every variable the file needs.
- **One `# <ID>: <intent>` comment per scenario.** When a `test-scenario` table exists, use its IDs
  (`TC-HAPPY-1`, `TC-ERR-2`, …) so a failure maps straight back to a row. One scenario may span
  several entries (TC-HAPPY-6).
- **Every entry has `HTTP <status>`.** Without it the entry asserts nothing and a 500 passes. Use
  `HTTP *` only for setup calls whose status genuinely does not matter.
- **Status line + body asserts both.** `HTTP 200` alone passes on a 200 with a wrong body.
- **Cover the negative paths** — wrong credentials, missing auth, validation, not-found — not just
  the happy flow.
- **Clean up** what the file created, so reruns and parallel runs don't collide.

## Request

```hurl
METHOD {{base_url}}/path
Header-Name: value
[Query]               # query params, URL-encoded for you
q: john doe
[Form]                # application/x-www-form-urlencoded
username: qa
[Multipart]           # multipart/form-data; file path relative to the .hurl file
note: hi
file: file,fixture.txt; text/plain
[Cookies]
theme: light
[BasicAuth]
bob: {{password}}
[Options]             # per-entry CLI options
retry: 3
retry-interval: 200ms
skip: true
<body>                # JSON/XML literal, ```multiline```, `oneline`, or file,payload.json;
```

- A JSON body literal sets `Content-Type: application/json` automatically — no header needed.
- Big or reused payloads → `file,payload.json;` (path relative to the .hurl file).
- `[Options] retry: N` re-runs the entry until its asserts pass — use it to poll async work
  (job status, eventual consistency) instead of sleeping.

## Response

```hurl
HTTP 201
Location: /users/u1          # implicit header assert (exact match)
[Captures]
name: <query> [filters]
[Asserts]
<query> [filters] <predicate> <value>
```

**Queries:** `status`, `header "X"`, `cookie "name"`, `body`, `bytes`, `jsonpath "$.a"`,
`xpath "//h1"`, `regex "id=(\d+)"`, `variable "name"`, `duration`, `url`, `sha256`, `certificate "Expire-Date"`.

**Predicates:** `==` `!=` `>` `>=` `<` `<=` `startsWith` `endsWith` `contains` `matches`
`exists` `isEmpty` `isString` `isInteger` `isFloat` `isNumber` `isBoolean` `isList` `isObject`
`isCollection` `isIsoDate` `isDate`. Negate with `not` (`jsonpath "$.x" not exists`).

**Filters** (chain between query and predicate): `count`, `first`, `last`, `nth 0`, `toInt`,
`toString`, `toFloat`, `split ","`, `replace "a" "b"`, `regex /(\d+)/`, `jsonpath "$.x"`,
`urlDecode`, `base64Decode`, `daysAfterNow`, …

Captures are only visible to **later** entries, and `[Asserts]` in the same entry can read them with
`variable "name"`.

## Variables and secrets

- Reference as `{{name}}`. Inside JSON strings, quote them: `"{{user_id}}"`; numbers unquoted:
  `{{count}}`.
- Non-secret env values (`base_url`, fixture IDs) → `{env}.env` file, `key=value` per line.
- Secrets (passwords, tokens, API keys) → **never in the repo**; the runner passes them with
  `--secret`, which also redacts them in output. List them in the header comment.
- Unique data per run: `{{newUuid}}`, `{{newDate}}` — avoids conflicts on reruns.

## Gotchas (verified on hurl 8.0.1)

- **`includes` is deprecated** — hurl warns; use `contains`.
- **JSONPath that matches one node returns a scalar, not a list.** `jsonpath "$.items[*].id"
  contains "u1"` works with ≥2 items and fails with exactly 1 ("types between actual and expected
  are not consistent"); `count` on a single filtered match errors with "invalid filter input type".
  To check membership, filter to the element and compare a scalar:
  `jsonpath "$.items[?(@.id == '{{user_id}}')].id" == "{{user_id}}"`.
- **Regex inside a string needs doubled backslashes** (`"^/users/u\\d+$"`); `/…/` regex literals
  don't (`matches /^u\d+$/`).
- **Durations take units** — `hurlfmt` normalizes `retry-interval: 200` to `200ms`; write the unit.
- **`[Form]` sends form-encoding.** An endpoint expecting JSON will 400 on it — use a JSON body.

## Validate before handing off

```bash
hurlfmt --check tests/hurl/users.hurl     # exit 0 clean, 2 parse error, 3 would reformat
hurlfmt --in-place tests/hurl/users.hurl  # apply canonical formatting
```

A parse error here is a syntax bug in the file, not in the API. Then run it with the `hurl-run`
skill — a file that was never executed against the target is not done.
