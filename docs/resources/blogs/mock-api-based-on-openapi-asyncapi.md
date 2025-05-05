---
title: Mock REST and Kafka APIs Using OpenAPI & AsyncAPI
description: Mock REST and Kafka APIs with Mokapi using OpenAPI or AsyncAPI. Simulate real world scenario using JavaScript — all in a single tool.
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

### 3. Call the API

```shell
curl http://localhost/pets
curl "http://localhost/pets?category=dog"
```

Mokapi will generate random data for those requests. If you query with ?category=dog, Mokapi will return only dogs 
thanks to smart mock data generation.

### Adding Custom Logic

You can add a JavaScript file to dynamically control responses:

```javascript
import { on } from 'mokapi';
import { fake } from 'mokapi/faker';

export default () => {
    on('http', (request, response) => {
       if (request.query.category === 'cat') {
           response.data = [
               {
                   id: fake({ type: 'string', format: 'uuid' }),
                   name: 'Molly',
                   category: 'cat'
               }
           ];
           return true;
       }
       return false;
    });
}
```

Run Mokapi with the script:
```shell
mokapi pet-api.yaml pets.js
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
    host: localhost:8092
    protocol: kafka
defaultContentType: application/json
operations:
  receivePetArrived:
    action: receive
    channel:
      $ref: '#/channels/store.pet.arrived'
  sendPetArrived:
    action: send
    channel:
      $ref: '#/channels/store.pet.arrived'
channels:
  store.pet.arrived:
    description: Details about a newly arrived pet.
    messages:
      userSignedUp:
        $ref: '#/components/messages/petArrived'
components:
  messages:
    petArrived:
      payload:
        $ref: '#/components/schemas/Pet'
  schemas:
    Pet:
      id:
        type: string
      name:
        type: string
      category:
        type: string
```

### 2. Create a Kafka producer

```javascript
import { every } from 'mokapi';
import { produce } from 'mokapi/kafka';

export default () => {
    every('10s', () => {
        produce({ topic: 'store.pet.arrived' });
    });
}
```

### 3. Run Mokapi with your AsyncAPI spec

```shell
mokapi pet-stream-api.yaml producer.js
```

Mokapi will simulate Kafka messages and allow you to inspect them via the dashboard.

## Integrating with Your Dev Workflow

- Use Mokapi in CI pipelines to test frontend against mocked backends
- Simulate Kafka events for microservice testing
- Share mock configurations with your team/organization

You can use Mokapi in a GitHub action run as docker container to streamline your automation.

```yaml
- name: Start Mokapi
  run: |
    docker run -d --rm --name mokapi -p 80:80 -p 8080:8080 -v ${{ github.workspace }}/mocks:/mocks mokapi/mokapi:latest /mocks
```

## Conclusion

Mokapi helps you mock REST and Kafka services quickly using standard API specs. Whether you’re building, testing, 
or demoing your application, Mokapi provides a simple yet powerful way to simulate realistic API behavior.

Give it a try: [https://mokapi.io](https://mokapi.io)