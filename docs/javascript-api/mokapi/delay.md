---
title: delay( time )
description: Delays a handler for the specified duration without blocking other handlers of the script.
---
# delay( time )

Returns a Promise that resolves after the specified duration. Use it in an `async` handler with `await` to
delay the response. Unlike [sleep](/docs/javascript-api/mokapi/sleep.md), `delay` only holds back the
handler that awaits it. Other handlers of the same script file keep answering requests.

| Parameter       | Type                 | Description                                                                                                                                                                      |
|-----------------|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| time            | number &#124; string | Duration in milliseconds or duration as string with unit. <br /> Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. <br /> Decimal numbers are truncated to whole milliseconds. `0` and negative values resolve right away. |

## Returns

| Type          | Description                                                      |
|---------------|------------------------------------------------------------------|
| Promise<void> | A Promise that resolves with `undefined` after the given duration |

## Error

Throws an exception if the time has the wrong type or the duration string has the wrong format, for example
`time: missing unit in duration "1"`. The error is thrown immediately, not through the Promise.

## Example

```javascript
import { on, delay } from 'mokapi'

export default function() {
    on('http', async function(request, response) {
        if (request.operationId === 'foo') {
            await delay(10) // delays 10 milliseconds
        }
        if (request.operationId === 'bar') {
            await delay('10s') // delays 10 seconds
        }
    })
}
```

The handler must be `async` and must `await` the call. Without `await`, the handler returns immediately and
nothing is delayed.

## Script reload

If you change the script file while a delay is pending, Mokapi reloads the script right away. The request
isn't aborted: the handler stops at the `await` and Mokapi answers immediately. Changes the handler made to the
response before the `await` are kept, the code after it doesn't run. Mokapi logs
`unable to execute event handler: event loop stopped`.
