---
title: "Javascript API: open"
description: Opens an existing file and reads its entire contents.
---
# open( filePath, [args] )

Opens an existing file and reads its entire contents.

``` box=tip
`open()` uses Mokapi's providers to access a file. When the 
file is changed, the script is reloaded and executed again.
```

| Parameter       | Type   | Description                                                        |
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

``` box=tip
You can also import a json file directly, check Modules page.
```

```json file=orders.json
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

## Example Reading Binary File

```javascript
export default function() {
    const image = open('./image.png', { as: 'binary' })
}
```

## Example Reading From HTTP Resource

```javascript
import { parse } from 'mokapi/yaml'

export default function() {
    const data = open('https://raw.githubusercontent.com/marle3003/mokapi/master/examples/openapi/users.yaml')
    const api = parse(data)
    console.log(api.info.title)
}
```