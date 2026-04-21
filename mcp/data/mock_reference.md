# Mock Reference

Retrieves technical reference material for writing mock scripts.
You can request either 'types' (API definitions) or 'scenarios' (example blueprints).

To read a specific definition or scenario, call this tool again
with the corresponding category and name (e.g., category='types', name='http')

## Types

Overview of TypeScript definitions for mock event handlers.
- Use these types to ensure correct syntax for "import { ... } from 'mokapi/...'".

| Type                           | Description                                                                                                                                                                                         | Import Statement                      | Parameter Name       |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------|----------------------|
| **Core**                       | Exposes the core scripting API for Mokapi. It allows you to intercept and manipulate protocol events (HTTP, Kafka, LDAP, SMTP), schedule jobs, generate mock data, and share state between scripts. | `import {...} from 'mokapi'`          | mokapi               |
| **HTTP**                       | Exposes functions to invoke HTTP request                                                                                                                                                            | `import {...} from 'mokapi/http'`     | http                 |
| **Kafka**                      | Exposes functions to produce Kafka message                                                                                                                                                          | `import {...} from 'mokapi/kafka'`    | kafka                |
| **Faker**                      | Exposes functions to generate random data based on JSON schema                                                                                                                                      | `import {...} from 'mokapi/faker'`    | faker                |
| **Mustache**                   | Exposes to render output based on mustache template                                                                                                                                                 | `import {...} from 'mokapi/mustache'` | mustache             |
| **Yaml**                       | Exposes functions to parse or stringify YAML files                                                                                                                                                  | `import {...} from 'mokapi/yaml'`     | yaml                 |

## Scenarios

A list of code examples and boilerplates.
Use these to understand how to implement specific use cases like latency, error simulation or conditional behavior.
Always check a scenario before writing a script from scratch to ensure you follow best practices.

| Scenario                        | Description                                                                                                                                                                                                                                                                                                                       |                                                                                                                                                                                                                                                                                                                 
|---------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| dynamic-path-params             | Access and use path parameters (e.g., /pets/{petId}) to retrieve or process specific resources based on request.path values.                                                                                                                                                                                                      |
| conditional-response            | Return different responses based on request data (path, query, headers, or body), such as selecting resources, updating state, or handling different HTTP methods.                                                                                                                                                                |
| static-error-simulation         | Return predefined error responses (e.g., 400, 404, 500) for specific endpoints or conditions without dynamic logic.                                                                                                                                                                                                               |
| dynamic-error-simulation        | Return error responses based on runtime conditions, such as missing resources, validation failures, or conflicting state.                                                                                                                                                                                                         |
| delay-latency                   | Simulate network latency or slow backend processing by delaying the response before returning data or errors.                                                                                                                                                                                                                     |
| forward-request-to-real-backend | Stop API drift in its tracks. Use Mokapi as a validation layer to enforce OpenAPI contracts between clients and backends, regardless of who's calling or what they're building. This scenario forwards incoming requests to real backend services while validating both requests and responses against the OpenAPI specification. |