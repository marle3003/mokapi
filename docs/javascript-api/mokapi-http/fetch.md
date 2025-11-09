---
title: Fetch – fetch( url, [opts] )
description: Use Mokapi's JavaScript API to send HTTP requests. Customize methods, headers, timeouts, and request data for testing and development.
---
# fetch( url, [body], [opts] )

The `fetch()` function provides a Promise-based interface for making HTTP requests in Mokapi’s JavaScript environment.  
It works similarly to the browser’s Fetch API.

| Parameter       | Type   | Description                                                                                                      |
|-----------------|--------|------------------------------------------------------------------------------------------------------------------|
| url             | string | Request URL e.g. https://foo.bar                                                                                 |
| opts (optional) | object | [FetchOptions](/docs/javascript-api/mokapi-http/fetchoptions.md) object containing additional request parameters |

## Returns

| Type              | Description                                                                                      |
|-------------------|--------------------------------------------------------------------------------------------------|
| Promise<Response> | A promise that resolves to a [Response](/docs/javascript-api/mokapi-http/httpresponse.md) object |

## Example

```javascript
import { put } from 'mokapi/http'

export default async function() {
    const res = await fetch('https://foo.bar/foo', {
        method: 'PUT',
        body: { 'foo': 'bar' },
        headers: { 'Content-Type': 'application/json' }
    })
    console.log(res.json())
}
```