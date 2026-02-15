---
title: HttpResponse
description: HttpResponse is an object that is used to construct the HTTP response in an HttpEventHandler.
---
# HttpResponse

`HttpResponse` is an object passed to an
[`HttpEventHandler`](/docs/javascript-api/mokapi/eventhandler/httpeventhandler.md).
It is used to define the outgoing HTTP response, including status code, headers,
and response body.

## Properties

| Name       | Type   | Description                                                                  |
|------------|--------|------------------------------------------------------------------------------|
| statusCode | number | HTTP status code used to select the OpenAPI response definition              |
| headers    | object | Response headers defined by the OpenAPI response header parameters           |
| body       | string | Raw response body. Takes precedence over `data`                              |
| data       | any    | Response data that will be encoded according to the OpenAPI response schema  |

## Methods

| Name                             | Description                  |
|----------------------------------|------------------------------|
| rebuild(statusCode, contentType) | Rebuilds the HTTP response   |

## Example

The following example demonstrates how to construct an HTTP response inside an
HTTP event handler.

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        response.statusCode = 200
        response.headers = {
            'Content-Type': 'application/json'
        }
        response.data = {
            message: 'Hello World'
        }
    })
}
```

## Default Response Generation

Mokapi automatically generates a valid HTTP response based on the OpenAPI
specification.

- The HTTP status code is automatically selected as the **first successful
  response (`200–299`) defined in the OpenAPI specification**, in the order they appear
- The response data is generated with valid example data according to the schema
  defined for the selected status code.
- Response headers defined in the OpenAPI specification are also generated with
  valid values

For example, if the OpenAPI specification defines a `201` response before a
`200` response, Mokapi selects `201` and generates the response body and headers
based on the schema defined for that status code.

This behavior ensures that every response is valid according to the OpenAPI
definition, even if the event handler does not explicitly modify the response.

## Body vs Data

Use body to return a raw response body without OpenAPI encoding and validating.

Use data to return structured data that should be validated and encoded
according to the OpenAPI response definition

If both body and data are set, body takes precedence

## Rebuilding a Response

When changing the HTTP status code or content type, the existing response data
and headers may no longer match the OpenAPI specification.

To address this, `HttpResponse` provides a helper function:

```typescript tab=Definition
rebuild(statusCode?: number, contentType?: string): void
```

This function rebuilds the entire HTTP response using the OpenAPI response
definition for the given status code and content type.

### What rebuild does

Calling rebuild will:
- Select the matching response definition from the OpenAPI specification 
- Set response.statusCode to the provided value 
- Generate valid response data based on the response schema 
- Generate valid response headers defined in the specification 
- Replace previously generated response data and headers

This ensures the response remains valid after changing the status code.

### When to Use rebuild

Use rebuild when:
- You change the response status code 
- You want to switch to a different response definition 
- You want Mokapi to regenerate valid example data and headers

You do not need to call rebuild if you only modify fields within the
already generated response.data.

### Example: Changing the Status Code Safely

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.path.petId === 10) {
            // Switch to a different OpenAPI response
            response.rebuild(404, 'application/json')

            // Modify only what matters
            response.data.message = 'Pet not found'
        }
    })
}
```

### Parameter Defaults

- If `statusCode` is not provided, Mokapi selects the OpenAPI `default` response.
- If `contentType` is not provided, Mokapi selects the first content type
  defined for the selected status code in the OpenAPI specification.

### Error handling

If `response.rebuild()` throws an error, and it is not caught, the current event
handler is skipped and no response modifications from that handler are applied.

```javascript
import { on } from 'mokapi'
import { read, writeString } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        let data: { request: HttpRequest, response: HTTPResponse}[] = []
        try {
            const s = read('./data.json');
            data = JSON.parse(s)
        } catch {}
        data.push({ request, response })
        writeString('./data.json', JSON.stringify(data))
    }, { prriority: -1 })
}
```