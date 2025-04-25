---
title: Args Object | Mokapi HTTP Module Function Parameters
description: The Args object is used by functions in the Mokapi HTTP module to access request parameters, headers, body data, and more for API mocking logic.
---
# Args

Args is an object used by functions in the module mokapi/http 
and contains request-specific arguments like HTTP headers.

| Name          | Type   | Description                                                                                          |
|---------------|--------|------------------------------------------------------------------------------------------------------|
| headers       | object | Key-value pair object                                                                                |
| maxRedirects  |        | The number of redirects to follow. Default value is 5. A value of 0 (zero) prevents all redirection. |

## Example of Accept header

```javascript
import { get } from 'mokapi/http'

export default function() {
    const res = get('https://foo.bar', {
        headers: {Accept: 'application/json'}
    })
    console.log(res.json())
}
```

## Example max redirects

```javascript
import { get } from 'mokapi/http'

export default function() {
    const res = get('https://foo.bar', {
        maxRedirects: 0
    })
    console.log(res.headers['Location'][0])
}
```