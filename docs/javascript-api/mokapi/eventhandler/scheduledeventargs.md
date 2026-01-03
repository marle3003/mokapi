---
title: ScheduledEventArgs
description: ScheduledEventArgs is an object used by every and cron function.
---
# ScheduledEventArgs

ScheduledEventArgs is an object used by [every](/docs/javascript-api/mokapi/every.md) and
[cron](/docs/javascript-api/mokapi/cron.md) function.

| Name                  | Type     | Default | Description                                                                                                                     |
|-----------------------|----------|---------|---------------------------------------------------------------------------------------------------------------------------------|
| times                 | number   | -1      | How many times the job should execute (-1 for unlimited).                                                                       |
| skipImmediateFirstRun | boolean  |         | Toggles behavior of first execution. If true job does not start immediately but rather wait until the first scheduled interval. |
| tags                  | object   | {}      | Optional tags for identifying the job.                                                                                          |

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