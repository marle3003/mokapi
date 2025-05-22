---
title: Mock REST and Kafka APIs Using OpenAPI & AsyncAPI
description: Mock REST and Kafka APIs with Mokapi using OpenAPI or AsyncAPI. Simulate real world scenario using JavaScript — all in a single tool.
dashboard:
  - title: HTTP Requests
    description: Incoming HTTP requests, including method, URL, headers, query parameters, and body
    img: /blogs/dashboard-petstore-request.png
  - title: Kafka Topics
    description: A live view of Kafka topic contents—see the messages being produced and consumed in real time.
    img: /blogs/dashboard-petstore-kafka-topic.png
  - title: Scheduled Jobs
    description: A list of all scheduled tasks defined in your scripts, along with their execution history, status, and any output or errors.
    img: /blogs/dashboard-petstore-jobs.png
---

# Mock REST and Kafka APIs with OpenAPI & AsyncAPI Using Mokapi

## Introduction

Mocking APIs speeds up development—especially when services are incomplete or unavailable. Whether you’re building a 
frontend, testing integrations, or developing in parallel, reliable mocks keep you productive.

Mokapi is an open-source tool designed to mock APIs like REST and Kafka using OpenAPI and AsyncAPI specifications.
With Mokapi, you can simulate API responses, customize behavior with scripts, and—most importantly—debug everything 
visually with a built-in dashboard.
The dashboard shows incoming HTTP requests, Kafka events, your JavaScript event handlers, and scheduled jobs—making it 
a powerful tool for debugging, understanding behavior, and refining your mocks.

In this article, you'll learn:
- What Mokapi is and why it's useful 
- How to mock a REST API using OpenAPI 
- How to mock a Kafka topic using AsyncAPI 
- How to extend mocks with custom behavior 
- How to integrate Mokapi into your dev workflow

## What is Mokapi?

Mokapi is a lightweight, developer-friendly tool for mocking REST and Kafka services. It supports OpenAPI and AsyncAPI 
specs out of the box and allows you to:
- Serve mock HTTP APIs
- Simulate Kafka topics and producers
- Write custom logic using JavaScript
- Visualize interactions in a web dashboard in real time\
  perfect for debugging and understanding mock behavior.

Mokapi is written in Go and available as a single binary, making it easy to run locally or in CI environments.

## Yet Another Mocking Tool?

Yes — but with a twist.

The focus of Mokapi is to ensure API specification compliance. For example, every incoming HTTP request is validated against the OpenAPI specification, and the response you define is also validated before it’s returned. This helps prevent contract mismatches early in development.

What makes Mokapi special compared to other mocking tools:
- Strict validation: Requests and responses must conform to your OpenAPI/AsyncAPI spec.
- Realistic data generation: You can use examples from your spec or generate meaningful mock data.
- Dynamic scripting: Inject behavior with JavaScript to simulate conditional logic, delays, or even stateful responses.
- Multiprotocol support: Supports both REST and Kafka-style messaging in one tool. 
- Live dashboard: See all requests, mock responses, and Kafka events in real time. 
- Developer-friendly: Single binary, easy CLI, scriptable and everything as code.

## Mocking a REST API with OpenAPI

### 1. Create an OpenAPI file

```yaml
openapi: 3.1.0
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: http://localhost
paths:
  /pets:
    get:
      summary: List all pets
      parameters:
        - in: query
          name: category
          schema:
            type: string
      responses:
        '200':
          description: A list of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                    name:
                      type: string
                    category:
                      type: string
```

### 2. Run Mokapi with your OpenAPI spec

```shell
mokapi pet-api.yaml
```

### 3. Make a request

```shell
curl http://localhost/pets
curl "http://localhost/pets?category=dog"
```

Mokapi will generate random data for those requests. If you query with ?category=dog, Mokapi will return only dogs 
thanks to smart mock data generation.

### Add Custom Logic with JavaScript

You can control responses using JavaScript:

```typescript
import { on } from 'mokapi';

const pets = [
    {
        id: 1,
        name: 'Yoda',
        category: 'dog'
    },
    {
        id: 2,
        name: 'Rex',
        category: 'dog'
    },
    {
        id: 3,
        name: 'Nugget',
        category: 'canary'
    }
];

export default () => {
    on('http', (request, response) => {
        if (request.query.category) {
          response.data = pets.filter(pet => pet.category === request.query.category)
        } else {
          response.data = pets
        }
        return true;
    });
}
```

