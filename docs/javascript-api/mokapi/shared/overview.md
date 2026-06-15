---
title: Shared State (mokapi/shared)
description: Persist and share data between multiple Mokapi scripts using the shared memory API.
---
# Shared State

Each Mokapi script runs in its own JavaScript runtime. A variable defined in one file isn't visible
in another, even if both scripts are part of the same mock. `shared` solves this: it's an in-memory
store that all scripts in the same Mokapi process can read from and write to.

This makes it possible to:

- Keep counters or statistics across requests (e.g. how many times an endpoint was called)
- Cache data computed by one script and reuse it in another
- Coordinate simulation state across multiple handlers, for example a device's current status
  that both a "real API" script and a "simulation API" script need to read and update
- Simulate application-level state without relying on a real database

Values are stored in memory and persist for the lifetime of the Mokapi process. Restarting Mokapi
clears all shared state.

## Basic Example

```javascript
import { on, shared } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        const count = shared.update('requests', c => (c ?? 0) + 1)
        response.headers['X-Request-Count'] = count.toString()
    })
}
```

`shared.update` reads the current value, applies the function, stores the result, and returns it,
all in one atomic step. This is the safest way to modify shared state, since two scripts updating
the same key at the same time can't interfere with each other.

## Namespaces

Because `shared` is a single store for the entire Mokapi process, keys can collide between scripts
that weren't written to work together. Use `namespace()` to scope a set of keys to your script:

```javascript
import { shared } from 'mokapi'

const store = shared.namespace('petstore')
store.set('sessions', [])
```

`store` behaves exactly like `shared`, but all its keys live under the `petstore` namespace and
won't collide with keys used by other scripts.

## API Reference

| Function                                                             | Description                                               |
|----------------------------------------------------------------------|-----------------------------------------------------------|
| [get(key)](/docs/javascript-api/mokapi/shared/get.md)                | Returns the value for a key, or `undefined` if not found. |
| [set(key, value)](/docs/javascript-api/mokapi/shared/set.md)         | Sets the value for a key, replacing any existing value.   |
| [update(key, updater)](/docs/javascript-api/mokapi/shared/update.md) | Atomically updates a value using an updater function.     |
| [has(key)](/docs/javascript-api/mokapi/shared/has.md)                | Checks whether a key exists.                              |
| [delete(key)](/docs/javascript-api/mokapi/shared/delete.md)          | Removes a key and its value.                              |
| [clear()](/docs/javascript-api/mokapi/shared/clear.md)               | Removes all entries. Use with caution.                    |
| [keys()](/docs/javascript-api/mokapi/shared/keys.md)                 | Returns all stored keys.                                  |
| [namespace(name)](/docs/javascript-api/mokapi/shared/namespace.md)   | Returns a `shared`-like object scoped to a namespace.     |