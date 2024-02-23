---
title: ScheduledEventArgs
description: ScheduledEventArgs is an object used by every function.
---
# ScheduledEventArgs

ScheduledEventArgs is an object used by [every](/docs/javascript-api/mokapi/every.md) and
[cron](/docs/javascript-api/mokapi/every.md) function.

| Name                     | Type    | Description                                                              |
|--------------------------|---------|--------------------------------------------------------------------------|
| times                    | number  | Defines the number of times the scheduled function is executed.          |
| runFirstTimeImmediately  | boolean | Toggles behavior of first execution. Default is true                     |

## Examples

Write once to console after one minute

```javascript
import { every } from 'mokapi'

export default function() {
    every('1m', function() {
        console.log('foo')
    }, {times: 1, runFirstTimeImmediately: false})
}
```