---
title: "Javascript API: fake"
description: Creates a fake based on the given schema.
---
# fake( schema )

Creates a fake based on the given schema.

| Parameter       | Type   | Description                                                                                                |
|-----------------|--------|------------------------------------------------------------------------------------------------------------|
| schema          | object | [OpenAPI Schema](https://swagger.io/docs/specification/data-models/)  object contains definition of a type |

## Returns

| Type | Description        |
|------|--------------------|
| any  | The generated fake |

## Examples

```javascript
import { fake } from 'mokapi/faker'

export default function() {
    console.log(fake({type: 'string'}))
    console.log(fake({type: 'number'}))
    console.log(fake({type: 'string', format: 'date-time'}))
    console.log(fake({type: 'string', pattern: '^\d{3}-\d{2}-\d{4}$'})) // 123-45-6789
}
```