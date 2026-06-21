---
title: shared.delete( key )
description: Removes a key and its value from Mokapi's shared storage.
---
# shared.delete( key )

Removes the specified key and its value from shared storage. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

| Parameter | Type   | Description             |
|-----------|--------|-------------------------|
| key       | string | The key to remove.      |

## Returns

`void`

## Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.method === 'DELETE' && request.key === '/simulations/delay') {
            shared.delete('delay')
            response.statusCode = 204
        }
    })
}
```

``` box=tip
Deleting a key that doesn't exist is not an error, it's simply a no-op.
```

## See Also

- [shared.has( key )](/docs/javascript-api/mokapi/shared/has.md)
- [shared.clear()](/docs/javascript-api/mokapi/shared/clear.md)