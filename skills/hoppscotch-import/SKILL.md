---
name: hoppscotch-import
description: Generate a Hoppscotch-importable collection JSON from any source of endpoint definitions — OpenAPI/Swagger spec, route-registration code, a Postman collection, a HAR file, or a plain list of endpoints. Use this whenever the user asks to "create a hoppscotch import file", "generate a hoppscotch collection", "export endpoints to hoppscotch", "turn this OpenAPI spec into a Hoppscotch collection", or wants API endpoints importable into Hoppscotch. Not for making requests against a running API (use http-request) or for authoring API documentation.
allowed-tools: Read, Glob, Grep, Write, Edit, WebFetch, Bash(find *), Bash(grep *), Bash(python3 *)
---

# Hoppscotch import

Generate a Hoppscotch **native collection JSON** from any source of endpoint definitions — a spec file, route-registration code, an existing collection, a HAR capture, or the user's own description (see Input sources). It works on any repo or prompt: normalize whatever source to one intermediate list of endpoints, then emit Hoppscotch v1 JSON.

Two principles hold throughout. **v1 is the target** — it imports cleanly into every Hoppscotch build, while newer collection formats (v2/v3) do not. And the import file is **derived**, never hand-edited: when the source changes, regenerate it.

## Format authority

Hoppscotch's import format schema is defined by the hoppscotch-cli e2e fixtures — fetch and treat as canonical:

```
https://raw.githubusercontent.com/hoppscotch/hoppscotch/main/packages/hoppscotch-cli/src/__tests__/e2e/fixtures/collections/coll-v1-req-v1.json
https://raw.githubusercontent.com/hoppscotch/hoppscotch/main/packages/hoppscotch-cli/src/__tests__/e2e/fixtures/collections/collection-level-auth-headers-coll.json
```

`coll-v1-req-v1.json` = collection + request v1 shape. `collection-level-auth-headers-coll.json` = `basic` auth + body fields. v1 is the most broadly importable target — newer collection formats (v2/v3) exist but v1 imports cleanly everywhere. If unsure about a field, grep the fixtures dir:

```
https://api.github.com/repos/hoppscotch/hoppscotch/git/trees/main?recursive=1
# → filter paths containing "fixtures/collections/"
```

## Native v1 schema (verified)

```jsonc
// collection root
{
  "v": 1,
  "name": "<collection>",
  "folders": [ /* nested collections, same shape as root minus top-level */ ],
  "requests": [ /* request v1 objects */ ]
}

// folder (nested collection): { "v":1, "name":"...", "folders":[], "requests":[...] }

// request v1
{
  "v": "1",
  "name": "<human label>",
  "method": "POST",
  "endpoint": "<<base>>/path",
  "params":  [ { "key": "q", "value": "1", "active": true } ],
  "headers": [ { "key": "Content-Type", "value": "application/json", "active": true } ],
  "body": { "contentType": "application/json", "body": "<raw JSON string>" },
  "auth": { /* see variants */ },
  "preRequestScript": "",
  "testScript": "",
  "requestVariables": []
}
```

Key rules:
- `v` is `1` (number) at collection/folder, `"1"` (string) at request.
- `body.body` is the **raw JSON as a string**, not an object. `contentType` is `"application/json"`, or `null` for none → `{ "contentType": null, "body": null }`.
- No-body requests: `body: { "contentType": null, "body": null }`.
- Always emit `preRequestScript: ""`, `testScript: ""`, `"requestVariables": []`.

### auth variants

```jsonc
{ "authType": "none",   "authActive": true }
{ "authType": "basic",  "authActive": true, "username": "<<user>>", "password": "<<pass>>" }
{ "authType": "bearer", "authActive": true, "token": "<<token>>" }
{ "authType": "inherit","authActive": true }                       // inherit parent folder/collection auth
```

If auth is a custom header scheme (not standard basic/bearer — e.g. signature-based), set `authType: "none"` and carry the credentials as active `headers` entries.

### variables

Hoppscotch env vars are referenced inline as `<<name>>` (double angle brackets). Use `<<base>>` for the host prefix in every `endpoint`. Promote every dynamic value (path params, secrets, query inputs) to a `<<var>>` rather than hardcoding — the user sets them once in their Hoppscotch environment.

## Input sources

Pick the richest source available. Prefer in this order; ask the user which if several are present or none is obvious.

