---
title: shared.clear()
description: Removes all stored entries from Mokapi's shared storage.
---
# shared.clear()

Removes all stored entries from shared storage. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

``` box=warning
This clears all shared state for the entire Mokapi process, not just the keys used by the
current script. If other scripts rely on shared state, calling `clear()` will affect them too.
If you only want to remove your own script's keys, use a [namespace](/docs/javascript-api/mokapi/shared/namespace)
and call `clear()` on the namespaced object instead.
```

## Returns

`void`

## Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.method === 'POST' && request.key === '/simulations/reset') {
            shared.clear()
            response.statusCode = 204
        }
    })
}
```

## See Also

- [shared.delete( key )](/docs/javascript-api/mokapi/shared/delete.md)
- [shared.namespace( name )](/docs/javascript-api/mokapi/shared/namespace.md)