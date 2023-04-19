# on( event, handler, [args] )

Attaches an event handler for the given event.

| Parameter       | Type     | Description                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------|
| event           | string   | Event type such as `http                                                                              |
 | handler         | function | An [EventHandler](/docs/javascript-api/mokapi/eventhandler.md) to execute when the event is triggered |
 | args (optional) | object   | [EventArgs](/docs/javascript-api/mokapi/eventargs.md) object contains additional event arguments.     | 

## Example Echo Server

```javascript tab=echo.js
import { on } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        if (request.url.path === '/echo') {
            response.data = {
                url: request.url.toString(),
                body: request.body,
            }
        }
    })
}
```

```yaml tab=echo.yaml
openapi: 3.0.0
info:
  title: Echo Server
  version: '1.0'
servers:
 - url: /echo
paths:
  /:
    post:
      summary: Echoes the contents of the request body.
      requestBody: 
       content: 
        '*/*':
         schema:
          type: string
      responses:
        "200":
          description: the echo response
          content:
            application/json:
              schema:
                type: object
```