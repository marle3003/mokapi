---
title: "Javascript API: date"
description: Returns a textual representation of the date.
---
# date ( [args] )

Returns a textual representation of the date.

| Parameter       | Type   | Description                                 |
|-----------------|--------|---------------------------------------------|
| args (optional) | object | DateArgs object contains optional arguments |

## DateArgs

| Name      | Type   | Description                                                  |
|-----------|--------|--------------------------------------------------------------|
| layout    | string | The format of the textual representation, default is RFC3339 |
| timestamp | number | The timestamp of the date, default is current UTC time       |

## Returns

| Type     | Description                         |
|----------|-------------------------------------|
| string   | A textual representation of a date  |

## Examples

```javascript
import { date } from 'mokapi'

export default function() {
    console.log(date())
    console.log(date({layout: 'UnixDate'}))
    console.log(date({timestamp: new Date().getTime()}))
}
```

## Supported Layouts

| Layout      | Example                             |
|-------------|-------------------------------------|
| RFC3339     | 2023-01-02T15:04:05Z07:00           |
| RFC882      | 02 Jan 23 15:04 MST                 |
| RFC822Z     | 02 Jan 23 15:04 0700                |
| RFC850      | Monday, 02-Jan-23 15:04:05 MST      |
| RFC1123     | Mon, 02 Jan 2023 15:04:05 MST       |
| RFC1123Z    | Mon, 02 Jan 2023 15:04:05 0700      |
| RFC3339Nano | 2023-01-02T15:04:05.999999999Z07:00 |
| UnixDate    | Mon Jan 2 15:04:05 MST 2023         |
| DateTime    | 2023-01-02 15:04:05                 |
| DateOnly    | 2023-01-02                          |
| TimeOnly    | 15:04:05                            |
