---
name: hurl-run
description: >
  Execute HTTP requests with the hurl CLI — run .hurl test files, or fire one-off requests with no
  file via stdin — and read back status, body, headers, latency, and assert failures. ALWAYS use
  when the user wants to hit, call, probe, or test an HTTP API endpoint, or run an API test suite —
  e.g. "test the /users endpoint", "does this API return 401 without a token", "call the staging
  endpoint with this payload", "run the hurl tests", "run users.hurl against staging", "check what
  this route responds with". To write a reusable .hurl scenario file, use hurl-write. Not for Go unit
  tests (use golang-unit-test) or browser UI flows.
allowed-tools: Read, Glob, Grep, Bash(hurl *), Bash(hurlfmt *), Bash(jq *)
---

How to execute HTTP against a running server with `hurl` and read the result. Two modes — **ad-hoc**
(no file, request piped on stdin) and **file** (run `.hurl` scenarios written via `hurl-write`). In
both, an observation is complete only with **status + body + latency**; for scenarios, the
**exit code** is the verdict.

## Before calling

Confirm the server is up. Connection refused is a missing precondition, not a test failure —
reporting it as a failed scenario is a misdiagnosis.

```bash
echo 'GET http://127.0.0.1:8787/health' | hurl --connect-timeout 5 --max-time 10
```

Set `--max-time` on anything that might hang. Base URLs, credentials, and fixture IDs may be defined
in the project (`variables.yaml`, `tests/hurl/*.env`) — read them instead of inventing hosts or IDs.

## Ad-hoc (no file)

Pipe a hurl entry on stdin. With no asserts, hurl prints the last response body:

```bash
echo 'GET http://127.0.0.1:8787/health' | hurl
```

With a body, headers, and a status expectation — use a heredoc; add `-i` for response headers:

```bash
hurl -i <<'EOF'
POST http://127.0.0.1:8787/login
{"username":"qa","password":"bad"}
HTTP 401
EOF
```

Status + latency + body in one machine-readable shot (`timings` are **microseconds**):

```bash
echo 'GET http://127.0.0.1:8787/health' | hurl --json \
  | jq -c '.entries[0].calls[0] | {status: .response.status, total_us: .timings.total}'
```

Add `HTTP <code>` (and `[Asserts]`) to any ad-hoc request whose answer you are judging — then the
exit code tells pass/fail instead of eyeballing.

## File mode (scenarios)

```bash
hurl --test --variables-file tests/hurl/local.env --secret password="${PASSWORD}" tests/hurl/users.hurl
```

- `--test` → summary per file, no body dump, files run in parallel.
- `--variables-file` → non-secret env values (`key=value` lines). Switch env by switching the file.
- `--secret NAME=VALUE` → secret variable; redacted as `***` in all output and exports. Never put
  the literal in the command — pass it from a shell variable.
- `HURL_VARIABLE_<name>=value` env vars also inject variables
  (`HURL_VARIABLE_base_url=http://127.0.0.1:8787 hurl --test ...`).
- Whole suite: `hurl --test --variables-file tests/hurl/local.env --glob 'tests/hurl/*.hurl'`.
- CI/report artifact: add `--report-junit out/junit.xml` (also `--report-html DIR`, `--report-tap FILE`).

## Reading the result

**Exit code is the verdict:**

| exit | meaning | report as |
|---|---|---|
| 0 | all asserts passed | pass |
| 2 | file failed to parse | broken test file — fix via `hurl-write`, not an API defect |
| 3 | runtime: connection refused, timeout, undefined `{{var}}` | precondition/config problem, not a defect |
| 4 | assert failed (status, header, body) | the finding — read the error |

An assert failure points at the line and shows actual vs expected:

```
error: Assert status code
  --> -:2:6
 2 | HTTP 200
   |      ^^^ actual value is <404>
```

For the full response behind a failure, rerun with `--error-format long` (prints the response
headers + body and the equivalent curl command). Deeper: `-v` (request/response headers, per-entry),
`--very-verbose` (adds libcurl logs).

Narrow a long file while debugging: `--to-entry N` stops after entry N, `--from-entry N` starts there
(entries numbered from 1; captures from skipped entries are not available).

```bash
hurl --variables-file tests/hurl/local.env --secret password="${PASSWORD}" --to-entry 2 --json tests/hurl/users.hurl \
  | jq -c '{success, entries: [.entries[] | {index, status: .calls[0].response.status}]}'
```

Exploring past the first failure: `--continue-on-error`.

## Gotchas (verified on hurl 8.0.1)

- **Without `HTTP <code>`, nothing is asserted** — a 404/500 exits 0. Always state the expected
  status when the answer matters.
- **JSON bodies get `Content-Type: application/json` automatically** — unlike `curl -d`, no header
  needed.
- **`--curl FILE` exports replayable curl commands, but secrets are written as `***`** — the
  export is for showing a repro, not for re-running with real credentials.
- **Undefined variables are exit 3, not 4.** A missing `--variables-file` or `--secret` looks like a
  runtime error, not a failed test — check the error text before raising a defect.
- **`--json` timings are in microseconds** (`748` = 0.75 ms).
- **zsh**: `$PIPESTATUS` is `$pipestatus` — when piping hurl into `jq`, capture hurl's exit with a
  separate run or `set -o pipefail`.
