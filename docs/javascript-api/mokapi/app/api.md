---
title: app.api( title )
description: Scope HTTP, Kafka, MQTT and WebSocket handlers to a single API by its info.title.
---
# app.api( title )

Returns an `ApiScope` bound to a single API. All routers obtained from this scope only run
handlers for that API.

| Parameter | Type   | Description                                                                 |
|-----------|--------|-----------------------------------------------------------------------------|
| title     | string | API title as defined in `info.title` of the OpenAPI or AsyncAPI specification |

## Returns

An `ApiScope` object with the following methods:

| Method        | Returns                                                                | Description                                  |
|---------------|------------------------------------------------------------------------|----------------------------------------------|
| `http()`      | [HttpRouter](/docs/javascript-api/mokapi/app/http-router.md)           | HTTP router scoped to this API               |
| `kafka()`     | [KafkaRouter](/docs/javascript-api/mokapi/app/kafka-router.md)         | Kafka router scoped to this API              |
| `mqtt()`      | [MqttRouter](/docs/javascript-api/mokapi/app/mqtt-router.md)           | MQTT router scoped to this API               |
| `websocket()` | [WebsocketRouter](/docs/javascript-api/mokapi/app/websocket-router.md) | WebSocket router scoped to this API          |

## When to Use

Scoping is required when multiple APIs define the same path, topic or channel. Without a
scope, a handler runs for all of them.

```javascript
import { app } from 'mokapi'

export default function() {
    const petstore = app.api('Petstore')
    petstore.http().get('/pets', (request, response) => {
        response.data = [{ name: 'Bello' }]
    })

    const zoo = app.api('Zoo API')
    zoo.http().get('/pets', (request, response) => {
        response.data = [{ name: 'Simba' }]
    })
}
```

## Alternative: Scoping an Existing Router

The HTTP, Kafka and WebSocket routers also provide an `api( title )` method. It is useful
when you already hold a router and want to narrow it down:

```javascript
app.http().api('Petstore').get('/pets', (request, response) => {
    response.data = []
})
```

Both forms are equivalent: `app.api('Petstore').http()` and `app.http().api('Petstore')`.
