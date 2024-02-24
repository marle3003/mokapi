---
title: cron( expression, handler, [args] )
description: Schedules a new periodic job using cron expression.
---
# cron( expression, handler, [args] )

Schedules a new periodic job using cron expression. 

``` box=info
By default, the first execution happens immediately, check ScheduledEventArgs
```

| Parameter       | Type     | Description                                                                                                         |
|-----------------|----------|---------------------------------------------------------------------------------------------------------------------|
| expression      | string   | Cron expression                                                                                                     |
| handler         | function | The handler function to be executed every `expression`. By default, the first execution happens immediately.        |
| args (optional) | object   | [ScheduledEventArgs](/docs/javascript-api/mokapi/scheduledeventargs.md) object contains additional event arguments. | 

## Example

Write every minute to console.

```javascript
import { every } from 'mokapi'

export default function() {
    cron('* * * * *', function() {
        console.log('foo')
    })
}
```
