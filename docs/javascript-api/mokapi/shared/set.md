---
title: shared.set( key, value )
description: Sets a value for the given key in Mokapi's shared storage.
---
# shared.set( key, value )

Sets a value for the given key. If the key already exists, its value is replaced. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

| Parameter | Type   | Description                       |
|-----------|--------|-----------------------------------|
| key       | string | The key to store the value under. |
| value     | any    | The value to store.               |

## Returns

`void`

## Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.method === 'PUT' && request.key === '/simulations/delay') {
            shared.set('delay', request.query.duration)
        }
    })
}
```

``` box=tip
If you're reading the current value before deciding what to write, consider
[shared.update( key, updater )](/docs/javascript-api/mokapi/shared/update) instead. It applies the
change atomically, so two scripts updating the same key at the same time can't interfere with
each other.
```

## See Also

- [shared.get( key )](/docs/javascript-api/mokapi/shared/get.md)
- [shared.update( key, updater )](/docs/javascript-api/mokapi/shared/update.md)