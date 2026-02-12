---
title: How to mock HTTP APIs with Mokapi
description: Mock any HTTP API with OpenAPI specification
---

# Mocking HTTP APIs

## Quick Start: Mock an API in seconds

Run this single command to start mocking Swagger's PetStore API:

```bash
mokapi https://petstore3.swagger.io/api/v3/openapi.json
```

Open your browser and navigate to `http://localhost/api/v3/pet/12`. You'll see a generated response like:

```json
{
  "id": 12,
  "name": "Bruiser",
  "category": {
    "id": 1,
    "name": "Dogs"
  },
  "photoUrls": ["https://example.com/photo1.jpg"],
  "status": "available"
}
```

That's it! Mokapi automatically generates realistic data based on your OpenAPI specification.

## What You'll Learn

By the end of this guide, you'll know how to:
- Launch a mock HTTP API using an OpenAPI specification
- Customize responses with Mokapi Scripts
- Work with both OpenAPI 3.0 and Swagger 2.0 specifications
- Simulate different API scenarios for testing

## Prerequisites

Before you start, make sure you have:

- Mokapi installed on your system ([installation guide](/docs/get-started/installation.md))
- An OpenAPI 3.0 or Swagger 2.0 specification file (or URL)
- For custom scripts, basic TypeScript or JavaScript knowledge

## Basic Usage

### Using a Remote Specification

Point Mokapi to any publicly accessible OpenAPI specification:

```bash
mokapi https://petstore3.swagger.io/api/v3/openapi.json
```
```text box=tip
Mokapi supports both simplified syntax (`mokapi &lt;url&gt;`) and
verbose flags (`mokapi --providers-http-url &lt;url&gt;`). This guide uses the
simplified syntax for clarity.
```

**What happens:**
1. Mokapi downloads the specification
2. Starts the HTTP server on hosts and ports defined in `servers` specification (default port: 80)
3. Creates HTTP endpoints for all defined paths
4. Generates responses matching your schema definitions
5. Starts a dashboard server on `http://localhost` (default port: 8080)

The dashboard shows all available endpoints, recent requests, and response statistics.
The API server and dashboard run independently and can use different ports.

**Try it:**
```bash
curl http://localhost/api/v3/pet/12
```

### Using a Local Specification File

```bash
mokapi /path/to/your/openapi.yaml
```

This works with both `.json` and `.yaml` files.

### Behind a Proxy?

If you need to fetch specifications through a proxy server:

```bash
mokapi --providers-http-proxy http://proxy.server.com:8080 -- https://petstore3.swagger.io/api/v3/openapi.json
```

## Customizing Responses with Mokapi Scripts

To control API behavior, you can create custom scripts that define specific responses, simulate errors, or add conditional logic.

### Example: Custom Response for a Specific Pet

Let's return a specific response when requesting pet ID 12.

**Step 1:** Create a script file `petstore.ts`

```typescript tab=petstore.ts
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        // Check if the request is for pet ID 12
        if (request.key === '/pet/{petId}' && request.path.petId === 12) {
            response.data = {
                id: 12,
                name: 'Garfield',
                category: {
                    id: 3,
                    name: 'Cats'
                },
                photoUrls: [],
                status: 'available'
            }
        }
        // Other pet IDs will receive auto-generated data
    })
}
```

**Step 2:** Start Mokapi with both the spec and your script

```bash
mokapi https://petstore3.swagger.io/api/v3/openapi.json /path/to/petstore.ts
```

**Step 3:** Test the result

```bash
# Request pet 12 - returns your custom "Garfield" response
curl http://localhost/api/v3/pet/12

# Request pet 99 - returns auto-generated random data
curl http://localhost/api/v3/pet/99
```

**Before (without script):**
```json
{"id": 12, "name": "RandomName", "category": {"id": 1, "name": "Dogs"}, ...}
```

**After (with script):**
```json
{"id": 12, "name": "Garfield", "category": {"id": 3, "name": "Cats"}, ...}
```


### Testing Error Handling

**Scenario:** Your application needs to handle `404 Not Found` responses gracefully.

With Mokapi Scripts, you can customize responses for specific test scenarios.

```javascript tab=script.js
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.path.petId === 999) {
            response.statusCode = 404
        }
    })
}
```

To run the script, pass it to Mokapi when starting the server:

