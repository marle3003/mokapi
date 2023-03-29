# OpenAPI

With an OpenAPI specification you can mock your HTTP API. By default, Mokapi uses
[Example Objects](https://swagger.io/docs/specification/adding-examples/)
and [Enums](https://swagger.io/docs/specification/data-models/enums/) defined in OpenAPI specification file to generate the response.
If neither an Example nor an Enum is defined, Mokapi generates random data based on the
[OpenAPI schema](https://swagger.io/docs/specification/data-models/). You can find more 
information about how Mokapi generates response data [here](/docs/openapi/static-data-generation)

However, if you need more dynamic responses, you can use Mokapi Script to generate responses
based on specific conditions or parameters. Mokapi Scripts allows you to create responses that can simulate a wide range of scenarios and edge cases.

## Swagger 2.0 support
Mokapi also supports Swagger 2.0 specification. After parsing a Swagger 2.0 file, mokapi converts it to OpenAPI 3.0

