---
title: HTTP PATCH Request – patch( url, [body], [args] )
description: Use Mokapi's JavaScript API to mock HTTP PATCH requests. Customize responses, handle request data, and simulate APIs for testing and development.
---
# patch( url, [body], [args] )

Make a PATCH request

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
import { patch } from 'mokapi/http'

export default function() {
    patch("https://foo.bar/foo")
}
```