| Source | How to read | What it gives you | Gaps to fill |
|---|---|---|---|
| **OpenAPI / Swagger** (`openapi*.yaml/json`, `swagger*.json`) | parse paths + components | path, method, name, body, params, auth, examples, tags | usually none |
| **Postman collection** (`*.postman_collection.json`) | parse `item[]` recursively | everything; near 1:1 | remap auth + body to Hoppscotch shape |
| **HAR** (`*.har`) | parse `log.entries[].request` | observed requests incl. real headers/body | sample only — not exhaustive; names inferred from path |
| **Route-registration code** | grep framework router calls (Fiber `app.Get/Post`, gin `r.GET`, echo `e.POST`, chi `r.Get`, gorilla `r.HandleFunc`, net/http `mux.HandleFunc`, express `app.get`, fastify `fastify.post`, Spring `@RequestMapping`, etc.) | path + method | auth, body, params, names — ask the user or infer from handler |
| **Plain list from user** | user pastes/types `METHOD path` lines or prose | minimal skeleton | everything else — ask user for bodies/auth as needed |

Auto-detect with:
```bash
find . -type f \( -name 'openapi*.yaml' -o -name 'openapi*.yml' -o -name 'openapi*.json' \
                 -o -name 'swagger*.yaml' -o -name 'swagger*.json' \
                 -o -name '*.postman_collection.json' -o -name '*.har' \) \
        -not -path '*/node_modules/*' -not -path '*/.git/*'
```
For route-code detection, grep the relevant router-registration files (the composition root / `routes.go` / handler dirs).

## Process

1. **Determine the input source.** Run the auto-detect above. If one spec/collection/HAR is found, use it. If several, or only route code, or nothing machine-readable — ask the user which source to use, or accept a pasted endpoint list. Normalize every source into one intermediate list: `{ method, path, name, params[], headers[], body?, auth? }` per endpoint.

2. **Confirm scope with the user:**
   - **Output path.** Default = co-locate next to the source (e.g. `foo/openapi.yaml` → `foo/hoppscotch.json`). Ask if the user wants a different location/filename.
   - **Endpoints to exclude.** Default drop health/liveness/readiness probes (`/healthz`, `/health`, `/livez`, `/ready`, `/readyz`). Confirm — they may also want docs/spec/static endpoints excluded or kept.
   - **Collection name** (root `name`) — default to service name (spec `info.title`, repo dir, or app name).
   - **Missing fields.** If the source lacks auth/body/params (route code, plain list), ask the user how endpoints authenticate and whether bodies are needed, rather than guessing silently.

3. **Map each endpoint to a request v1 object.** One request per `(path, method)` pair. Path params (`{id}`, `:id`, `<id>`) → `<<var>>` placeholders in the endpoint string (e.g. `/tenants/{id}` → `<<base>>/tenants/<<tenant_id>>`).

4. **Group by folder.** If the source carries tags/groups (OpenAPI `tags`, Postman folders), map them to top-level `folders`. Else a flat `requests` list or a single folder is fine.

5. **Seed bodies.** Prefer source examples (OpenAPI `examples`, Postman `raw`, HAR `postData`). Else synthesize a minimal valid body from the schema's required fields, filling each with a type-appropriate placeholder (strings → `"example"`, numbers → `0`, booleans → `false`). GET/DELETE with no body → `{ "contentType": null, "body": null }`.

6. **Write + validate.** Write the JSON to the agreed output path, then validate (always, before handing off):
   ```bash
   python3 -c 'import json;p="<OUTPUT_PATH>";d=json.load(open(p));n=sum(len(f["requests"]) for f in d["folders"])+len(d["requests"]);print("OK",len(d["folders"]),"folders,",n,"total req");[print(" ",f["name"],"->",len(f["requests"])) for f in d["folders"]]'
   ```
   (Replace `<OUTPUT_PATH>` with the agreed path. `python3 -m json.tool <path>` is an alternate parse check.)

## Gotchas

- `body.body` MUST be a string. Embedding an object there silently breaks import.
- Mixing collection/request `v` types (number vs string) breaks the importer — use exactly the shapes above.
- Hoppscotch has no native signature/HMAC auth type. For signature-scheme auth, leave `authType: "none"` and put the key/signature headers as `<<var>>`-valued `headers`. A pre-request script can compute the signature if full automation is needed.
- `<<base>>` must not include a trailing slash; paths start with `/`. `<<base>>/webhook`, not `<<base>>webhook`.
- Route-code scans see path+method only. Auth applied by middleware (not at the route call) won't be visible — infer from the framework's middleware stack or ask the user.
- HAR is a capture, not a spec — it only contains requests actually made. Do not claim it is exhaustive; say what was captured.
- The import file is **derived** from its source. When the source changes, regenerate — do not hand-edit the import file.
