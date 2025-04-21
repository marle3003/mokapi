---
title:  "Mock OpenAPI 3.1 Authentication: API Key & Bearer Token"
description: "Mock OpenAPI 3.1 authentication with API Key & Bearer Token. Customize authentication using Mokapi Scripts for secure API testing."
icon: bi-key
tech: http
---

# OpenAPI 3.1 Authentication: API Key & Bearer Token

## Introduction

Introduction
This tutorial demonstrates how to mock OpenAPI 3.1 authentication using two widely-used methods: API Key and Bearer 
Token (JWT). With Mokapi you can simulate authentication behaviors without needing a real backend.

By following this guide, you will:

- Learn how to define API authentication schemes in OpenAPI 3.1.
- Understand how to use Mokapi Scripts to manage authentication headers.
- Be able to test the mock API with curl commands to verify authentication logic. 

This approach is useful for testing and simulating secured APIs during development, enabling teams to ensure proper authentication functionality before integration.

## 1. Create the OpenAPI Specification

To get started, we first define the OpenAPI 3.1 specification that describes two 
authentication methods: API Key and Bearer Token.

Create a file named `openapi.yaml` with the following content:

```yaml tab=openapi.yaml
openapi: 3.1.0
info:
  title: Secure API
  version: 1.0.0
paths:
  /protected/apikey:
    get:
      summary: Protected with API Key
      security:
        - apiKeyAuth: [] # use the same name as under securitySchemes
      responses:
        "204":
          description: Successful response for API Key authentication
        "401":
          description: Unauthorized
  /protected/bearer:
    get:
      summary: Protected with Bearer Token
      security:
        - bearerAuth: [] # use the same name as under securitySchemes
      responses:
        "204":
          description: Successful response for Bearer authentication
        "401":
          description: Unauthorized
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: API-Key
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

Explanation:

- API Key Authentication is used in the header (X-API-Key).
- Bearer Token Authentication is used with a JWT token in the Authorization header.

### Global Security (Optional)

If most of your API requires authentication, you can define global security requirements instead of per-path:

```yaml
# Apply the API key globally to all operations
security:
  # use the same name as under securitySchemes
  - apiKeyAuth: []
  - bearerAuth: []
```

This means clients must send either an API key or a Bearer token.

## 2. Customize Authentication with Mokapi Scripts

Next, we create a Mokapi script to handle the authentication for both methods.

Create a new file named `auth.js` with the following content:

```javascript tab=auth.js
import { on } from 'mokapi';

export default function() {
    on('http', function(request, response) {
        switch (request.key) {
            case '/protected/apikey':
                if (request.header['API-Key'] === 'my-secret-api-key') {
                    // because mokapi selects first successfull response
                    // we don't need to do anything here
                } else {
                    response.statusCode = 401
                }
                return true
            case '/protected/bearer':
                if (authHeader && authHeader.startsWith("Bearer ")) {
                    const token = authHeader.split(" ")[1];
                    if (token === "valid-jwt-token") {
                        // Token is valid, proceed
                    } else {
                        response.statusCode = 401
                    }
                }
                return true
        }
    })
}
```

Explanation:

- For /protected/apikey, we check if the request header contains a valid API-Key.
- For /protected/bearer, we check the Authorization header for a valid Bearer token (in this case, valid-jwt-token).

## 3. Run Mokapi Locally

Now that we have both the OpenAPI specification and authentication script, we can start Mokapi with the following command:

```shell
mokapi openapi.yaml auth.js
```

This will start the mock server, which listens for requests on the endpoints defined in api.yaml (default on port 80) 
and applies the authentication logic from auth.js.

## 4. Test the Mock API

Once the server is running, we can test the authentication using curl commands.

```shell
curl -i -H "API-Key: my-secret-api-key" http://localhost/protected/apikey
```

Expected output:

```shell
HTTP/1.1 204 No Content
```

Test Bearer Token Authentication

```shell
curl -i -H "Authorization: Bearer valid-jwt-token" http://localhost/protected/bearer
```

Expected output:

```shell
HTTP/1.1 204 No Content
```

Test invalid authentication

```shell
curl -i -H "API-Key: wrong-api-key" http://localhost/protected/apikey
```

Expected output:

```shell
HTTP/1.1 401 Unauthorized
Content-Length: 0
```

## 5. Conclusion

In this tutorial, we:
- Created an OpenAPI 3.1 specification with API Key and Bearer Token authentication.
- Used Mokapi and Mokapi Scripts to mock the authentication process.
- Tested the mock API with simple curl commands.