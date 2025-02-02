---
title: Learn how to mock your API with realistic test data
description: Mokapi can randomly mock realistic test data or you can control it with JavaScript to create responses that fit your use case.
---

# Learn how to mock your API with realistic test data

By default, Mokapi generates test data randomly depending on 
the data types and structure with a tree algorithm. You can extend and 
modify this tree for your individual needs. If you need more dynamic 
responses, you can use Mokapi JavaScript to have more control over responses based on your specific
conditions or parameters. Mokapi Scripts allows you to create responses that can simulate a wide 
range of scenarios and edge cases.

``` box=tip
Additionally, the dashboard allows you to quickly generate realistic test data for any mocked API.
```

## Faker Tree

<img src="/mock-realistic-test-data.png" width="700" alt="Mokapi provides a powerful data generator" title="Visible in Mokapi's Dashboard" />
<br />

Mokapi analysis the data structure and data types using a tree to generate meaningful random data.
A possible response from the mocked Swaggers Petstore could be as follows

```json tab=Data
{
  "id": 26601,
  "name": "Hercules",
  "category": {
    "id": 83009,
    "name": "cheetah"
  },
  "photoUrls": [
    "https://www.chiefniches.com/monetize/systems/markets",
    "http://www.globalholistic.com/rich"
  ],
  "tags": [
    {
      "id": 39278,
      "name": "BerylBay"
    },
    {
      "id": 51507,
      "name": "OrionTrail"
    }
  ],
  "status": "available"
}
```

```yaml tab=Schema
{
  "type": "object",
  "properties": {
    "tags": {
      "type": "array",
      "items": {
        "ref": "#/components/schemas/Tag",
        "type": "object",
        "properties": {
          "id": {
            "type": "integer",
            "format": "int64"
          },
          "name": {
            "type": "string"
          }
        },
        "xml": {
          "name": "Tag"
        }
      },
      "xml": {
        "name": "tag",
        "wrapped": true
      }
    },
    "status": {
      "type": "string",
      "description": "pet status in the store",
      "enum": ["available", "pending", "sold"]
    },
    "id": {
      "type": "integer",
      "format": "int64"
    },
    "category": {
      "ref": "#/components/schemas/Category",
      "type": "object",
      "properties": {
        "id": {
          "type": "integer",
          "format": "int64"
        },
        "name": {
          "type": "string"
        }
      },
      "xml": {
        "name": "Category"
      }
    },
    "name": {
      "type": "string",
      "example": "doggie"
    },
    "photoUrls": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "xml": {
        "name": "photoUrl",
        "wrapped": true
      }
    }
  },
  "xml": {
    "name": "Pet"
  }
}
```

You can extend and modify this tree, and see it in the dashboard at runtime.

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

``` box=warning title="Warning: JavaScript file is not thread-safe"
A JavaScript file cannot have 
an event handler and a faker extension. You need to have these in different files.
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
[providers](/docs/configuration/dynamic/overview.md).

A simple Time API might look like the script below. To learn more about how to write Mokapi Script, see [JavaScript API](/docs/javascript-api/overview.md).

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
