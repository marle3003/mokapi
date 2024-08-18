---
title: How to run Mokapi
description: Here you will learn how to run Mokapi
---
# How to run Mokapi

This quick start shows you how to start Mokapi and mock an HTTP REST API. We will use [Swagger](https://swagger.io/)'s 
Petstore to mock your first API with Mokapi. If you have not installed Mokapi yet, visit [Installation](/docs/guides/get-started/installation.md)

## Start Mokapi
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

## Read more

- [HTTP API](/docs/guides/http/overview.md)
- [Test-Data](/docs/guides/get-started/test-data.md)
- [Dashboard](/docs/guides/get-started/dashboard.md)

