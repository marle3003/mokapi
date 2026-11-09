---
title: HttpRequest
description: HttpRequest is an object that provides access to request-specific data in an HTTP event handler.
---
# HttpRequest

`HttpRequest` is an object passed to an
[`HttpEventHandler`](/docs/javascript-api/mokapi/eventhandler/httpeventhandler.md).
It contains request-specific data extracted from the incoming HTTP request
and parsed according to the OpenAPI specification.

The available properties depend on the OpenAPI definition and the incoming request.

| Name        | Type   | Description                                                      |
|-------------|--------|------------------------------------------------------------------|
| method      | string | HTTP request method, such as `GET` or `POST`                     |
| url         | object | Parsed URL of the request, see [url](#url)                       |
| key         | string | Path value that matched the OpenAPI path template                |
| operationId | string | `operationId` defined in the OpenAPI specification               |
| path        | object | Path parameters defined by the OpenAPI path parameters           |
| query       | object | Query parameters defined by the OpenAPI query parameters         |
| header      | object | Header parameters defined by the OpenAPI header parameters       |
| cookie      | object | Cookie parameters defined by the OpenAPI cookie parameters       |
| body        | any    | Request body parsed according to the OpenAPI request body schema |
| api         | string | Name of the API, as defined in the OpenAPI `info.title` field    |

## url

`request.url` describes the URL the client called.

| Name   | Type   | Description                                                                                       |
|--------|--------|---------------------------------------------------------------------------------------------------|
| scheme | string | `https` if the request was received over TLS, otherwise `http`                                    |
| host   | string | Host name from the `Host` header of the request, without the port                                 |
| port   | number | Port from the `Host` header. If missing, the port of the request URL, else `443` (https) or `80`  |
| path   | string | Decoded request path, for example `/pets/my pet` for a request to `/pets/my%20pet`                |
| query  | string | Raw query string without the leading `?`, for example `limit=10&offset=0`                         |

`request.url.toString()` returns the scheme, host, path and query, for example
`http://localhost/pets?limit=10`. The port is not part of the string. Use
`request.query` to access parsed query parameters.

## Example

The following example demonstrates how to access request data inside an HTTP event handler.

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.method === 'GET' && request.operationId === 'getUser') {
            const userId = request.path.id
            const includeDetails = request.query.details

            response.body = {
                id: userId,
                details: includeDetails,
                requestedPath: request.url.path
            }
        }
    })
}
