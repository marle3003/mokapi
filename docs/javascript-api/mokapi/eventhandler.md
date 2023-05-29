---
title: Javascript API: EventHandler
description: EventHandler is a function that is executed when an event is triggered.
---
# EventHandler

EventHandler is a function that is executed when an event is triggered.
EventHandler has event-specific parameters like [HttpRequest](/docs/javascript-api/mokapi/httprequest.md)
that contains data about an HTTP request.

## Returns

| Type    | Description                                           |
|---------|-------------------------------------------------------|
| boolean | Whether Mokapi should log execution of event handler. |

## HttpEventHandler

| Parameter | Type   | Description                                                                                            |
|-----------|--------|--------------------------------------------------------------------------------------------------------|
| request   | object | [HttpRequest](/docs/javascript-api/mokapi/httprequest.md) object contains data of a HTTP request       |
| response  | object | [HttpResponse](/docs/javascript-api/mokapi/httpresponse.md) object contains data for the HTTP response |

```javascript
import {on} from 'mokapi'

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