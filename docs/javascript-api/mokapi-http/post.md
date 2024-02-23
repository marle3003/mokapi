---
title: post( url, [body], [args] )
description: Make an HTTP POST request
---
# post( url, [body], [args] )

Make a POST request

| Parameter       | Type   | Description                                                                                      |
|-----------------|--------|--------------------------------------------------------------------------------------------------|
| url             | string | Request URL like https://foo.bar                                                                 |
| body            | any    | Request body, objects will be encoded as application/json                                        |
| args (optional) | object | [Args](/docs/javascript-api/mokapi-http/args.md) object containing additional request parameters |

## Returns

| Type     | Description                                                     |
|----------|-----------------------------------------------------------------|
| Response | [Response](/docs/javascript-api/mokapi-http/response.md) object |

## Example

```javascript
import { post } from 'mokapi/http'

export default function() {
    const res = post('https://foo.bar/foo', { 'foo': 'bar' }, {
        headers: { 'Content-Type': 'application/json' }
    })
    console.log(res.json())
}
```