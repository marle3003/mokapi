---
title: Response Object (mokapi/http)
description: Returned by HTTP request methods in module mokapi/http.
---
# Response

Response object contains response data from an HTTP request 
from methods in the module mokapi/http.

| Name       | Type     | Description                                                                                                                                   |
|------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| body       | string   | Response body content                                                                                                                         |
| headers    | object   | All response headers sent by the server in canonical form (for example 'Content-Type'). Accessing a header value returns an array of strings. |
| statusCode | number   | HTTP status code returned by the server                                                                                                       |
| json()     | function | Parses the response body as JSON. Returns a JS object, array or value                                                                         |

## Example

A Mokapi script that will make an HTTP request and print all response headers and the body as JSON.

```javascript
import { get } from 'mokapi/http'

export default function() {
    const res = get('https://foo.bar')
    for (const name in res.headers) {
        console.log(`${header}: ${res.headers[name]}`)
    }
    console.log(`content type: ${res.headers['Content-Type'][0]}`)
    console.log(`body: ${res.json()}`)
}
```