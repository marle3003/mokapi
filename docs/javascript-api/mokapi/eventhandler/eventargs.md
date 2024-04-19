---
title: EventArgs
description: EventArgs is an object used by on function.
---
# EventArgs

EventArgs is an object used by [on](/docs/javascript-api/mokapi/on.md) function.

| Name                    | Type    | Description                                                  |
|-------------------------|---------|--------------------------------------------------------------|
| tags                    | object  | Adds or overrides existing tags used in dashboard            |

## Examples

Add additional tag

```javascript
import { every } from 'mokapi'

export default function() {
    on('1m', function(request, response) {
    }, { tags: { foo: 'bar' } })
}
```