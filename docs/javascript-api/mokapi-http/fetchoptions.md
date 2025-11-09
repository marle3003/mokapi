---
title: FetchOptions | Mokapi HTTP Module Function Parameters
description: The FetchOptions object defines parameters for the fetch() function in Mokapi's HTTP module, including method, headers, body, timeout, and redirect settings.
---
# FetchOptions

The `FetchOptions` object defines additional parameters for the [`fetch()`](/docs/javascript-api/mokapi-http/fetch.md) function in the `mokapi/http` module.  
It allows you to customize request behavior such as HTTP method, headers, body, timeout, and redirect handling.

| Name         | Type           | Description                                                                                                  |
|--------------|----------------|--------------------------------------------------------------------------------------------------------------|
| method       | string         | The HTTP method used for the request (e.g. `"GET"`, `"POST"`, `"PUT"`). Defaults to `"GET"`.                 |                                                                 |
| body         | object         | The request body to send. Automatically serialized to JSON when `Content-Type` is set to `application/json`. |
| headers      | object         | Key-value pairs representing HTTP headers to include with the request.                                       |
| maxRedirects | number         | The number of redirects to follow. Default value is 5. A value of 0 (zero) prevents all redirection.         |
| timeout      | number, string | Maximum time to wait for the request to complete. Default timeout is 60 seconds ("60s")                      |

## Example of Accept header

```javascript
import { fetch } from 'mokapi/http'

export default async function() {
    const res = await fetch('https://foo.bar', {
        headers: {Accept: 'application/json'}
    })
    console.log(res.json())
}
```

## Send a PUT request with a JSON body

```javascript
import { get } from 'mokapi/http'

export default function() {
    const res = get('https://foo.bar', {
        method: 'PUT',
        body: { foo: 'bar' }
    })
    console.log(res.json())
}
```