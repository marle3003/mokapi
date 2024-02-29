---
title: "every( interval, handler, [args] )
description: Schedules a new periodic job with an interval
---
# every( interval, handler, [args] )

Schedules a new periodic job with an interval. Interval string 
is a possibly signed sequence of decimal numbers, each with 
optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".

``` box=info
By default, the first execution happens immediately, check ScheduledEventArgs
```

| Parameter       | Type     | Description                                                                                                         |
|-----------------|----------|---------------------------------------------------------------------------------------------------------------------|
| interval        | string   | Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".                                                     |
| handler         | function | The handler function to be executed every `interval`. By default, the first execution happens immediately.          |
| args (optional) | object   | [ScheduledEventArgs](/docs/javascript-api/mokapi/scheduledeventargs.md) object contains additional event arguments. | 

## Example

Write every minute to console.

```javascript
import { every } from 'mokapi'

export default function() {
    every('1m', function() {
        console.log('foo')
    })
}
```
