---
title: HttpRouter
description: Register HTTP handlers by path, method or operationId, and add middleware with use().
---
# HttpRouter

`HttpRouter` registers HTTP handlers by path, method or OpenAPI operation.
Get one via `app.http()` or `app.api('My API').http()`.

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Petstore').http()
        .get('/pets', (request, response) => {
            response.data = [{ id: 1, name: 'Bello' }]
        })
        .post('/pets', (request, response) => {
            response.statusCode = 201
        })
}
```

Handlers receive an [HttpRequest](/docs/javascript-api/mokapi/eventhandler/httprequest.md) and an
[HttpResponse](/docs/javascript-api/mokapi/eventhandler/httpresponse.md), the same as handlers
registered with [`on('http')`](/docs/javascript-api/mokapi/on.md).

``` box=info
Paths must match the **path template** of the OpenAPI specification, not the request URL.
Use `/pets/{petId}`, not `/v1/pets/123`.
```

## Methods

| Method                                          | Returns                                                     | Description                                                   |
|-------------------------------------------------|-------------------------------------------------------------|---------------------------------------------------------------|
| `get( path, handler, [opts] )`                  | `HttpRouter`                                                | Registers a handler for `GET` requests                        |
| `post( path, handler, [opts] )`                 | `HttpRouter`                                                | Registers a handler for `POST` requests                       |
| `put( path, handler, [opts] )`                  | `HttpRouter`                                                | Registers a handler for `PUT` requests                        |
| `patch( path, handler, [opts] )`                | `HttpRouter`                                                | Registers a handler for `PATCH` requests                      |
| `delete( path, handler, [opts] )`               | `HttpRouter`                                                | Registers a handler for `DELETE` requests                     |
| `head( path, handler, [opts] )`                 | `HttpRouter`                                                | Registers a handler for `HEAD` requests                       |
| `options( path, handler, [opts] )`              | `HttpRouter`                                                | Registers a handler for `OPTIONS` requests                    |
| `trace( path, handler, [opts] )`                | `HttpRouter`                                                | Registers a handler for `TRACE` requests                      |
| `query( path, handler, [opts] )`                | `HttpRouter`                                                | Registers a handler for `QUERY` requests                      |
| [`use( handler, [opts] )`](#use)                | `HttpRouter`                                                | Registers middleware that runs for all methods                |
| [`route( path )`](#route)                       | [HttpRoute](/docs/javascript-api/mokapi/app/http-route.md)  | Groups handlers for one path                                  |
| [`operation( operationId, handler, [opts] )`](#operation) | `HttpRouter`                                      | Registers a handler for an OpenAPI operation                  |
| [`api( title )`](#api)                          | `HttpRouter`                                                | Returns a router scoped to the given API                      |

### Parameters

| Parameter      | Type     | Description                                                                                           |
|----------------|----------|-------------------------------------------------------------------------------------------------------|
| path           | string   | OpenAPI path template, e.g. `/pets/{petId}`                                                           |
| handler        | function | An [HttpEventHandler](/docs/javascript-api/mokapi/eventhandler/httpeventhandler.md)                   |
| opts (optional)| object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) such as `priority`, `track`, `tags` |

All methods return the router, so calls can be chained.

## use

Registers a middleware handler that runs for **all** HTTP methods. Use it for shared logic
such as authentication checks, setting up request context or adding common headers.

Middleware runs in registration order relative to other handlers. Register it **before** the
method handlers, or use `priority` to control the order explicitly.

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Petstore').http()
        .use((request, response) => {
            if (!request.header['Authorization']) {
                response.statusCode = 401
                response.stopPropagation() // no further handlers run
            }
        })
        .get('/pets', (request, response) => {
            response.data = []
        })
}
```

See [`response.stopPropagation()`](/docs/javascript-api/mokapi/eventhandler/httpresponse.md) and
[`response.context`](/docs/javascript-api/mokapi/eventhandler/httpresponse.md) for passing data from
middleware to later handlers.

## route

Returns an [HttpRoute](/docs/javascript-api/mokapi/app/http-route.md) for a single path. Use it to
register several methods for the same path without repeating it:

```javascript
app.api('Petstore').http()
    .route('/pets/{petId}')
        .get((request, response) => {
            response.data = { id: request.path.petId }
        })
        .delete((request, response) => {
            response.statusCode = 204
        })
```

## operation

Registers a handler for the operation with the given `operationId`. Since an `operationId`
identifies both path and method, no further scoping is needed. This is especially handy for
long paths with many parameters.

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Petstore').http()
        .operation('listPets', (request, response) => {
            response.data = [{ id: 1, name: 'Bello' }]
        })
        .operation('createPet', (request, response) => {
            response.statusCode = 201
            response.data = { id: 2, name: request.body.name }
        })
}
```

## api

Returns a new `HttpRouter` scoped to the given API title. This is equivalent to
[`app.api( title ).http()`](/docs/javascript-api/mokapi/app/api.md).

```javascript
app.http().api('Petstore').get('/pets', (request, response) => {
    response.data = []
})
```

## Custom HTTP Methods

Handlers for other HTTP methods, such as `PURGE`, can be registered by calling the method
name directly:

```javascript
app.http().purge('/cache', (request, response) => {
    response.statusCode = 200
})
```

``` box=warning
Because any method name is accepted, TypeScript cannot detect typos such as `app.http().gte()`.
```
