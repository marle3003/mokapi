---
title: put( url, [body], [args] )
description: Make an HTTP PUT request
---
# put( url, [body], [args] )

Make a PUT request

| Parameter       | Type   | Description                                                                                      |
|-----------------|--------|--------------------------------------------------------------------------------------------------|
| url             | string | Request URL like https://foo.bar                                                                 |
| body            | any    | Request body, objects will be encoded as application/json                                        |
| args (optional) | object | [Args](/docs/javascript-api/mokapi-http/args.md) object containing additional request parameters |

## Returns

| Type     | Description                                                         |
|----------|---------------------------------------------------------------------|
| Response | [Response](/docs/javascript-api/mokapi-http/httpresponse.md) object |

## Example

```javascript
import { put } from 'mokapi/http'

export default function() {
    const res = put('https://foo.bar/foo', { 'foo': 'bar' }, {
        headers: { 'Content-Type': 'application/json' }
    })
    console.log(res.json())
}
```