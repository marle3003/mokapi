---
title: "Javascript API: marshal"
description: Returns marshalled string representation of value
---
# marshal( value, [encoding] )

Returns marshalled string representation of value (>= v0.9.7)

| Parameter           | Type   | Description           |
|---------------------|--------|-----------------------|
| value               | any    | The value to marshal  |
| encoding (optional) | object | Encoding information  |

## Error

Throws an exception if an error occurs during marshalling

## Example

```javascript
import { marshal } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        response.body = marshal({username: 'foo'}) // {"username":"foo"}
        response.body = marshal({username: 'foo'}, {
            schema: {
                type: 'object',
                properties: {
                    username: {
                        type: 'string'
                    }
                }
            }
        }) // {"username":"foo"}
        response.body = marshal({username: 'foo'}, {
            schema: {
                type: 'object',
                xml: {name: 'user'}
            },
            contentType: 'application/json'
        }) // <user><username>foo</username></user>
    })
}
```