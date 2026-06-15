---
title: shared.has( key )
description: Checks whether a key exists in Mokapi's shared storage.
---
# shared.has( key )

Checks whether the given key exists in shared storage. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

| Parameter | Type   | Description             |
|-----------|--------|-------------------------|
| key       | string | The key to check.       |

## Returns

| Type    | Description                                       |
|---------|---------------------------------------------------|
| boolean | `true` if the key exists, otherwise `false`.      |

## Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (!shared.has('sessionToken')) {
            response.statusCode = 401
            response.data = { message: 'No active session' }
            return
        }

        response.data = { token: shared.get('sessionToken') }
    })
}
```

``` box=tip
`has` is useful when `undefined` could be a meaningful stored value and you need to distinguish
"key is missing" from "key exists but is set to `undefined`". For most cases, checking
`shared.get(key) ?? defaultValue` is simpler.
```

## See Also

- [shared.get( key )](/docs/javascript-api/mokapi/shared/get.md)
- [shared.delete( key )](/docs/javascript-api/mokapi/shared/delete.md)