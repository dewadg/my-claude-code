---
name: api-tester
description: Perform HTTP requests to a server, read the responses, verify behavior as defined by the user. Use when you need to test APIs.
allowed-tools: curl, cat, jq
---

# API Tester

## Instructions

1. Read thoroughly the user's instruction regarding the steps, test cases, configurations/variables
2. Use curl to perform HTTP requests
3. Verify the API behavior and report the results

## Examples

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

