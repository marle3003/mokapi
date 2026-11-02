---
title: HttpRoute
description: Group HTTP method handlers and middleware for a single OpenAPI path.
---
# HttpRoute

`HttpRoute` groups handlers for a single path. Get one via
[`router.route( path )`](/docs/javascript-api/mokapi/app/http-router.md#route).

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Petstore').http()
        .route('/pets/{petId}')
            .use((request, response) => {
                if (!request.header['Authorization']) {
                    response.statusCode = 401
                    response.stopPropagation() // get and delete will not run
                    return
                }
                response.context.user = request.header['Authorization']
            })
            .get((request, response) => {
                response.data = { id: request.path.petId, owner: response.context.user }
            })
            .delete((request, response) => {
                response.statusCode = 204
            })
}
```

## Methods

Unlike [HttpRouter](/docs/javascript-api/mokapi/app/http-router.md), the methods of `HttpRoute` take
no path, because the route already defines it.

| Method                        | Description                                                   |
|-------------------------------|---------------------------------------------------------------|
| `get( handler, [opts] )`      | Registers a handler for `GET` requests on this route          |
| `post( handler, [opts] )`     | Registers a handler for `POST` requests on this route         |
| `put( handler, [opts] )`      | Registers a handler for `PUT` requests on this route          |
| `patch( handler, [opts] )`    | Registers a handler for `PATCH` requests on this route        |
| `delete( handler, [opts] )`   | Registers a handler for `DELETE` requests on this route       |
| `head( handler, [opts] )`     | Registers a handler for `HEAD` requests on this route         |
| `options( handler, [opts] )`  | Registers a handler for `OPTIONS` requests on this route      |
| `trace( handler, [opts] )`    | Registers a handler for `TRACE` requests on this route        |
| `query( handler, [opts] )`    | Registers a handler for `QUERY` requests on this route        |
| `use( handler, [opts] )`      | Registers middleware that runs for all methods on this route  |

All methods return the route, so calls can be chained.

### Parameters

| Parameter       | Type     | Description                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------|
| handler         | function | An [HttpEventHandler](/docs/javascript-api/mokapi/eventhandler/httpeventhandler.md)                   |
| opts (optional) | object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) such as `priority`, `track`, `tags` |

## Middleware on a Route

Middleware registered with `use()` runs for every method on this route, and only on this
route. This makes it a good fit for path-specific logic such as loading a resource once and
sharing it with all method handlers via
[`response.context`](/docs/javascript-api/mokapi/eventhandler/httpresponse.md).

```javascript
app.api('Petstore').http()
    .route('/pets/{petId}')
        .use((request, response) => {
            response.context.requestedAt = Date.now()
        })
        .get((request, response) => {
            response.data = { id: request.path.petId, requestedAt: response.context.requestedAt }
        })
```

Custom HTTP methods such as `PURGE` can be registered the same way as on
[HttpRouter](/docs/javascript-api/mokapi/app/http-router.md#custom-http-methods):
`route('/cache').purge(handler)`.
