---
title: cron( expression, handler, [args] )
description: Schedules a new periodic job using cron expression.
---
# cron( expression, handler, [args] )

Schedules a new periodic job using cron expression. 

``` box=info
By default, the first execution happens immediately, check ScheduledEventArgs
```

## Parameters

| Name            | Type     | Description                                                                                                                         |
|-----------------|----------|-------------------------------------------------------------------------------------------------------------------------------------|
| expression      | string   | A cron expression defining the execution schedule.                                                                                  |
| handler         | function | A function executed at intervals specified by `expression`. The first execution happens immediately.                                |
| args (optional) | object   | [ScheduledEventArgs](/docs/javascript-api/mokapi/eventhandler/scheduledeventargs.md) object providing additional event arguments.   | 

## Understanding Cron Expressions

A cron expression consists of five fields, representing:

```plaintext
minute hour day month day-of-week
```

| Field        | Allowed Values    | Special Characters |
|--------------|-------------------|--------------------|
| Minute       | 0-59              | , - * /            |
| Hour         | 0-23              | , - * /            |
| Day of Month | 1-31              | , - * / ?          |
| Month        | 1-12 (or JAN-DEC) | , - * /            |
| Day of Week  | 0-6 (or SUN-SAT)  | , - * / ?          |

| Character | Meaning                                                                                                                                                                            |
|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| *         | Matches any value in the field (e.g., * in the minute field means "every minute").                                                                                                 |
| ,         | 	Specifies multiple values (e.g., 0,15,30,45 runs at those specific times).                                                                                                        |
| -         | 	Defines a range of values (e.g., 1-5 means Monday to Friday in the day-of-week field).                                                                                            |
| /         | 	Specifies step values (e.g., */5 in minutes means "every 5 minutes").                                                                                                             |
| ?         | 	Used in the day-of-month or day-of-week field to ignore the value when the other field is set (e.g., "0 12 ? * MON" runs every Monday at noon, ignoring the day-of-month field).  |

Examples:
- "* * * * *" → Runs every minute
- "0 12 * * *" → Runs daily at 12:00 PM
- "*/5 * * * *" → Runs every 5 minutes

## Examples

### Run a Job Every Minute

Write every minute to console.

```javascript
import { cron } from 'mokapi'

export default function() {
    cron('* * * * *', function() {
        console.log('Executing every minute')
    })
}
```

### Run a Job Every Hour at the 30-Minute Mark

```javascript
export default function() {
    cron('30 * * * *', function() {
        console.log('Running at minute 30 of every hour')
    })
}
```

### Run a Job Every Monday at 9 AM

```javascript
export default function() {
    cron('0 9 * * 1', function() {
        console.log(`Scheduled job executed`)
    })
}
```