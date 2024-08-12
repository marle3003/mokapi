---
title: get( url, [args] )
description: Make an HTTP GET request
---
# get( url, [args] )

Make a GET request

| Parameter         | Type   | Description                                                                                      |
|-------------------|--------|--------------------------------------------------------------------------------------------------|
| url               | string | Request URL like https://foo.bar                                                                 |
| args (optional)   | object | [Args](/docs/javascript-api/mokapi-http/args.md) object containing additional request parameters |

## Returns

| Type     | Description                                                         |
|----------|---------------------------------------------------------------------|
| Response | [Response](/docs/javascript-api/mokapi-http/httpresponse.md) object |

## Example

```javascript
import { get } from 'mokapi/http'

export default function() {
    const res = get('https://foo.bar')
    console.log(res.json())
}
```