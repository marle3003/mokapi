---
title: HttpResponse
description: HttpResponse is an object that is used to construct the HTTP response in an HttpEventHandler.
---
# HttpResponse

`HttpResponse` is an object passed to an
[`HttpEventHandler`](/docs/javascript-api/mokapi/eventhandler/httpeventhandler.md).
It is used to define the outgoing HTTP response, including status code, headers,
and response body.

| Name       | Type   | Description                                                                  |
|------------|--------|------------------------------------------------------------------------------|
| statusCode | number | HTTP status code used to select the OpenAPI response definition              |
| headers    | object | Response headers defined by the OpenAPI response header parameters           |
| body       | string | Raw response body. Takes precedence over `data`                              |
| data       | any    | Response data that will be encoded according to the OpenAPI response schema  |

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

## Body vs Data

Use body to return a raw response body without OpenAPI encoding and validating.

Use data to return structured data that should be validated and encoded
according to the OpenAPI response definition

If both body and data are set, body takes precedence