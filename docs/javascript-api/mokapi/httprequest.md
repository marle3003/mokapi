---
title: Javascript API: HttpRequest
description: HttpRequest is an object used by HttpEventHandlers
---
# HttpRequest

HttpRequest is an object used by [HttpEventHandler](/docs/javascript-api/mokapi/eventhandler.md)
that contains request-specific data such as HTTP headers.

| Name        | Type   | Description                                                              |
|-------------|--------|--------------------------------------------------------------------------|
| method      | string | Request method like `GET`                                                |
| url         | object | Url represents a parsed URL                                              |
| key         | string | Path value specified by the OpenAPI path                                 |
| operationId | string | OperationId defined in OpenAPI                                           |
| path        | map    | Object contains path parameters specified by OpenAPI path parameters     |
| query       | map    | Object contains query parameters specified by OpenAPI query parameters   |
| header      | map    | Object contains header parameters specified by OpenAPI header parameters |
| cookie      | map    | Object contains cookie parameters specified by OpenAPI cookie parameters |
| body        | any    | Any contains request body specified by OpenAPI request body              |

