---
title: Args
description: Args is an object used by functions in the module mokapi/http
---
# Args

Args is an object used by functions in the module mokapi/http 
and contains request-specific arguments like HTTP headers.

| Name     | Type    | Description           |
|----------|---------|-----------------------|
| headers  | object  | Key-value pair object |

## Example

```javascript
import { get } from 'mokapi/http'

export default function() {
    const res = get('https://foo.bar', {
        headers: {Accept: 'application/json'}
    })
    console.log(res.json())
}
```