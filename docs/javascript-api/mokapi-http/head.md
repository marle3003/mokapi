---
title: head( url, [args] ) - Mock HTTP HEAD Requests with Mokapi JavaScript API
description: Use Mokapi's JavaScript API to mock HTTP HEAD requests. Customize responses, handle request data, and simulate APIs for testing and development.
---
# head( url, [args] )

Make a HEAD request

| Parameter       | Type   | Description                                                                                      |
|-----------------|--------|--------------------------------------------------------------------------------------------------|
| url             | string | Request URL like https://foo.bar                                                                 |
| args (optional) | object | [Args](/docs/javascript-api/mokapi-http/args.md) object containing additional request parameters |

## Returns

| Type     | Description                                                         |
|----------|---------------------------------------------------------------------|
| Response | [Response](/docs/javascript-api/mokapi-http/httpresponse.md) object |

## Example

```javascript
import { head } from 'mokapi/http'

export default function() {
    head("https://foo.bar/foo")
}
```