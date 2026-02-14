---
title: EventHandler
description: EventHandler is a function that is executed when an event is triggered.
---
# EventHandler

An `EventHandler` is a function that is executed whenever a registered event is triggered.
The parameters passed to the handler depend on the event type (for example, an HTTP event
provides an `HttpRequest` and `HttpResponse` object).

Multiple handlers can be registered for the same event.

## Usage

Event handlers are registered using the `on` function.
The first argument specifies the event type, the second argument is the handler function.

## Example: Handling an HTTP event

The following example registers an HTTP event handler that responds only to a specific
operation (operationId === 'time').

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