# Mokapi Javascript

You can provide custom behavior for your API, such as 
returning an 404 HTTP respond to a given HTTP request, 
transmitting data from a resource or producing data for a 
Kafka topic.

## Making a Time API

A simple Time API looks like this:

```yaml tab=OpenAPI
openapi: 3.0.0
info:
  title: Time API
  version: 0.1.0
servers:
  - url: /time
paths:
  /:
    get:
      summary: Get the current date and time,
      operationId: time
      responses:
        '200': 
          description: The current date and time
          content:
            application/json:
              schema: 
                type: string
                format: date-time
```

```javascript tab=Javascipt
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