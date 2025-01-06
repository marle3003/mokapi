---
title: decode( input )
description: Decode base64 string
---
# decode( input )

Decodes a string value that is encoded with base-64 to its equivalent unencoded string representation.

| Parameter | Type   | Description                        |
|-----------|--------|------------------------------------|
| input     | string | The base64 input string to decode. |

## Returns

| Type    | Description                                     |
|---------|-------------------------------------------------|
| string  | The base64 decoded version of the input string. |

## Example

```javascript
import { base64 } from 'mokapi/encoding'

export default function() {
    console.log(`The base64 decoded string of 'aGVsbG8gd29ybGQ=' is: ${base.decode('aGVsbG8gd29ybGQ=')}`)
}
```