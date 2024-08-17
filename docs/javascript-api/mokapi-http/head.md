---
title: head( url, [args] )
description: Make an HTTP HEAD request
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