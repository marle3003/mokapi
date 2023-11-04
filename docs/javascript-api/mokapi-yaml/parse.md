---
title: "Javascript API: parse"
description: Parses a YAML string, constructing the JavaScript value or object described by the string.
---
# parse( text )

Parses a YAML string, constructing the JavaScript value or object described by the string.

| Parameter | Type   | Description                                                          |
|-----------|--------|----------------------------------------------------------------------|
| text      | string | The string to parse as YAML.                                         |

## Returns

| Type | Description                                     |
|------|-------------------------------------------------|
| any  | The value corresponding to the given YAML text. |

## Example Reading YAML File

```javascript tab=orders.js
import { parse } from 'mokapi/yaml'

export default function() {
    const orders = parse(open('orders.yaml'))
    console.log(outout[0].orderId)
}
```

```yaml tab=orders.yaml
- orderId: 345
  customer: Bob
- orderId: 874
  customer: Alice
```