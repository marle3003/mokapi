---
title: fake( schema )
description: Creates a fake based on the given schema.
---
# fake( schema )

Generates a random value (a fake) that is valid against the given schema. Mokapi uses the same data generator
to build responses for your mocked APIs, so fake is useful whenever a script needs realistic test data.

| Parameter       | Type   | Description                                                                                                                                                 |
|-----------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| schema          | object | [OpenAPI Schema](https://swagger.io/docs/specification/data-models/) or [JSON Schema](https://json-schema.org) object that describes the value to generate. |

## Returns

| Type | Description                                          |
|------|------------------------------------------------------|
| any  | A generated value that is valid against the schema.  |

## Examples

### Simple values

```javascript
import { fake } from 'mokapi/faker'

export default function() {
    console.log(fake({ type: 'string' }))
    console.log(fake({ type: 'number' }))
    console.log(fake({ type: 'string', format: 'date-time' }))
    console.log(fake({ type: 'string', pattern: '^\\d{3}-\\d{2}-\\d{4}$' })) // e.g. 123-45-6789
}
```

### Objects and arrays

Property names are used as hints, so a property called email receives an email address and firstname receives a first name.

```javascript
import { fake } from 'mokapi/faker'

export default function() {
    const user = fake({
        type: 'object',
        properties: {
            id: { type: 'integer', minimum: 1 },
            firstname: { type: 'string' },
            lastname: { type: 'string' },
            email: { type: 'string', format: 'email' },
            roles: {
                type: 'array',
                items: { type: 'string', enum: ['admin', 'editor', 'viewer'] },
                minItems: 1,
                maxItems: 2,
                uniqueItems: true
            }
        },
        required: ['id', 'firstname', 'lastname', 'email']
    })
    console.log(user)
}
```

## Customizing the generated data

The generator is built as a tree of nodes. You can change existing nodes or add your own to produce data that fits your business domain, for example addresses
from your own country or product names from your catalog. See [findByName( name )](./findByName.md).