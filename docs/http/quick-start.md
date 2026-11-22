---
title: "HTTP API Mocking: Mock the Swagger Petstore with Mokapi"
description: Start a mocked REST API from an OpenAPI specification, send requests, inspect them in the dashboard and return your own response with a Mokapi Script.
---
# Mocking the Petstore REST API

## Overview

This guide mocks Swagger's Petstore with Mokapi. By the end, you'll have:

1. Started Mokapi with the Petstore OpenAPI specification.
2. Sent a request and received a generated response.
3. Inspected the request in the dashboard.
4. Written a Mokapi Script that returns your own data for one pet.
5. Returned an error for a pet that doesn't exist.

If Mokapi isn't installed yet, see the [installation guide](/docs/get-started/installation.md).

## Start Mokapi

Pass the URL of the specification to Mokapi:

```bash
mokapi https://raw.githubusercontent.com/marle3003/mokapi/main/examples/swagger/petstore-3.0.json
```

Or run it in Docker:

```bash
docker run --rm -p 80:80 -p 8080:8080 mokapi/mokapi:latest https://raw.githubusercontent.com/marle3003/mokapi/main/examples/swagger/petstore-3.0.json
```

The specification defines the server `http://localhost/api/v3`, so Mokapi serves the API on port
80 under the path `/api/v3`. The dashboard listens on port 8080.

``` box=info
When you use Mokapi behind a corporate proxy, you probably need to skip TLS verification when
Mokapi fetches the specification: `--providers-http-tls-skip-verify` or, in Docker,
`-e MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY=true`.
```

## Send a Request

Request a pet:

```bash
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```

Mokapi validates the request against the specification and generates a response that matches
the schema of `getPetById`. The values differ with every request, for example:

```json
{
  "id": 4,
  "name": "Waldo",
  "category": {
    "id": 1,
    "name": "Dogs"
  },
  "photoUrls": [
    "http://www.chiefcontent.name/cross-media/out-of-the-box/metrics"
  ],
  "tags": [
    { "id": 7100881306113605000, "name": "liIe4zU0aX0WUp" }
  ],
  "status": "pending"
}
```

The Petstore returns JSON and XML. Without an `Accept` header, Mokapi picks one of the content
types at random.

## Inspect the Request in the Dashboard

1. Open the dashboard, by default at http://localhost:8080.
2. Click **HTTP** in the sidebar and then **Swagger Petstore - OpenAPI 3.0**.
3. The **Paths** sub tab lists all endpoints with their number of requests and errors.
4. Open the **Requests** sub tab and click your request to see its headers, the response and
   the duration next to the table.

See [Analyze your mocked HTTP API in Dashboard](/docs/http/dashboard.md) for filters and the
endpoint view.

## Return Your Own Data

Generated data is fine for a first look. For a test you often need a known value. Create a
file `petstore.js` that returns a fixed pet for ID 4:

```javascript tab=petstore.js
import { on } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    if (request.key === '/pet/{petId}' && request.path.petId === 4) {
      response.data = {
        id: 4,
        name: 'Garfield',
        category: { id: 3, name: 'Cats' },
        photoUrls: [],
        status: 'available'
      }
    }
  })
}
```

`request.key` is the path from the specification, so the script doesn't depend on the base path
`/api/v3`. Start Mokapi with the specification and the script:

```bash
mokapi https://raw.githubusercontent.com/marle3003/mokapi/main/examples/swagger/petstore-3.0.json petstore.js
```

Request the pet again:

```bash
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```

```json
{"id":4,"name":"Garfield","category":{"id":3,"name":"Cats"},"photoUrls":[],"status":"available"}
```

Every other pet ID still returns generated data. Mokapi watches the script, so you can change
it while Mokapi is running.

## Return an Error

The Petstore defines a `404` response for `getPetById`. To test how your application handles a
missing pet, set the status code in the same handler. Add it to `petstore.js` next to the
condition for pet 4, or replace the file content if you only want the error case:

```javascript tab=petstore.js
import { on } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    if (request.key === '/pet/{petId}' && request.path.petId === 999) {
      response.statusCode = 404
    }
  })
}
```

```bash
curl -i http://localhost/api/v3/pet/999
```

The response has status `404 Not Found`, and the request shows up in the request log of the
dashboard with that status code.

## Next Steps

- [Mocking HTTP APIs](/docs/http/overview.md): Delays, stateful mocks and Swagger 2.0 support.
- [Generate Realistic Test Data](/docs/get-started/test-data.md): Control the data Mokapi generates.
- [JavaScript API](/docs/javascript-api/overview.md): Everything you can do in a Mokapi Script.
