---
title: shared.update( key, updater )
description: Atomically updates a value in Mokapi's shared storage using an updater function.
---
# shared.update( key, updater )

Atomically updates a value using an updater function. The current value is passed into the
updater function, and the value it returns is stored and also returned by `update`. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

| Parameter | Type                          | Description                                                                                     |
|-----------|-------------------------------|-------------------------------------------------------------------------------------------------|
| key       | string                        | The key to update.                                                                              |
| updater   | (value: T \| undefined) => T  | Function that receives the current value (or `undefined` if not set) and returns the new value. |

## Returns

| Type | Description                     |
|------|---------------------------------|
| T    | The new value after the update. |

## Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        const count = shared.update('requests', c => (c ?? 0) + 1)
        response.headers['X-Request-Count'] = count.toString()
    })
}
```

``` box=tip
Use `update` instead of `get` followed by `set` whenever the new value depends on the current
one. Reading and writing in two separate steps creates a window where another script could
update the value in between, leading to lost updates. `update` performs both steps atomically.
```

## See Also

- [shared.get( key )](/docs/javascript-api/mokapi/shared/get.md)
- [shared.set( key, value )](/docs/javascript-api/mokapi/shared/set.md)