---
title:  Secure API Authentication with OAuth2 and JWT in OpenAPI 3.1
description: Learn how to secure your API using OAuth2 and JWT authentication in OpenAPI 3.1. This tutorial covers defining security schemes, validating JWT tokens with Mokapi, and generating tokens in Bash for API requests.
icon: bi-shield-lock
---

# OpenAPI 3.1 Authentication: OAuth2 and JWT

In this tutorial, we'll demonstrate how to secure an API using OAuth2 authentication and JWT tokens in OpenAPI 3.1. We'll mock an authentication service with Mokapi, generate a JWT token in Bash, and use it to make an authenticated request.

## Step 1: Define OpenAPI Security

First, we define the security scheme in our OpenAPI 3.1 specification.

```yaml
openapi: 3.1.0
info:
  title: Secure API
  version: 1.0.0
paths:
  /protected:
    get:
      summary: Protected with API Key
      security:
        - oauth2Auth: [read, write] # use the same name as under securitySchemes
      responses:
        "200":
          description: Successful response for API Key authentication
          content:
            'application/json':
              schema: {}
        "401":
          description: Unauthorized
          content:
            'application/json':
              schema: { }
        "403":
          description: Insufficient scope
          content:
            'application/json':
              schema: {}
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

This configuration defines an OAuth2 authorization code flow with read and write scopes.

## Step 2: Implement Authentication Logic in Mokapi

We use Mokapi scripts to validate JWT tokens and enforce scope-based access control.

```javascript
import { on } from 'mokapi'
import { base64 } from 'mokapi/encoding'

export default function() {
    on('http', function(request, response) {
        const authHeader = request.header['Authorization'];
        if (authHeader && authHeader.startsWith('Bearer ')) {
            const token = authHeader.substring(7);
            try {
                const { payload } = decodeJWT(token)
                if (payload.scopes && payload.scopes.includes('read')) {
                    console.log('Access granted: read scope');
                } else {
                    response.statusCode = 403;
                    response.data = { error: 'Insufficient scope' };
                }
            } catch (err) {
                response.statusCode = 401;
                response.data = { error: 'Invalid token' };
            }
        } else {
            response.statusCode = 401;
            response.data = { error: 'Unauthorized' };
        }
        return true
    });
};

function decodeJWT(jwt) {
    const parts = jwt.split('.');
    console.log(parts.length)
    if (parts.length !== 3) {
        throw new Error('Invalid JWT format');
    }
    const [encodedHeader, encodedPayload, signature] = parts;
    const header = JSON.parse(base64UrlDecode(encodedHeader));
    const payload = JSON.parse(base64UrlDecode(encodedPayload));
    return { header, payload, signature };
}

// Base64URL decode function
function base64UrlDecode(base64Url) {
    let b = base64Url.replace(/-/g, '+').replace(/_/g, '/');  // Replace URL-safe chars with standard Base64
    const padding = base64.length % 4 === 0 ? '' : '='.repeat(4 - base64.length % 4); // Add padding if needed
    return base64.decode(b + padding)
}
```

This script:

- Extracts and decodes the JWT token from the request header.
- Validates the token's scope.
- Returns a 403 (Forbidden) response if the user lacks the required scope.
- Returns a 401 (Unauthorized) response if the token is missing or invalid.

## Step 3: Generate a JWT Token in Bash

We create a JWT token using HMAC SHA-256 signing and use it to authenticate API requests.

```shell
#!/bin/bash

# Your secret key (used to sign the JWT token)
SECRET_KEY="your_secret_key_here"

# Header (JWT header in JSON)
HEADER='{
  "alg": "HS256",
  "typ": "JWT"
}'

# Payload (Claims or your data, including scopes)
PAYLOAD='{
  "sub": "1234567890",
  "name": "John Doe",
  "iat": '$(date +%s)',
  "scopes": ["read", "write", "delete"]
}'

# Encode the header and payload to Base64Url (replace + and / with - and _)
encode_base64url() {
    echo -n "$1" | openssl base64 -e -A | tr '+/' '-_' | tr -d '='
}

# Create Base64Url encoded header and payload
HEADER_ENCODED=$(encode_base64url "$HEADER")
PAYLOAD_ENCODED=$(encode_base64url "$PAYLOAD")

# Create the message to be signed (Header and Payload combined)
MESSAGE="$HEADER_ENCODED.$PAYLOAD_ENCODED"

# Create the signature using HMAC SHA-256 (your secret key)
SIGNATURE=$(echo -n "$MESSAGE" | openssl dgst -sha256 -hmac "$SECRET_KEY" -binary | openssl base64 -e | tr '+/' '-_' | tr -d '=')

# Combine Header, Payload, and Signature to form the JWT
JWT="$MESSAGE.$SIGNATURE"

# Output the JWT
echo "Generated JWT Token: $JWT"

# Make the request
curl -X GET 'http://localhost/protected' \
     -H 'Authorization: Bearer $JWT'
```

This script:

1. Encodes the JWT header and payload in Base64Url format. 
2. Generates the HMAC SHA-256 signature. 
3. Combines the three JWT components. 
4. Sends a GET request to /protected with the JWT in the Authorization header.

## Step 4: Test the Authentication

Run the Bash script to generate a JWT and make the request:

```shell
chmod +x test.sh
./test.sh
```

If the token is valid and contains the read scope, the response should be:

```json
{
  "message": "Access granted: read scope"
}
```

If the token is missing, invalid, or lacks proper scope, an error message is returned.

## Conclusion

In this tutorial, we:
- Defined OAuth2 authentication in OpenAPI 3.1. 
- Implemented JWT validation in Mokapi. 
- Created and signed a JWT token in Bash. 
- Used the token to authenticate API requests.