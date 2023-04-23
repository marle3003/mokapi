# render( template, scope )

Renders the given mustache template with the given data

| Parameter | Type   | Description                                                          |
|-----------|--------|----------------------------------------------------------------------|
| template  | string | A [mustache](http://mustache.github.io) template.                    |
| scope     | object | A scope object that contains the data needed to render the template. |

## Returns

| Type    | Description            |
|---------|------------------------|
| string  | The rendered template  |

## Example Reading JSON File

```javascript
import { render } from 'mokapi/mustache'
import { fake } from 'mokapi/faker'

export default function() {
    const scope = {
        firstname: fake({
            type: 'string',
            format: '{firstname}'
        }),
        calc: () => ( 3 + 4 )
    }

    const output = render("{{firstname}} has {{calc}} apples", scope)
    console.log(outout)
}
```