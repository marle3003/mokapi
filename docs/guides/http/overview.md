---
title: How to mock HTTP APIs with Mokapi
description: Mock any HTTP API with OpenAPI specification
---
# Mocking HTTP APIs

Mokapi makes it easy to mock HTTP APIs, enabling developers to test and debug their applications with minimal effort. Whether you need to validate request handling, simulate complex API responses, or troubleshoot edge cases, Mokapi provides a versatile and developer-friendly solution tailored for HTTP API testing.

Designed to integrate seamlessly with your projects, Mokapi lets you create mock APIs using the OpenAPI Specification. It generates dynamic HTTP responses based on your API definitions, eliminating the need for a live server during development. This flexibility empowers developers to experiment, prototype, and troubleshoot more effectively.

With Mokapi, you can go beyond basic mocks by writing custom scripts to control the behavior of your APIs. This allows you to simulate a wide range of scenarios, such as conditional responses or stateful interactions. Mokapi also supports fetching API definitions directly from URLs or files, making it simple to get started with existing OpenAPI documents.

By leveraging Mokapi, you can streamline your development workflow, reduce dependencies on external systems, and deliver robust HTTP API integrations with confidence.

Learn how to create your first HTTP API mock with Mokapi and begin ensuring the reliability and robustness of your application.

## Before you start

You can run Mokapi in multiple ways based on your needs. Learn how to configure and launch Mokapi on your local machine by following the instructions [here](/docs/guides/get-started/installation.md).


## Launch Mokapi with Swagger's PetStore API


To get started quickly, you can use Swagger's PetStore API specification hosted online:

```bash tab=CLI
mokapi --providers-http-url https://petstore3.swagger.io/api/v3/openapi.json
```

That’s all you need! Open your browser and navigate to http://localhost/api/v3/pet/12 to see Mokapi's generated API response based on the PetStore specification.

``` box=info
If you encounter issues fetching the specification file from swagger.io, you might need to configure a proxy server using this command:  
--providers-http-proxy http://proxy.server.com:port
```

## Customizing HTTP Responses with Mokapi

For more dynamic control, Mokapi allows you to define custom HTTP responses using Mokapi Scripts. This lets you simulate various scenarios, adjust responses based on request parameters, or handle specific conditions.

### Example: Custom Response for particular Pet ID

Create a petstore.ts script to define a custom response for /pet/12:

```typescript tab=petstore.ts (TypeScript)
import { on } from 'mokapi'

export default function() {
    on('http', (request, response) => {
        if (request.path.petId === 12) {
            response.data = {
                id: 12,
                name: 'Garfield',
                category: {
                    id: 3,
                    name: 'Cats'
                },
                photoUrls: []
            }
        }
    })
}
```

Start the mock server with the following command, referencing both the API specification and your custom script:

```bash tab=CLI
mokapi --providers-http-url https://petstore3.swagger.io/api/v3/openapi.json --providers-file-filename /path/to/petstore.ts
```

Now, when you visit [http://localhost/api/v3/pet/12](http://localhost/api/v3/pet/12), Mokapi will return your custom-defined response for Garfield. Requests for other pet IDs will still generate random data based on the API specification.

For further details on creating dynamic data, see [Test-Data](/docs/guides/get-started/test-data.md).

## Swagger 2.0 support

Mokapi supports the Swagger 2.0 specification, making it easy to work with older API definitions. When you provide a Swagger 2.0 file, Mokapi automatically converts it to OpenAPI 3.0, ensuring compatibility and consistent response generation. This seamless conversion allows you to benefit from OpenAPI 3.0 features while using your existing Swagger 2.0 specifications. Keep this in mind when referencing elements from a Swagger 2.0 file, as the structure and syntax might differ slightly in OpenAPI 3.0.

### Example: Referencing a Schema from Swagger 2.0

Here is a simple Swagger 2.0 specification file:

```yaml
swagger: '2.0'
info:
  title: A Swagger 2.0 specification file
  version: 1.0.0
paths:
  /pets:
    get:
      responses: 
        '200':
          description: A list of pets.
          schema:
            $ref: '#/definitions/Pet'
definitions: 
  Pet:
    type: object
```

In OpenAPI 3.0, you can reference the Pet schema from the Swagger 2.0 specification file as shown below:

```yaml
openapi: '3.0'
info:
  title: A OpenAPI 3.0 specification file
  version: 1.0.0
paths:
  /pets:
    get:
      responses: 
        '200':
          description: A list of pets.
          content:
              application/json:
                schema:
                  $ref: 'path/to/swagger.yaml#/components/schemas/Pet'
```

Swagger 2.0 uses #/definitions for internal schemas, while OpenAPI 3.0 utilizes #/components/schemas. However, Mokapi automatically resolves these differences for you.

With Mokapi's support for Swagger 2.0, you can modernize your API testing and mocking processes without abandoning your existing specifications. Whether you're transitioning to OpenAPI 3.0 or maintaining legacy systems, Mokapi makes the process seamless and efficient.