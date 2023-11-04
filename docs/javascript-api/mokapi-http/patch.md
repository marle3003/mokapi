---
title: "Javascript API: patch"
description: Make an HTTP PATCH request
---
# patch( url, [body], [args] )

Make a PATCH request

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
import { patch } from 'mokapi/http'

export default function() {
    patch("https://foo.bar/foo")
}
```