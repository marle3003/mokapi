---
title: Mocking OpenAPI Security Schemes
description: Learn how Mokapi handles OpenAPI security schemes (HTTP Basic, Bearer, API keys, OAuth2), how to read credentials in a script and how to return 401 for invalid ones.
---

# Mocking OpenAPI Security Schemes

Mokapi reads the credentials that your OpenAPI security schemes describe and passes them to your
[Mokapi Script](/docs/javascript-api/overview.md). It does not reject a request on its own: a
request without credentials gets the same `200` response as one with credentials. If your mock
should return `401`, you decide that in a script, see
[Return 401 for invalid credentials](#return-401-for-invalid-credentials).

## What Mokapi does with a security scheme

For every request, Mokapi looks at the `security` requirements of the operation. If the operation
has no `security` key, it uses the global `security` of the specification. For each scheme it looks for the
credential:

| Scheme                                | Credential Mokapi looks for                              | Key in the script                |
|---------------------------------------|----------------------------------------------------------|----------------------------------|
| `http` with `scheme: basic`           | `Authorization` header                                   | `request.header['Authorization']`|
| `http` with `scheme: bearer`          | `Authorization` header                                   | `request.header['Authorization']`|
| `apiKey` with `in: header`            | header with the configured `name`                        | `request.header[name]`           |
| `apiKey` with `in: query`             | query parameter with the configured `name`               | `request.query[name]`            |
| `apiKey` with `in: cookie`            | cookie with the configured `name`                        | `request.cookie[name]`           |
| `oauth2`                              | `Authorization` header                                   | `request.header['Authorization']`|

Keep these points in mind:

- Mokapi only checks that the credential is present. It does not validate the format (the `Basic`
  or `Bearer` prefix, base64 encoding, JWT structure), the token, the OAuth2 flow or the scopes.
- The value is passed unchanged. For Basic and Bearer it includes the prefix, for example
  `Bearer 123`.
- If the credential is missing, the key is not set in `request`, so the script reads `undefined`.
  For `http` schemes, the request log of the [dashboard](/docs/http/dashboard.md) still lists the
  `Authorization` header, with an empty value.
- Header names are used exactly as written in the scheme. For an `apiKey` in a header with
  `name: X-API-KEY`, use `request.header['X-API-KEY']`. Mokapi finds the header regardless of the
  case the client sends.
- Schemes of type `openIdConnect` and `mutualTLS` are not supported. Mokapi logs a warning and
  ignores them. For `http` schemes other than `basic` and `bearer`, Mokapi does not read the
  credential.

> Credentials are only available in `request` when the scheme is part of an active `security`
> requirement. Without `security` on the operation or at the root of the specification,
> `request.header['Authorization']` is `undefined`, even if the client sends the header.

## Define security schemes

Declare the schemes in `components.securitySchemes`.

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

```yaml tab="API key in header"
components:
  securitySchemes:
    apiKeyHeader:
      type: apiKey
      in: header
      name: X-API-Key
```

```yaml tab="API key in query"
components:
  securitySchemes:
    apiKeyQuery:
      type: apiKey
      in: query
      name: api_key
```

```yaml tab="API key in cookie"
components:
  securitySchemes:
    apiKeyCookie:
      type: apiKey
      in: cookie
      name: auth_token
```

```yaml tab=OAuth2
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

Mokapi does not provide the authorization or token endpoints of an OAuth2 flow. A client sends the
access token in the `Authorization` header, and that's what Mokapi passes to your script.

## Apply security schemes

Add `security` at the root of the specification to apply it to all operations, or on an operation
to override the global setting.

```yaml tab=Global
security:
  - bearerAuth: []
```

```yaml tab=Operation
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

If the operation defines `security`, Mokapi ignores the global one for this operation. An operation
without a `security` key inherits the global requirement. An empty list (`security: []`) switches
the global requirement off: Mokapi evaluates no security scheme for this operation, so no credential
is copied into `request` and the request log shows no security parameter.

```yaml
paths:
  /health:
    get:
      security: [] # public, even if a global security is defined
      responses:
        '200':
          description: OK
```

Entries in the list are alternatives (OR). Mokapi uses the first entry whose credentials are all
present and ignores the remaining ones, so their credentials do not show up in `request`. Schemes
within one entry (`- bearerAuth: []` plus `apiKeyHeader: []` in the same item) must all be present
(AND), and Mokapi reads every scheme of that entry.

## Read credentials in a script

The names in `request` match the scheme definitions above.

```javascript tab="Header"
import { on } from 'mokapi'

export default function() {
  on('http', (request) => {
    console.log(request.header['X-API-Key'])
  })
}
```

```javascript tab="Query"
import { on } from 'mokapi'

export default function() {
  on('http', (request) => {
    console.log(request.query['api_key'])
  })
}
```

```javascript tab="Cookie"
import { on } from 'mokapi'

export default function() {
  on('http', (request) => {
    console.log(request.cookie['auth_token'])
  })
}
```

```javascript tab="Authorization"
import { on } from 'mokapi'

export default function() {
  on('http', (request) => {
    // Basic, Bearer and OAuth2: "Bearer <token>", "Basic <base64>"
    console.log(request.header['Authorization'])
  })
}
```

## Return 401 for invalid credentials

Check the credential in the event handler and set `response.statusCode`. The specification needs a
`401` (or `default`) response for the operation, otherwise Mokapi can't build the response and
returns an error.

```yaml
paths:
  /protected-resource:
    get:
      security:
        - bearerAuth: []
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
        '401':
          description: Unauthorized
```

```javascript
import { on } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    if (request.key !== '/protected-resource') {
      return
    }
    if (request.header['Authorization'] !== 'Bearer my-secret-token') {
      response.statusCode = 401
    }
  })
}
```

A request without the header or with another token now gets `401 Unauthorized`, and a request with
`Authorization: Bearer my-secret-token` gets the generated `200` response. To also return a body,
use [response.rebuild()](/docs/javascript-api/mokapi/eventhandler/httpresponse.md) with the `401`
status and set `response.data`.

## Security in the dashboard

The **Security** tab of an operation in the [dashboard](/docs/http/dashboard.md) lists the
security requirement that applies to that operation: the operation's own `security` if it defines
one, otherwise the global one. For an operation with `security: []` the tab shows no scheme. The request log shows the credentials Mokapi read, such as the
`Authorization` header or the API key, as parameters of the request.
