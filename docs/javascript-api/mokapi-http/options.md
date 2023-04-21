# options( url, [body], [args] )

Make a OPTIONS request

| Parameter       | Type   | Description                                                                                      |
|-----------------|--------|--------------------------------------------------------------------------------------------------|
| url             | string | Request URL like https://foo.bar                                                                 |
| body            | any    | Request body, objects will be encoded as application/json                                        |
| args (optional) | object | [Args](/docs/javascript-api/mokapi-http/args.md) object containing additional request parameters |

## Returns

| Type     | Description                                                             |
|----------|-------------------------------------------------------------------------|
| Response | [HttpResponse](/docs/javascript-api/mokapi-http/httpresponse.md) object |

## Example

```javascript
import { options } from 'mokapi/http'

export default function() {
    const res = options("https://foo.bar/foo")
    console.log(res.headers['Allow'])
}
```