---
title: Running Your First Mocked REST API
description: Learn how to mock a REST API and analyze HTTP request and response in the dashboard.
---
# Running Your First Mocked REST API with Mokapi

This quick-start guide will help you set up Mokapi and mock an HTTP REST API using 
the [Swagger Petstore](https://swagger.io/) specification. If you haven’t installed 
Mokapi yet, check out the [Installation Guide](/docs/guides/get-started/installation.md).

## Mocking Petstore API
Run the following command to start Mokapi using the Petstore OpenAPI specification directly from the web:

```  
mokapi --providers-http-url https://petstore3.swagger.io/api/v3/openapi.json
```

``` box=info
When you use Mokapi behind a corporate proxy, you probably need 
to skip SSL verification: "--providers-http-tls-skip-verify".
```

This launches Mokapi with the Petstore API on port 80 and the dashboard on port 8080. 
The mock API is accessible at /api/v3, as defined in the OpenAPI specification. 
Mokapi supports many options that can be passed in multiple places, check 
[options reference](/docs/configuration/reference.md) for more info.

## Test Your Mock REST API

You can now send a request to the mock API and receive a dynamically generated response:

```
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```

Mokapi generates a random response based on the OpenAPI specification. If multiple 
response formats are available, specifying the Accept header ensures you get the expected format.

``` box=info
 If no Accept header is sent, Mokapi will randomly select a content type 
 from the specification, which may lead to unexpected formats in responses.
```

## Analyze API and HTTP Request/Response

Mokapi provides a dashboard where you can see your mocked Petstore API and analyze your HTTP request.
Open `http://localhost:8080` in your browser.

After navigating to the Petstore REST API, you can see all mocked endpoints including your logged request.

<img src="/petstore-rest-api-endpoints.jpg" width="700" alt="Dashboard shows all mocked Petstore's endpoints including logged requests." title="" />

You can analyze your request and the corresponding response by clicking on the logged request in the *Recent Requests* table.

<img src="/petstore-rest-api-request.jpg" width="700" alt="Logged request to the mocked Petstore REST API containing headers, body and metrics" title="" />

## Next Steps

- [Test-Data](/docs/guides/get-started/test-data.md)
- [Dashboard](/docs/guides/get-started/dashboard.md)
- [Mocking HTTP API](/docs/guides/http)

