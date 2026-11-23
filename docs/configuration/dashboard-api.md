---
title: Dashboard API
description: The REST API behind the Mokapi dashboard, including the endpoint that validates an HTTP request and response against your OpenAPI contracts.
---
# Dashboard API

The dashboard reads its data from an API on the same port, under `/api`. Most endpoints exist
for the dashboard and may change between releases. These are useful on their own, for example in
a test or a script. To change the port or the path prefix, see
[Dashboard Configuration](/docs/configuration/dashboard.md).

The endpoints:

| Path                              | Description                                               |
|-----------------------------------|-----------------------------------------------------------|
| GET /api/info                     | Version and build time of Mokapi, the active protocols    |
| GET /api/services                 | All services with their protocol                          |
| GET /api/events                   | Recorded events, newest first, in pages (`page`, `size`)  |
| GET /api/events/{id}              | One event                                                 |
| GET /api/jobs                     | Scheduled jobs                                            |
| GET /api/search/query?q=...       | Full-text search, see [Search panel](/docs/get-started/dashboard.md#search-panel)       |
| POST /api/services/http/validate  | Validate an HTTP request and response, see below          |

## Validate a request and response

`POST /api/services/http/validate` checks an HTTP request and its response against the
OpenAPI contracts loaded in Mokapi. Use it to verify that the traffic of your application,
or the output of a test, matches the specification. The endpoint has no side effects: Mokapi
doesn't create an event or a metric and doesn't run any script.

```bash
curl -X POST http://localhost:8080/api/services/http/validate \
  -H "Content-Type: application/json" \
  -d '{
    "api": "Petstore",
    "request": {
      "method": "POST",
      "url": "/pet",
      "headers": { "Content-Type": "application/json" },
      "body": "{\"name\": \"Rex\"}"
    },
    "response": {
      "statusCode": 200,
      "headers": { "Content-Type": "application/json" },
      "body": "{\"id\": 1}"
    }
  }'
```

| Field                | Required | Description                                                             |
|----------------------|----------|-------------------------------------------------------------------------|
| api                  | no       | `info.title` of the OpenAPI specification. Without it, Mokapi searches all HTTP APIs. |
| partialResponse      | no       | `true` to validate the response body without `required` and `dependentRequired`. Available since Mokapi 0.60.0. See [Partial responses](#partial-responses). |
| request.method       | yes      | HTTP method, case-insensitive                                           |
| request.url          | yes      | Path and query string, for example `/pet/1?verbose=true`. An origin is ignored. |
| request.headers      | no       | Object with header names and string values                              |
| request.body         | no       | Request body as a string                                                |
| response.statusCode  | yes      | HTTP status code                                                        |
| response.headers     | no       | Object with header names and string values                              |
| response.body        | no       | Response body as a string                                               |

Header values and bodies are always strings. Mokapi converts header values to the type defined
in the specification, so the value `"5"` is valid for an integer header.

### Result

If the request is well-formed, Mokapi answers with status `200`, even if the validation fails.

```json
{
  "valid": false,
  "operation": {
    "api": "Petstore",
    "method": "POST",
    "path": "/pet"
  },
  "errors": [
    {
      "location": "request.body",
      "schema": "#/required",
      "message": "required properties are missing: email"
    }
  ]
}
```

| Field     | Description                                                                          |
|-----------|--------------------------------------------------------------------------------------|
| valid     | `true` if `errors` is empty                                                          |
| partialResponse | `true` if the request asked for partial responses. Not present otherwise. Mokapi sets it in every result, also if no operation matches or several APIs match. |
| operation | The matched operation: `api`, `method` in uppercase and `path` as written in the contract. Missing if no operation matches. |
| errors    | List of violations, an empty array if the request and response are valid             |

Each error has a `location`, a `message` and, for schema violations, a `schema` with the
JSON pointer to the failed schema keyword.

| Location                  | Checks                                              |
|---------------------------|-----------------------------------------------------|
| operation                 | No operation found, or several APIs match           |
| request.path.`<name>`     | Path parameter                                      |
| request.query.`<name>`    | Query parameter                                     |
| request.header.`<name>`   | Request header                                      |
| request.cookie.`<name>`   | Request cookie                                      |
| request.body              | Request body                                        |
| response.status           | No response defined for the status code             |
| response.header.`<name>`  | Response header                                     |
| response.body             | Response body                                       |

### Partial responses

Set `"partialResponse": true` if the response is only part of the real one, for example a mock
in a Storybook story that delivers only the properties it needs. Mokapi then validates the
response body without `required` and `dependentRequired`, on every nesting level, so a missing
property is no error. If several `oneOf` subschemas match, the first match wins and the body is
valid. A body that matches none of them is still an error.

Mokapi still checks everything the body contains: types, formats, enums, `additionalProperties`
and ranges. The request, including parameters and body, stays strict. Required response headers
and `response body is required` are still reported.

`oneOf` variants that differ only by their required properties can't be told apart by a partial
body, so any partial body matches them.

```bash
curl -X POST http://localhost:8080/api/services/http/validate \
  -H "Content-Type: application/json" \
  -d '{
    "api": "Petstore",
    "partialResponse": true,
    "request": { "method": "GET", "url": "/pet/1" },
    "response": {
      "statusCode": 200,
      "headers": { "Content-Type": "application/json" },
      "body": "{\"name\": \"Rex\"}"
    }
  }'
```

The result contains `"partialResponse": true`. A client can use this to detect a Mokapi
version that doesn't support the option. Mokapi versions before 0.60.0 ignore the field and
validate strictly.

### Finding the operation

Mokapi resolves the operation like it does for a mock request: a literal path wins over a
parameterized one, and the path of the `servers` URL is the base path of the API. The base
path is matched case-insensitively. A server URL with variables, such as
`https://{env}.example.com/v1`, is not supported and never matches.

If no operation matches, the result has `valid: false`, no `operation` and a single error with
the location `operation`. If operations in several APIs match, the error lists these APIs and
you set `api` to choose one.

### What is checked

Request parameters and the request body are validated, and Mokapi reports all errors instead
of stopping at the first. Extra headers are allowed. Security requirements aren't checked.

For the response, Mokapi checks that the status code is defined, the response headers and the
body. Header names are case-insensitive. A `Content-Type` header defined in the specification is
ignored, as OpenAPI requires.

The content type of the body must be defined for the status code, whatever the type is. Only
JSON bodies (`application/json` and `+json` types) are validated against the schema. Bodies
of other types, such as XML or plain text, are not checked.

A `HEAD` operation must be defined as `head:` in the specification. Mokapi doesn't validate its
response body.

If the response has no `Content-Type` header, Mokapi works as follows:

- An empty body is valid if the status code defines no content, otherwise the error is
  `response body is required`.
- Otherwise Mokapi tries each media type defined for the status code. If none fits, the error is
  `no matching content type defined`. If the status code defines exactly one JSON media type,
  Mokapi reports the schema errors of that type instead.

### Status codes

| Status | Description                                                                   |
|--------|-------------------------------------------------------------------------------|
| 200    | The request was validated. Check `valid` in the body.                         |
| 400    | Invalid JSON, a missing `request`, `response`, `request.method`, `request.url` or `response.statusCode`, or an unparseable `request.url` |
| 404    | The API in `api` doesn't exist                                                |
| 405    | The HTTP method is not `POST`                                                 |

Errors use the format `{ "message": "api 'Petstore' not found" }`.

### CORS

The API allows requests from any origin, so a browser application on another origin, for
example Storybook, can call the endpoint. An `OPTIONS` request to any `/api/...` path
answers with `204` and these headers:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST
Access-Control-Allow-Headers: Content-Type
```
