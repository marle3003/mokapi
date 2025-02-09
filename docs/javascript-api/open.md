---
title: open( filePath, [args] )
description: Opens an existing file and reads its entire contents.
---
# open( filePath, [args] )

Reads the entire contents of an existing file.

``` box=tip
`open()` uses Mokapi's providers to access a file. If the file changes, the script automatically reloads and executes again.
```

## Parameters

| Name            | Type   | Description                                                        |
|-----------------|--------|--------------------------------------------------------------------|
| filePath        | string | The path to the file, absolute or relative, as file path or as URL |
| args (optional) | object | Args object containing additional arguments                        |

## Args

| Name       | Type   | Description                                                                                              |
|------------|--------|----------------------------------------------------------------------------------------------------------|
| as         | string | By default contents of the file are read as `string`, but with `binary` the file will be read as binary. |

## Returns

| Type                | Description               |
|---------------------|---------------------------|
| string &#124; array | The contents of the file. |

## Example Reading JSON File

``` box=tip url=[See Modules](../modules.md)
JSON files can also be imported directly.
```

```json tab=orders.json
[
  {
    "orderId": 1234512345,
    "customer": "Alice",
    "items": [
      {
        "itemId": 768, 
        "quantity": 3
      }
    ]
  }
]
```

```javascript
export default function() {
    const orders = JSON.parse(open('./orders.json'))
    console.log(orders[0].orderId)
}
```

## Example: Reading a Binary File

```javascript
export default function() {
    const image = open('./image.png', { as: 'binary' })
}
```

## Example: Reading from an HTTP Resource

```javascript
import { parse } from 'mokapi/yaml'

export default function() {
    const data = open('https://raw.githubusercontent.com/marle3003/mokapi/master/examples/openapi/users.yaml')
    const api = parse(data)
    console.log(api.info.title)
}
```