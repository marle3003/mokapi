---
title: options( url, [body], [args] ) - Mock HTTP OPTIONS Requests with Mokapi JavaScript API
description: Use Mokapi's JavaScript API to mock HTTP OPTIONS requests. Customize responses, handle request data, and simulate APIs for testing and development.
---
# options( url, [body], [args] )

Make a OPTIONS request

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
import { options } from 'mokapi/http'

export default function() {
    const res = options("https://foo.bar/foo")
    console.log(res.headers['Allow'])
}
```