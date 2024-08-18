---
title: How to run your first mock REST API
description: Learn how to mock a REST API and analyze HTTP request and response in the dashboard.
---
# How to run your first mock REST API

This quick start shows you how to start Mokapi and mock an HTTP REST API. We will use [Swagger](https://swagger.io/)'s 
Petstore specification to mock your first API with Mokapi. If you have not installed Mokapi yet, visit 
[Installation](/docs/guides/get-started/installation.md)

## Mocking Petstore API
Start Mokapi with the following command:

```  
mokapi --providers-http-url https://petstore3.swagger.io/api/v3/openapi.json
```

This starts Mokapi with the Petstore on port 80 and the dashboard on port 8080.
Let's check to see if it's working:

```
curl --header "Accept: application/json" http://localhost/api/v3/pet/4
```

Mokapi will create a response with randomly generated data.

``` box=tip
When you use Mokapi behind a corporate proxy, you probably need 
to skip SSL verification: "--providers-http-tls-skip-verify".
```

## Analyze API and HTTP Request/Response

Mokapi provides a dashboard where you can see your mocked Petstore API and analyze your HTTP request.
Open `http://localhost:8080` from your browser.

## Read more

- [Test-Data](/docs/guides/get-started/test-data.md)
- [Dashboard](/docs/guides/get-started/dashboard.md)
- [Mocking HTTP API](/docs/guides/http/overview.md)

