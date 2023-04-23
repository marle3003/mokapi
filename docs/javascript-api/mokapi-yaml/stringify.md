# stringify( value )

Converts a JavaScript value to a YAML string.

| Parameter | Type | Description                            |
|-----------|------|----------------------------------------|
| value     | any  | The value to convert to a YAML string. |

## Returns

| Type   | Description                                |
|--------|--------------------------------------------|
| string | A YAML string representing the given value |

## Example

```javascript
import { on } from 'mokapi'
import { stringify } from 'mokapi/yaml'

export default function() {
    const orders = [
        {
            orderId: 345,
            customer: 'Bob'
        },
        {
            orderId: 874,
            customer: Alice
        }
    ]
    
    on('http', function(request, response) {
        response.body = stringify(orders)
    })
}
```