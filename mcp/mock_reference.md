# Mock Reference

Retrieves technical reference material for writing mock scripts.
You can request either 'types' (API definitions) or 'scenarios' (example blueprints).

To read a specific definition or scenario, call this tool again
with the corresponding category and name (e.g., category='types', name='http')

## Types

Overview of TypeScript definitions for mock event handlers.
- Use these types to ensure correct syntax for "import { ... } from 'mokapi/...'".

| Type                           | Import Statement                      | Parameter Name |
|--------------------------------|---------------------------------------|----------------|
| **Core**                       | `import {...} from 'mokapi'`          | mokapi         |
| **HTTP**                       | `import {...} from 'mokapi/http'`     | http           |
| **Kafka**                      | `import {...} from 'mokapi/kafka'`    | kafka          |
| **Faker**                      | `import {...} from 'mokapi/faker'`    | faker          |
| **Mustache**                   | `import {...} from 'mokapi/mustache'` | mustache       |
| **Yaml**                       | `import {...} from 'mokapi/yaml'`     | yaml           |

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