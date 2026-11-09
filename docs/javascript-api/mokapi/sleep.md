---
title: sleep( time )
description: Suspends the execution for the specified duration.
---
# sleep( time )

Suspends the execution for the specified duration.

| Parameter       | Type                 | Description                                                                                                                     |
|-----------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------|
| time            | number &#124; string | Duration in milliseconds or duration as string with unit. <br /> Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h` |

`sleep` blocks every handler of the script file while it waits. Requests to other handlers of the same file
are answered only after the sleep is over. To delay a single handler, use
[delay](/docs/javascript-api/mokapi/delay.md) instead.

## Error

Throws an exception if time duration format has wrong format

## Example

```javascript
import { on, sleep } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        if (request.operationId === 'foo') {
            sleep(10) // sleeps 10 milliseconds
            return true
        }
        if (request.operationId === 'bar') {
            sleep('10s') // sleeps 10 seconds
            return true
        }
        return false
    })
}
```