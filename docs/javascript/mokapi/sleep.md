# sleep( time )

Suspends the execution for the specified duration.

| Parameter       | Type                 | Description                                                                                                                     |
|-----------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------|
| time            | number &#124; string | Duration in milliseconds or duration as string with unit. <br /> Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h` |

## Returns

| Type  | Description                                             |
|-------|---------------------------------------------------------|
| error | Returns an error if received wrong time duration format |

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