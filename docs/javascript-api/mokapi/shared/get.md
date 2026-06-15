---
title: shared.get( key )
description: Returns the value associated with the given key from Mokapi's shared storage.
---
# shared.get( key )

Returns the value associated with the given key. See [Shared State](/docs/javascript-api/mokapi/shared/overview.md)
for an overview of how shared storage works across scripts.

| Parameter | Type   | Description           |
|-----------|--------|-----------------------|
| key       | string | The key to retrieve.  |

## Returns

| Type | Description                                     |
|------|-------------------------------------------------|
| any  | The stored value, or `undefined` if not found.  |

## Example

This handler counts how many times an endpoint has been called, using `shared.set` to update
the value and `shared.get` to read it back on the next request.

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        const count = (shared.get('counter') ?? 0) + 1
        shared.set('counter', count)

        response.headers['X-Request-Count'] = count.toString()
    })
}
```

## See Also

- [shared.set( key, value )](/docs/javascript-api/mokapi/shared/set.md)
- [shared.update( key, updater )](/docs/javascript-api/mokapi/shared/update.md)