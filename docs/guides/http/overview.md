---
title: How to mock HTTP APIs with Mokapi
description: Mock any HTTP API with OpenAPI specification
---
# Mocking HTTP APIs

You can mock an HTTP API in Mokapi with the [OpenAPI specification](https://swagger.io/specification/).
Mokapi produces an HTTP response depending on the specification and generates data randomly.

## Launch Mokapi with Swaggers PetStore API

This example uses the API specification directly from [swagger.io](https://swagger.io).

```bash tab=CLI
mokapi --providers-http-url https://petstore3.swagger.io/api/v3/openapi.json
```

**That's it**. You can open a browser and go to [http://localhost/api/v3/pet/12](http://localhost/api/v3/pet/12) to see 
Mokapi's generated API response.

``` box=info
You may need to configure a proxy server if you're having trouble fetching the specification 
file from swagger.io: --providers-http-proxy http://proxy.server.com:port
```

## Control Mokapi's HTTP response

If you need more dynamic responses, you can use Mokapi Script to generate responses
based on specific conditions or parameters. Mokapi Scripts allows you to create responses that 
can simulate a wide range of scenarios and edge cases.

Create a petstore.ts file where you will define the response for /pet/12.

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

Start your mock server with the following command:

```bash tab=CLI
mokapi --providers-http-url https://petstore3.swagger.io/api/v3/openapi.json --providers-file-filename /path/to/petstore.ts
```

Browse [http://localhost/api/v3/pet/12](http://localhost/api/v3/pet/12) and see Mokapi's response with Garfield as a cat. 
Requests with a different pet ID will still return random data.

More information about data generation:

- [Declarative Data Generation](/docs/guides/get-started/test-data.md)
- [Mokapi Scripts](/docs/guides/get-started/test-data.md)

## Swagger 2.0 support
Mokapi also supports Swagger 2.0 specification. After parsing a Swagger 2.0 file, mokapi converts it to OpenAPI 3.0

