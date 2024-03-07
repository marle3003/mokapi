---
title: How to mock HTTP APIs with Mokapi
description: Mock any HTTP API with OpenAPI specification
---
# Overview

You can mock an HTTP API in Mokapi with the [OpenAPI specification](https://swagger.io/specification/).
Mokapi produces an HTTP response depending on the specification and generates data randomly.

However, if you need more dynamic responses, you can use Mokapi Script to generate responses
based on specific conditions or parameters. Mokapi Scripts allows you to create responses that 
can simulate a wide range of scenarios and edge cases.

- [Declarative Data Generation](/docs/guides/get-started/test-data.md)
- [Mokapi Scripts](/docs/guides/get-started/test-data.md)

## Swagger 2.0 support
Mokapi also supports Swagger 2.0 specification. After parsing a Swagger 2.0 file, mokapi converts it to OpenAPI 3.0

