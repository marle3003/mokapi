---
title: shared.keys()
description: Returns a list of all keys stored in Mokapi's shared storage.
---
# shared.keys()

Returns a list of all keys currently stored in shared storage. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

## Returns

| Type     | Description                    |
|----------|--------------------------------|
| string[] | An array of all stored keys.   |

## Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.key === '/simulations/state') {
            const state = {}
            for (const key of shared.keys()) {
                state[key] = shared.get(key)
            }
            response.data = state
        }
    })
}
```

``` box=tip
This is most useful for debugging or building a "show me everything that's currently set"
endpoint while developing a simulation. Note that `keys()` returns every key in shared storage,
including ones set by other scripts. If your script uses a [namespace](/docs/javascript-api/mokapi/shared/namespace),
call `keys()` on the namespaced object to see only your own keys.
```

## See Also

- [shared.get( key )](/docs/javascript-api/mokapi/shared/get.md)
- [shared.namespace( name )](/docs/javascript-api/mokapi/shared/namespace.md)