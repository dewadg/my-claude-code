---
name: http-api-test
description: Perform HTTP requests to a server, read the responses, verify behavior as defined by the user. ALWAYS use when the user wants to test HTTP API endpoints.
allowed-tools: Bash(curl *) Bash(jq *)
effort: low
---

Perform HTTP requests to a server, read the responses, verify behavior as defined by the user.

## Input
HTTP endpoint with server running AND the list of scenario, request body/params to cover.

## Output

### Output during implementation
```
## Testing endpoint: <HTTP endpoint URL>

Working on task 3/7: <scenario>
[...implementation happening...]
✓ Scenario complete

Working on task 4/7: <scenario>
[...implementation happening...]
✓ Scenario complete
```

### Output when completed
```
## HTTP API endpoint testing completed

**Endpoint:** <full-endpoint-url>
**Progress:** 7/7 scenario covered ✓

### Completed This Session
- [x] Scenario 1
- [x] Scenario 2
...
```

## Steps

1. **Prepare**
    - Check if the server is running
    - Generate request bodies/params for each scenario

2. **Go through each scenario**
    - Hit the endpoint with request body/params
    - Observe the status code and the response body
    - Measure the latency

## Making HTTP requests

GET (fetch data)
```bash
curl https://api.example.com/users
```

POST (create new data)
```bash
curl -X POST https://api.example.com/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'
```

PUT (update entire resource)
```bash
curl -X PUT https://api.example.com/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated", "email": "john@example.com"}'
```

PATCH (update part of resource)
```bash
curl -X PATCH https://api.example.com/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated"}'
```

DELETE (remove data)
```bash
curl -X DELETE https://api.example.com/users/123
```

HEAD (get headers only)
```bash
curl -I https://api.example.com/users
```

OPTIONS (check allowed methods)
```bash
curl -X OPTIONS https://api.example.com/users
```
