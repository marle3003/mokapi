# delete( url, [body], [args] )

Make a DELETE request

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
import http from 'mokapi/http'

export default function() {
    http.delete("https://foo.bar/foo")
}
```

```javascript
import { del } from 'mokapi/http'

export default function() {
    del("https://foo.bar/foo")
}
```