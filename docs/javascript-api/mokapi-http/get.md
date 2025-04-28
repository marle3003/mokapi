---
title: get( url, [args] ) - Mock HTTP GET Requests with Mokapi JavaScript API
description: Use Mokapi's JavaScript API to mock HTTP GET requests. Customize responses, handle request data, and simulate APIs for testing and development.
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