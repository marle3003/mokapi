---
title: EventArgs
description: EventArgs is an object used to configure event handlers registered with the on function.
---
# EventArgs

`EventArgs` is an optional configuration object passed to the
[`on`](/docs/javascript-api/mokapi/on.md) function when registering an event handler.
It allows controlling how and when an event handler is executed.

| Name     | Type    | Description                                                                                            |
|----------|---------|--------------------------------------------------------------------------------------------------------|
| tags     | object  | Adds or overrides existing tags that are used in dashboard                                             |
| priority | integer | Defines the execution priority of the event handler. Handlers with a higher value are executed first.  |

If no priority is specified, the default priority is `0`.

## Example: Adding custom tags

The following example registers an event handler and adds a custom tag.

```javascript
import { every } from 'mokapi'

export default function() {
    on('1m', function(request, response) {
        // handler logic
    }, { tags: { foo: 'bar' } })
}
```

## Example: Controlling execution order with priority

When multiple handlers are registered for the same event, the priority
property controls the order in which they are executed.

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', (req, res) => {
        res.data.stage = 'early'
    }, { priority: 10 })

    on('http', (req, res) => {
        res.data.stage = 'default'
    })

    on('http', (req, res) => {
        res.data.stage = 'late'
    }, { priority: -10 })
}
```

In this example:
- The handler with priority 10 is executed first 
- The handler without a priority is executed next (priority 0)
- The handler with priority -10 is executed last