Run it like this:
```shell
mokapi pet-api.yaml petstore.ts
```

## Mocking Kafka with AsyncAPI

### 1. Create an AsyncAPI file

```yaml
asyncapi: 3.0.0
info:
  title: Petstore Stream API
  version: 1.0.0
servers:
  kafkaServer:
    host: localhost:9092
    protocol: kafka
defaultContentType: application/json
operations:
  receivePetArrived:
    action: receive
    channel:
      $ref: '#/channels/petstore.order-event'
    messages:
      - $ref: '#/components/messages/OrderMessage'
  sendPetArrived:
    action: send
    channel:
      $ref: '#/channels/petstore.order-event'
    messages:
      - $ref: '#/components/messages/OrderMessage'
channels:
  petstore.order-event:
    description: Details about a newly placed pet store order.
    messages:
      OrderMessage:
        $ref: '#/components/messages/OrderMessage'
components:
  messages:
    OrderMessage:
      payload:
        $ref: '#/components/schemas/Order'
  schemas:
    Order:
      properties:
        id:
          type: integer
        petId:
          type: integer
        status:
          type: string
          enum: [ placed, accepted, completed ]
        shipDate:
          type: string
          format: date-time
        placed:
          type: object
          properties:
            timestamp:
              type: string
              format: date-time
          required:
            - timestamp
        accepted:
          type: object
          properties:
            timestamp:
              type: string
              format: date-time
          required:
            - timestamp
        completed:
          type: object
          properties:
            timestamp:
              type: string
              format: date-time
          required:
            - timestamp
      required:
        - id
        - petId
        - status
        - placed
```

### 2. Create a Kafka producer script

```typescript
import { every } from 'mokapi';
import { produce } from 'mokapi/kafka';

const orders = [
  {
    id: 1,
    petId: 1,
    status: 'placed',
    placed: {
      timestamp: new Date().toISOString()
    }
  },
  {
    id: 2,
    petId: 2,
    status: 'accepted',
    shipDate: new Date(Date.now() + 26*60*60*1000).toISOString(),
    placed: {
      timestamp: new Date(Date.now() - 60*60*1000).toISOString()
    },
    accepted: {
      timestamp: new Date().toISOString()
    }
  },
  {
    id: 3,
    petId: 3,
    status: 'completed',
    shipDate: new Date(Date.now() + 24*60*60*1000).toISOString(),
    placed: {
      timestamp: new Date(Date.now() - 2.5*60*60*1000).toISOString()
    },
    accepted: {
      timestamp: new Date(Date.now() - 1.5*60*60*1000).toISOString()
    },
    completed: {
      timestamp: new Date().toISOString()
    }
  }
]

export default () => {
    let index = 0;
    every('30s', () => {
        console.log('producing new random Kafka message')
        produce({
          topic: 'petstore.order-event',
          messages: [
            {
              key: orders[index].id,
              data: orders[index],
            }
          ]
        });
        index++;
    }, { times: orders.length });
}
```

### 3. Run Mokapi

```shell
mokapi pet-stream-api.yaml producer.ts
```

Mokapi will simulate Kafka messages and allow you to inspect them via the dashboard.

## Integrating with Your Dev Workflow

- Use Mokapi in CI pipelines to test frontend against mocked backends
- Simulate Kafka events for microservice testing
- Share mock configurations with your team/organization

You can use Mokapi in a GitHub action run as a docker container to streamline your automation.

```yaml
- name: Start Mokapi
  run: |
    docker run -d --rm --name mokapi \
    -p 80:80 -p 8080:8080 \
    -v ${{ github.workspace }}/mocks:/mocks \
    mokapi/mokapi:latest /mocks
```

## Dashboard

Mokapi’s built-in dashboard provides deep visibility into your mocked services—whether they’re REST APIs, Kafka topics, 
or scheduled jobs.

{{ carousel key="dashboard" }}


## Conclusion

Mokapi helps you mock REST and Kafka services quickly using standard API specs. Whether you’re building, testing, 
or demoing your application, Mokapi provides a simple yet powerful way to simulate realistic API behavior.

**Give it a try**: [https://mokapi.io](https://mokapi.io)\
**See the full example repo**: [https://github.com/marle3003/mokapi-petstore](https://github.com/marle3003/mokapi-petstore)

---

*This article is also available on [LinkedIn](https://www.linkedin.com/pulse/mock-rest-kafka-apis-openapi-asyncapi-using-mokapi-marcel-lehmann-snasf)*

---
