---
title: HTTP DELETE Request – del(url[, body], args)
description: Use Mokapi's JavaScript API to mock HTTP DELETE requests. Customize responses, handle request data, and simulate APIs for testing and development.
---
# del( url, [body], [args] )

Make a DELETE request

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
import { del } from 'mokapi/http'

export default function() {
    del("https://foo.bar/foo")
}
```