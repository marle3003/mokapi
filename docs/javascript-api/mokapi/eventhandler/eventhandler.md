---
title: "EventHandler"
description: EventHandler is a function that is executed when an event is triggered.
---
# EventHandler

EventHandler is a function that is executed when an event is triggered. EventHandler has event-specific parameters like HttpRequest that contains data about an HTTP request.

## Returns

| Type    | Description                                           |
|---------|-------------------------------------------------------|
| boolean | Whether Mokapi should log execution of event handler. |

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