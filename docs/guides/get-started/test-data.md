---
title: Test Data Generator
description: Use randomly generated test data or use Mokapi Script to create responses that can simulate a wide range of scenarios and edge cases.
---

# Test Data Generator

By default, Mokapi generates test data randomly depending on 
the data types. If you need more dynamic responses, you can 
use Mokapi Script to generate responses based on specific 
conditions or parameters. Mokapi Scripts allows you to 
create responses that can simulate a wide range of scenarios
and edge cases.

``` box=tip
You can also use Mokapi's endpoint to generate test data on the fly,
see Mokapi's Dashboard for details
```

## Declarative Test Data

Providing a detailed specification for data types can greatly
enhance the usefulness and accuracy of randomly generated
data by ensuring that the generated data aligns with real-word
scenarios and is consistent and error-free. 

A list of more options that can be used, refer to
the [reference page](/docs/references/declarative-data.md).

```yaml
schema:
  type: object
  properties:
    date:
      type: string
      format: date # 2017-07-21
    time:
      type: string
      format: date-time # 2017-07-21T17:32:28Z
    email:
      type: string
      format: email # demetrisdach@yost.org
    guid:
      type: string
      format: uuid # dd5742d1-82ad-4d42-8960-cb21bd02f3e7
```

## Mokapi Script
You can provide custom behavior for your API, such as
returning an 404 HTTP respond to a specific HTTP request,
transmitting data from a resource or producing data for a
Kafka topic. Mokapi reads your custom scripts from 
[providers](/docs/configuration/providers/overview.md).

A simple Time API might look like this:

```javascript tab=time.js
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

```yaml tab=api.yaml
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
To learn more about how to write Mokapi Script, see [JavaScript API](/docs/javascript-api/javascript.md).