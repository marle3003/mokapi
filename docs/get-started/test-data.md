---
title: Generate Realistic Test Data
description: Mokapi generates random realistic test data or lets you customize responses with JavaScript to match your specific use case and scenarios.
subtitle: Mokapi automatically generates realistic test data from your API specifications. Customize responses with JavaScript or use declarative formats to match your exact testing needs.
cards:
  items:
    - title: Dashboard Guide
      href: docs/get-started/dashboard
      description: Learn how to validate and visualize your mock data
    - title: Write Scripts
      href: /docs/javascript-api/overview
      description: Add dynamic behavior to your mocks with JavaScript
    - title: Configure Mokapi
      href: /docs/configuration/overview
      description: Customize ports, providers, and other settings
    - title: Explore Tutorials
      href: /resources
      description: Follow step-by-step guides for REST, Kafka, LDAP, and SMTP
---

# Generate Realistic Test Data

Creating reliable test data is essential for accurate API testing and development.
Mokapi provides a powerful data generation engine that creates realistic test data
based on your API specifications.

You have multiple options for controlling generated data:

- **Automatic Generation**  
  Mokapi analyzes your schema and generates context-aware, realistic data automatically
- **Declarative Formats**  
  Use standard formats like `date`, `email`, `uuid` in your schema
- **JavaScript Customization**  
  Extend the generator with custom logic for specific fields or patterns
- **Full Response Control**
  Write complete response handlers with conditional logic and dynamic behavior

## Automatic Test Data Generation

By default, Mokapi analyzes your API's data structure and types to generate
meaningful responses. It also tailors responses based on request parameters,
ensuring relevant data is returned.

### Context-Aware Generation

For example, if a request filters for available pets, the response includes
only pets with `status: "available":

```bash tab=Single Pet
GET http://localhost/api/v3/pet/4
HTTP/1.1 200 OK
Content-Type: application/json
{
    "id": 4,
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
```bash tab=Filtered List
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

Notice how the status field in the second example matches the query parameter.
Mokapi understands the request context and generates appropriate data.

## Declarative Formats in Schema

Use standard OpenAPI formats to generate specific data types automatically. 
Mokapi recognizes these formats and produces realistic values:

| Format        | Type     | Example Output                                            |
|---------------|----------|-----------------------------------------------------------|
| date          | string   | 2017-07-21                                                |
| time          | string   | 17:32:28Z                                                 |
| date-time     | string   | 2017-07-21T17:32:28Z                                      |
| email         | string   | shanellewehner@cruickshank.biz                            |
| uuid          | string   | dd5742d1-82ad-4d42-8960-cb21bd02f3e7                      |
| uri           | string   | http://www.regionale-enable.info/virtual/portals/redefine |
| password      | string   | w*YoR94jFL@X                                              |
| ipv4          | string   | 68.244.174.93                                             |
| {zip} {city}  | string   | 82910 San Jose                                            |

### Example Schema

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

This declarative approach requires no JavaScript, just add the `format` field to your schema.

## Customize with JavaScript

For more control, extend Mokapi's data generator with JavaScript. This is useful
when you need values from a specific list, complex patterns, or custom logic.

### Example: Custom Enum Values

Generate random values from a predefined list:

```javascript
import { fake, findByName, ROOT_NAME } from 'mokapi/faker';

export default function() {
    const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly'];
    const root = findByName(ROOT_NAME);
    root.children.unshift({
        name: 'Frequency',
        attributes: [ 'frequency' ],
        fake: () => {
            return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
        }
    });
};
```

Now any field named `frequency` will randomly return one of these values.

``` box=tip title="Best Practice: Use Enums in Spec"
It's recommended to define possible values using `enum` in your OpenAPI 
specification. This allows Mokapi to randomly select from the list automatically
and clearly documents all available options for API users. The JavaScript
approach shown here demonstrates Mokapi's flexibility for cases where
enum isn't suitable.
```

## Full Response Control with Scripts

For complete control over API behavior, use Mokapi Scripts to customize 
responses, simulate errors, add delays, or implement complex logic.

### Example: Dynamic Time API

Create an API that returns the current timestamp:

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

This script intercepts requests to the `time` operation and returns the current
timestamp instead of a static mock value.

## Validate in the Dashboard

The Mokapi dashboard lets you:

- **Generate sample data:** See what Mokapi produces for your schemas
- **Validate against spec:** Ensure mock data matches your OpenAPI definition
- **Inspect requests:** View generated responses for different request parameters
- **Test edge cases:** Verify your custom generators work as expected

Access the dashboard at `http://localhost:8080` when Mokapi is running.

## What's Next?

Explore related topics to get more out of Mokapi's test data generation:

{{ card-grid key="cards" }}
