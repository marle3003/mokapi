---
title: ScheduledEventArgs
description: ScheduledEventArgs is an object used by every and cron function.
---
# ScheduledEventArgs

ScheduledEventArgs is an object used by [every](/docs/javascript-api/mokapi/every.md) and
[cron](/docs/javascript-api/mokapi/cron.md) function.

| Name                  | Type    | Description                                                                                                                     |
|-----------------------|---------|---------------------------------------------------------------------------------------------------------------------------------|
| times                 | number  | Defines the number of times the scheduled function is executed.                                                                 |
| skipImmediateFirstRun | boolean | Toggles behavior of first execution. If true job does not start immediately but rather wait until the first scheduled interval. |
| tags                  | object  | Adds or overrides existing tags used in dashboard                                                                               |

## Examples

Write once to console after one minute

```javascript
import { every } from 'mokapi'

export default function() {
    every('1m', function() {
        console.log('foo')
    }, { times: 1, skipImmediateFirstRun: true })
}
```