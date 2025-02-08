---
title: Mocking APIs with Realistic Test Data
description: Mokapi generates random realistic test data or lets you customize responses with JavaScript to match your specific use case and scenarios.
---

# Mocking APIs with Realistic Test Data

Creating reliable test data is essential for accurate API testing and development.
Mokapi provides a powerful data generation engine that creates realistic test data
based on API specifications. You can customize and control the generated data using
JavaScript scripts or declarative configurations to adapt it to real-world scenarios.

## Automatic Test Data Generation

By default, Mokapi analyzes your API’s data structure and types to generate meaningful responses.
Additionally, Mokapi tailors responses based on request parameters, ensuring relevant data is returned.
For example, if a request filters for available pets, the response will include only pets with the status *available*.

```bash tab=/pet/4
GET http://localhost/api/v3/pet/4
HTTP/1.1 200 OK
Content-Type: application/json
{
    "id": 10,
    "name": "Larry",
    "category": {
        "id": 1,
        "name":"dog"
    },
    "photoUrls": [ 
        "https://www.dynamicsyndicate.biz/reintermediate/models"
    ],
    "tags": [
        {
            "id": 46110,
            "name": "Lux"
        }
    ],
    "status": "pending"
}
```
```bash tab=/pet/findByStatus
GET http://localhost/api/v3/pet/findByStatus?status=available
HTTP/1.1 200 OK
Content-Type: application/json
[
    {
        "name": "Yoda",
        "photoUrls": [
            "http://www.internalbandwidth.com/b2c/magnetic/dot-com/collaborative"
        ],
        "status": "available"
    }
]
```

## Extending Test Data with JavaScript

Mokapi’s test data engine is highly extensible using JavaScript, allowing you to customize responses 
based on specific attributes. In the example below, a random value from a predefined list is assigned to the frequency attribute.

``` box=info
It's recommended to define possible values as an enum within the API specification. This example illustrates the flexibility Mokapi offers.
```

```javascript
export default function() {
    if (!findByName('Frequency')) {
        const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
        const node = findByName('Strings')
        node.append({
            name: 'Frequency',
            test: (r) => { return r.lastName() === 'frequency' },
            fake: (r) => {
                return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
            }
        })
    }
}
```

## Scripting API Behavior

Mokapi Script allows you to customize API responses and simulate various behaviors, such as returning specific HTTP 
statuses or producing Kafka messages.

Example: A simple time API returning the current timestamp:
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

## Declarative Test Data Definition

If you prefer a declarative approach, Mokapi uses formats like dates, emails, or UUIDs directly in your schema:

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

## Test and Validate in the Dashboard
In the dashboard, you can easily generate sample data and validate it against your API’s specification. 
This allows you to ensure your mock data matches the expected structure and behavior. The following chapter will 
guide you through the process in more detail.

## Next Steps

- [Using Mock API Dashboard](dashboard.md)
- [Explore Mokapi Scripts](../../javascript-api/overview.md)
- [Learn More About Declarative Test Data](../../references/declarative-data.md)
