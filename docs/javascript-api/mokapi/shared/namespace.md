---
title: shared.namespace( name )
description: Creates or returns a namespaced shared memory store to avoid key collisions between scripts.
---
# shared.namespace( name )

Creates or returns a namespaced shared memory store. Namespaces help avoid key collisions between
unrelated scripts that happen to use the same key names. See
[Shared State](/docs/javascript-api/mokapi/shared/overview.md) for an overview of how shared
storage works across scripts.

| Parameter | Type   | Description                   |
|-----------|--------|-------------------------------|
| name      | string | The namespace identifier.     |

## Returns

| Type          | Description                                                                                                                                               |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| SharedMemory  | A shared memory object scoped to the given namespace. Supports the same `get`, `set`, `update`, `has`, `delete`, `clear`, and `keys` methods as `shared`. |

## Example

```javascript
import { shared } from 'mokapi'

const petstore = shared.namespace('petstore')

export default function() {
    petstore.set('sessions', [])

    // Keys are scoped: this does not affect or read shared.get('sessions')
    // or any other script's "sessions" key.
}
```

``` box=tip
If your script defines several related keys, for example `sessions`, `users`, and `config`,
consider grouping them under one namespace named after your script or feature. This keeps your
keys readable and avoids accidental collisions with other scripts that might use the same
generic names.
```

## See Also

- [Shared State overview](/docs/javascript-api/mokapi/shared/overview.md)
- [shared.clear()](/docs/javascript-api/mokapi/shared/clear.md)