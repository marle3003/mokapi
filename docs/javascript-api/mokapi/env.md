# env( name )

Gets the value of an environment variable.

| Parameter | Type     | Description                          |
|-----------|----------|--------------------------------------|
| name      | string   | The name of the environment variable |

## Returns

| Type     | Description                           |
|----------|---------------------------------------|
| string   | The value of the environment variable |

## Example

```javascript
import { env } from 'mokapi'

export default function() {
    console.log(env('foo'))
}
```