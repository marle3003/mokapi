---
title: HttpEventHandler
description: HttpEventHandler is a function that is executed when an HTTP event is triggered.
---
# HttpEventHandler

An `HttpEventHandler` is a function that is executed whenever an HTTP event is triggered.
It allows inspecting the incoming request and modifying the outgoing response.

Multiple HTTP event handlers can be registered.  
If more than one handler is registered, they are executed in order based on their `priority`.

| Parameter | Type   | Description                                                                                                         |
|-----------|--------|---------------------------------------------------------------------------------------------------------------------|
| request   | object | [HttpRequest](/docs/javascript-api/mokapi/eventhandler/httprequest.md) object contains data of a HTTP request       |
| response  | object | [HttpResponse](/docs/javascript-api/mokapi/eventhandler/httpresponse.md) object used to construct the HTTP response |


## Example: Handling a specific operation

The following example registers a handler for HTTP events and returns the current date
when the request’s `operationId` is `time`.

```javascript
import { on, date } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        if (request.operationId === 'time') {
            response.body = date()
        }
    })
}
```

## Example: Controlling execution order with priority

Multiple HTTP event handlers can be registered for the same event.
The order in which they are executed is controlled by the priority option.

Handlers with a higher priority value are executed first.
If no priority is specified, the default priority is 0.

```javascript
import { on } from 'mokapi'

export default () => {
    let counter = 0

    // Executed last (priority: -1)
	on('http', (req, res) => {
		res.data.foo = 'handler1';
		res.data.handler1 = counter++ 
	}, { priority: -1 })
    
    // Executed first (priority: 10)
	on('http', (req, res) => {
		res.data.foo = 'handler2';
		res.data.handler2 = counter++ 
	}, { priority: 10 })

    // Executed second (priority: 0)
	on('http', (req, res) => {
		res.data.foo = 'handler3';
		res.data.handler3 = counter++ 
	})
}
```