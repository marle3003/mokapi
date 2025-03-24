---
title: date ( [args] )
description: Returns a textual representation of the date.
---
# date ( [args] )

Returns a textual representation of a date, optionally formatted according to a specified layout or based on a given timestamp.

## Parameters

| Name            | Type   | Description                                              |
|-----------------|--------|----------------------------------------------------------|
| args (optional) | object | A `DateArgs containing optional formatting parameters.   |

## DateArgs

| Name      | Type   | Description                                                                 |
|-----------|--------|-----------------------------------------------------------------------------|
| layout    | string | Specifies the format of the date. Defaults to RFC3339.                      |
| timestamp | number | A Unix timestamp (milliseconds since epoch). Defaults to current UTC time.  |

## Returns

| Type     | Description                                      |
|----------|--------------------------------------------------|
| string   | A formatted textual representation of the date.  |

## Examples

### Get the current date in RFC3339 format (default)

```javascript
import { date } from 'mokapi'

export default function() {
    console.log(date()) // Example: "2025-02-08T14:30:00Z"
}
```

### Get the date in Unix format

```javascript
console.log(date({ layout: 'UnixDate' })) // Example: "Sat Feb  8 14:30:00 UTC 2025"
```

### Format a specific timestamp

```javascript
const timestamp = 1738991400000 // Example timestamp
console.log(date({ timestamp, layout: 'RFC1123' })) 
// Example: "Sat, 08 Feb 2025 14:30:00 UTC"
```

## Supported Layouts

| Layout      | Example                             | Description                               |
|-------------|-------------------------------------|-------------------------------------------|
| RFC3339     | 2023-01-02T15:04:05Z07:00           | Default ISO 8601 format with timezone.    |
| RFC3339Nano | 2023-01-02T15:04:05.999999999Z07:00 | Higher precision version of RFC3339.      |
| RFC1123     | Mon, 02 Jan 2023 15:04:05 MST       | Common format used in HTTP headers.       |
| RFC1123Z    | Mon, 02 Jan 2023 15:04:05 0700      | RFC1123 format with numeric timezone.     |
| RFC882      | 02 Jan 23 15:04 MST                 | Variant of RFC822 with explicit timezone. |
| UnixDate    | Mon Jan 2 15:04:05 MST 2023         | Common UNIX date format.                  |
| DateTime    | 2023-01-02 15:04:05                 | Simple format without timezone.           |
| DateOnly    | 2023-01-02                          | Displays only the date.                   |
| TimeOnly    | 15:04:05                            | Displays only the time.                   |

## Custom Layouts Using Golang Format

Because Mokapi is written in Go, it supports the same layout formatting as Go's time package. 
You can define custom date formats by using the reference time `Mon Jan 2 15:04:05 MST 2006`. For instance:

```javascript
console.log(date({layout: '02-Jan-2006 15:04:05'}))
```

This outputs a custom date format like: 08-Feb-2025 14:30:00

Use this layout to create any custom date format you need by arranging the time and date components as desired.