```bash
mokapi https://petstore3.swagger.io/api/v3/openapi.json ./script.js
```

You can test the behavior of your mocked API with the request:

```bash
curl http://localhost/api/v3/pet/999
```

### Simulating Network Delays

**Scenario:** Test how your app behaves with slow API responses.

Use Mokapi Scripts to add latency:

```javascript tab=script.js
import { on, sleep } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.path.petId === 999) {
            // delay the response by 5 seconds
            sleep('5s');
        }
    })
}
```

### Testing Different Data States

**Scenario:** Verify your UI displays different pet statuses correctly.

Mokapi generates varied data on each request - refresh to see different values for enums and optional fields.
Set specific responses for petId 1 and 2 and for all others Mokapi will respond with random data

```javascript tab=script.js
import { on, sleep } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        switch (request.path.petId) {
            case 1: // Note: path parameter petId is defined as integer in spec
                response.data = {
                    id: 1,
                    name: 'Max',
                    photoUrls: []
                }
                return
            case 2:
                response.data = {
                    id: 2,
                    name: 'Bella',
                    photoUrls: []
                }
                return
        }
    })
}
```

### Example: Stateful Interactions

Simulate creating and retrieving a pet:

```typescript
import { on } from 'mokapi'

let createdPets = new Map()

export default function() {
    on('http', (request, response) => {
        // Handle POST /pet (create)
        if (request.key === '/pet/{petId}' && request.method === 'POST') {
            const newPet = request.body
            createdPets.set(newPet.id, newPet)
            response.statusCode = 201
            response.data = newPet
        }
        
        // Handle GET /pet/{petId} (retrieve)
        if (request.key === '/pet/{petId}' && request.method === 'GET' && createdPets.has(request.path.petId)) {
            response.data = createdPets.get(request.path.petId)
        }
    })
}
```

## Understanding Request Matching

When writing Mokapi Scripts, you'll often need to identify which API endpoint was called.

### Using `request.key`

The `request.key` property contains the path pattern from your OpenAPI specification:
```typescript
// OpenAPI spec defines: /pet/{petId}
// User requests: http://localhost/api/v3/pet/12

on('http', (request, response) => {
    console.log(request.key)        // "/pet/{petId}"
    console.log(request.url.path)       // "/api/v3/pet/12"
    console.log(request.operationId) // "getPetById" defined in OpenAPI spec
})
```

**Best Practice:** Always check `request.key` to match the correct endpoint:
```typescript
// ✅ Reliable - matches the OpenAPI path pattern
if (request.key === '/pet/{petId}') {
    // Your logic here
}

// ❌ Fragile - breaks if base path changes
if (request.url.path.startsWith('/api/v3/pet/')) {
    // Your logic here
}
```

## Working with Swagger 2.0

Mokapi fully supports Swagger 2.0 specifications. When you provide a Swagger 2.0 file,
Mokapi automatically converts it to OpenAPI 3.0 internally.

### Schema Reference Translation

**Swagger 2.0:**
```yaml
definitions:
  Pet:
    type: object
```

**OpenAPI 3.0 (how Mokapi sees it):**
```yaml
components:
  schemas:
    Pet:
      type: object
```

Mokapi handles this translation automatically - you don't need to modify your specs.

### What This Means for You

**✅ Your Swagger 2.0 specs work immediately** - no conversion needed  
**✅ Reference resolution handled automatically** - Mokapi translates between formats  
**✅ Schema paths differ** - Mokapi transforms the path

## Best Practices

### ✅ Start with Auto-Generated Data

Let Mokapi handle the basics. Adjust only data for you specific need.

```javascript
import { on, sleep } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.key === '/pet/{petId}' && request.path.petId === 10) {
            // Instead of replacing entire responses, you can modify specific fields:
            response.data.name = 'Garfield'
        }
    })
}
```

### ✅ Version Your Specifications

Keep your OpenAPI specs and Mokapi Scripts in version control alongside your code.

### ✅ Use Scripts for Edge Cases

Focus your scripts on:
- Error scenarios (404, 500, validation errors)
- Authentication/authorization testing
- Specific business logic you want to test
- Stateful workflows

## Next Steps

**Ready for more advanced features?**

- [Test Data Generation](/docs/get-started/test-data.md) - Create realistic, varied test data
- [Mokapi Scripts Guide](/docs/javascript-api/overview) - Full scripting reference and examples
