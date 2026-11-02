---
title: app
description: Register handlers for HTTP, Kafka, MQTT and WebSocket in a structured, route-oriented style using protocol routers.
---
# app

`app` is the central object for registering handlers in a structured, route-oriented style.
Instead of writing one handler that inspects every request, you register handlers directly
for a path, operation, topic or channel.

`app` is a singleton: every import returns the same instance.

```javascript
import { app } from 'mokapi'

export default function() {
    app.http().get('/pets', (request, response) => {
        response.data = [{ id: 1, name: 'Bello' }]
    })
}
```

## Protocol Routers

Each protocol has its own router. The router knows which events the protocol supports and
offers matching methods, for example `get()` for HTTP or `produce()` for Kafka.

| Method                                                               | Returns                                                                       | Description                                       |
|----------------------------------------------------------------------|-------------------------------------------------------------------------------|---------------------------------------------------|
| `app.http()`                                                         | [HttpRouter](/docs/javascript-api/mokapi/app/http-router.md)                  | Handle HTTP requests by path, method or operation |
| `app.kafka()`                                                        | [KafkaRouter](/docs/javascript-api/mokapi/app/kafka-router.md)                | Handle and produce Kafka messages                 |
| `app.mqtt()`                                                         | [MqttRouter](/docs/javascript-api/mokapi/app/mqtt-router.md)                  | Handle and publish MQTT messages                  |
| `app.websocket()`                                                    | [WebsocketRouter](/docs/javascript-api/mokapi/app/websocket-router.md)        | Handle WebSocket connections and send messages    |
| [`app.api( title )`](/docs/javascript-api/mokapi/app/api.md)         | [ApiScope](/docs/javascript-api/mokapi/app/api.md)                            | Scope all routers to a single API                 |

## Scoping to an API

Routers returned directly by `app` are not bound to a specific API. Their handlers run for
**every** API that defines the matching path, topic or channel. This is useful for
cross-cutting concerns such as logging or common headers.

Use [`app.api( title )`](/docs/javascript-api/mokapi/app/api.md) to scope handlers to one API.
The title must match `info.title` of the OpenAPI or AsyncAPI specification.

```javascript
import { app } from 'mokapi'

export default function() {
    // Runs for all APIs that define GET /pets
    app.http().get('/pets', (request, response) => {
        response.headers['X-Mock'] = 'true'
    })

    // Runs only for the API with info.title 'Petstore'
    app.api('Petstore').http().get('/pets', (request, response) => {
        response.data = [{ id: 1, name: 'Bello' }]
    })
}
```

``` box=tip
Scope your handlers with `app.api()` as soon as more than one API defines the same path,
topic or channel. Otherwise a handler may run for an API you did not intend.
```

## Handler Order

All handlers registered for the same route run in registration order. Use the `priority`
option to change the order explicitly. When several handlers write the same value, the last
one wins.

```javascript
app.http().get('/pets', handler, { priority: 10 })
```

All router methods that register a handler accept the same options as
[EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md): `priority`, `track` and `tags`.

## app vs. on()

Both styles register event handlers and can be combined in the same script.

```javascript tab=app
import { app } from 'mokapi'

export default function() {
    app.api('Petstore').http().get('/pets/{petId}', (request, response) => {
        response.data = { id: request.path.petId, name: 'Bello' }
    })
}
```

```javascript tab=on
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.api === 'Petstore' && request.key === '/pets/{petId}' && request.method === 'GET') {
            response.data = { id: request.path.petId, name: 'Bello' }
        }
    })
}
```

Use `app` when you handle specific paths, operations, topics or channels. It keeps scripts
short, avoids `if` chains and gives you route-scoped features such as middleware with
[`use()`](/docs/javascript-api/mokapi/app/http-router.md#use),
[`response.context`](/docs/javascript-api/mokapi/eventhandler/httpresponse.md) and
[`response.stopPropagation()`](/docs/javascript-api/mokapi/eventhandler/httpresponse.md).

Use [`on()`](/docs/javascript-api/mokapi/on.md) for protocols without a router, such as LDAP and
SMTP.
