---
title: encode( input )
description: Encode base64 string
---
# encode( input )

Encodes a string value to its equivalent string representation that is encoded with base-64.

| Parameter | Type                 | Description                                                     |
|-----------|----------------------|-----------------------------------------------------------------|
| input     | string / ArrayBuffer | The input string to base64 encode.                              |

## Returns

| Type    | Description                              |
|---------|------------------------------------------|
| string  | The base64 encoding of the input string. |

## Example

```javascript
import { base64 } from 'mokapi/encoding'

export default function() {
    console.log(`The base64 encoding of 'hello world' is: ${base.encode('hello world')}`)
}
```