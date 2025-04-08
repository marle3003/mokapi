---
title: "Mocking OpenAPI Security: HTTP, OAuth2, and API Keys with Mokapi"
description: Learn how to mock OpenAPI security schemes, including HTTP Basic, Bearer, OAuth2, and API Keys (header, query, cookie) using Mokapi. Perfect for testing authentication flows!
---

# Mocking OpenAPI Security Schemes with Mokapi

## Introduction

OpenAPI provides multiple security schemes to secure APIs, including HTTP authentication (Basic and Bearer), OAuth2, and API keys passed via headers, query parameters, or cookies. Mokapi can be used to mock these authentication mechanisms for testing and development.

## Defining Security Schemes in OpenAPI

### HTTP Authentication

OpenAPI supports both Basic Authentication and Bearer Token Authentication under the http scheme.

```yaml tab=Basic
components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic
```

```yaml tab=Bearer
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
```
``` box=info
Mokapi will respond with an error when header Authorization is missing.
```

Access in Mokapi Script:

```javascript
import { on } from 'mokapi'

export default function() {
    on('http', (request) => {
        console.log(request.header['Authorization']);
    });
};
```

### OAuth2 Authentication

OAuth2 allows authentication with various flows. Here’s an example using the authorizationCode flow:

```yaml
components:
  securitySchemes:
    oauth2Auth:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/oauth/authorize
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: Read access
            write: Write access
```

### API Key Authentication

API keys can be passed in different ways: header, query parameter, or cookie.

API Key in Header

```yaml tab="Config"
components:
  securitySchemes:
    apiKeyHeader:
      type: apiKey
      in: header
      name: X-API-Key
```

```javascript tab="Script"
import { on } from 'mokapi'

export default function() {
    on('http', (request) => {
        console.log(request.header['X-API-Key']);
    });
};
```

API Key in Query Parameter

```yaml tab="Config"
components:
  securitySchemes:
    apiKeyQuery:
      type: apiKey
      in: query
      name: api_key
```

```javascript tab="Script"
import { on } from 'mokapi'

export default function() {
    on('http', (request) => {
        console.log(request.query['X-API-Key']);
    });
};
```

API Key in Cookie

```yaml tab="Config"
components:
  securitySchemes:
    apiKeyCookie:
      type: apiKey
      in: cookie
      name: auth_token
```

```javascript tab="Script"
import { on } from 'mokapi'

export default function() {
    on('http', (request) => {
        console.log(request.cookie['X-API-Key']);
    });
};
```

## Applying Security Schemes to Endpoints

You can apply security schemes globally or per endpoint.

### Global Security

```yaml
security:
  - basicAuth: []
  - bearerAuth: []
  - oauth2Auth:
      - read
  - apiKeyHeader: []
```

### Per-Endpoint Security

```yaml
paths:
  /protected-resource:
    get:
      security:
        - bearerAuth: []
        - apiKeyQuery: []
      responses:
        '200':
          description: Successful response
```

## Conclusion

By using Mokapi, you can mock various authentication mechanisms in OpenAPI, allowing for easier testing without requiring a real authentication provider. Whether you are using Basic, Bearer, OAuth2, or API Key authentication, Mokapi provides a flexible way to simulate secure APIs.