---
name: http-request
description: >
  Make HTTP requests against a running server with curl and read back the status code, response body,
  and latency. ALWAYS use when the user wants to hit, call, probe, or test an HTTP API endpoint —
  e.g. "test the /users endpoint", "does this API return 401 without a token", "call the staging
  endpoint with this payload", "check what this route responds with". Not for Go unit tests (use
  golang-unit-test) and not for testing UI flows in a browser.
allowed-tools: Read, Glob, Bash(curl *), Bash(jq *)
---

How to make an HTTP call and read the result. Every request goes out through `curl` and comes back
with three things observed: **status code, response body, latency**. A call that only prints the body
is not a complete observation — you cannot tell a 200 from a 500 that happened to return JSON.

## The canonical invocation

```bash
curl -sS -X POST "https://api.example.com/users" \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}' \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'
```

Flag by flag, because each one is load-bearing:

- `-s` silences the progress meter, which otherwise corrupts the output you are trying to read.
- `-S` keeps real errors (DNS failure, connection refused) visible despite `-s`.
- `-w` appends the status code and total latency after the body. This is what makes the observation
  complete — without it you see a body and have to guess the status.
- **Do not use `-f`/`--fail`.** It makes curl discard the response body on a 4xx/5xx, which throws
  away the error message you most needed to read.

`%{time_total}` is the full wall-clock time of the request. If you need to distinguish "server is
slow" from "connection setup is slow", add `%{time_connect}` and `%{time_starttransfer}` to the same
`-w` string.

## Methods

```bash
# GET — fetch
curl -sS "https://api.example.com/users" \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# GET with query params — let curl encode them, don't hand-build the string
curl -sS -G "https://api.example.com/users" \
  --data-urlencode "q=john doe" \
  --data-urlencode "limit=10" \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# POST — create
curl -sS -X POST "https://api.example.com/users" \
  -H "Content-Type: application/json" \
  -d '{"name": "John"}' \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# PUT — replace whole resource
curl -sS -X PUT "https://api.example.com/users/123" \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated", "email": "john@example.com"}' \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# PATCH — update part of resource
curl -sS -X PATCH "https://api.example.com/users/123" \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated"}' \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# DELETE — remove
curl -sS -X DELETE "https://api.example.com/users/123" \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# HEAD — headers only, no body
curl -sS -I "https://api.example.com/users"

# OPTIONS — allowed methods
curl -sS -X OPTIONS "https://api.example.com/users" -i
```

Use `-i` to include response headers alongside the body when the headers are the thing under test
(`Location` after a create, `Set-Cookie`, rate-limit headers, CORS).

## Auth

```bash
# Bearer token
curl -sS "https://api.example.com/me" \
  -H "Authorization: Bearer ${TOKEN}" \
  -w '\n---\nstatus: %{http_code}\nlatency: %{time_total}s\n'

# API key header
curl -sS "https://api.example.com/me" -H "X-API-Key: ${API_KEY}"

# Basic auth
curl -sS -u "${USER}:${PASS}" "https://api.example.com/me"
```

Always pass credentials as a shell variable, never a literal in the command. A literal token lands in
shell history and in the transcript.

## Bodies and payloads

```bash
# Body from a file — use this when the JSON is long or has quoting that fights the shell
curl -sS -X POST "https://api.example.com/users" \
  -H "Content-Type: application/json" \
  -d @payload.json

# Form-encoded
curl -sS -X POST "https://api.example.com/login" \
  -d "username=john" -d "password=${PASS}"

# Multipart file upload
curl -sS -X POST "https://api.example.com/avatar" \
  -F "file=@./avatar.png" \
  -F "user_id=123"
```

`-d` sets `Content-Type: application/x-www-form-urlencoded` by default, so a JSON body **must** carry
an explicit `-H "Content-Type: application/json"` or the server will reject it — this is the single
most common reason a hand-written call 400s.

## Reading the response

Pipe the body to `jq` to pull out a field. Note that `-w` output is not JSON, so send it to stderr or
drop it when piping:

```bash
# Extract one field
curl -sS "https://api.example.com/users/123" | jq -r '.name'

# Pretty-print the whole body
curl -sS "https://api.example.com/users" | jq

# Status code only, body discarded
curl -sS -o /dev/null -w '%{http_code}\n' "https://api.example.com/users"

# Status and body captured separately, both usable
body=$(curl -sS -w '\n%{http_code}' "https://api.example.com/users/123")
status="${body##*$'\n'}"
body="${body%$'\n'*}"
echo "status: ${status}"
echo "${body}" | jq
```

If `jq` fails to parse, the response was not JSON — print the raw body before concluding anything.
An HTML error page or an empty body is itself the finding.

## Configuration

Base URLs, credentials, and fixture data may be defined in a `variables.yaml` in the project. Read it
before constructing calls and use the values from it rather than inventing hosts or IDs.

## Before calling

Confirm the server is actually up. A connection-refused error is not a test failure, it is a missing
precondition, and reporting it as a failed scenario is a misdiagnosis:

```bash
curl -sS -o /dev/null -w 'status: %{http_code}\n' --max-time 5 "https://api.example.com/health"
```

Set `--max-time` on any call that might hang so a dead server surfaces as a timeout instead of
blocking indefinitely.
