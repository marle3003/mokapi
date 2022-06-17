# OpenAPI

Mokapi uses the [OpenAPI 3.0](https://swagger.io/docs/specification/about/) specification
to mock any REST APIs. With [Example Objects](https://swagger.io/docs/specification/adding-examples/)
and [Enums](https://swagger.io/docs/specification/data-models/enums/) you can describe how
Mokapi responds to a request. If neither Example nor Enum are defined, Mokapi generates
random values based on the [OpenAPI schema](https://swagger.io/docs/specification/data-models/).
If the response should be dependent on the request, you should take a look at [Mokapi Actions](#/docs/actions/intro)

## Swagger 2.0 support
Mokapi also supports the Swagger 2.0 specification. After parsing a Swagger 2.0 file, mokapi converts it to OpenAPI 3.0

