---
title: HttpEventHandler
description: EventHandler is a function that is executed when an event is triggered.
---
# HttpEventHandler

HttpEventHandler is a function that is executed when an HTTP event is triggered.

| Parameter | Type   | Description                                                                                                         |
|-----------|--------|---------------------------------------------------------------------------------------------------------------------|
| request   | object | [HttpRequest](/docs/javascript-api/mokapi/eventhandler/httprequest.md) object contains data of a HTTP request       |
| response  | object | [HttpResponse](/docs/javascript-api/mokapi/eventhandler/httpresponse.md) object contains data for the HTTP response |

## Returns

| Type    | Description                                           |
|---------|-------------------------------------------------------|
| boolean | Whether Mokapi should log execution of event handler. |

## Example

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        if (request.operationId === 'time') {
            response.data = new Date().toISOString()
            return true
        }
        return false
    })
}
```