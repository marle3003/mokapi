# Available Mocking Libraries

Overview of TypeScript definitions for mock event handlers.
Provides URIs to technical definitions (d.ts) for 'mokapi', 'http', 'kafka', etc.
- Use these types to ensure correct syntax for "import { ... } from 'mokapi/...'".
- Refer to mokapi://lib/mocking/scenarios for boilerplate examples and usage patterns.
  Mandatory for generating valid mock scripts.

| Library      | Resource URI                          | Import Statement                      |
|--------------|---------------------------------------|---------------------------------------|
| **Core**     | `mokapi://lib/mocking/types/mokapi`   | `import {...} from 'mokapi'`          |
| **HTTP**     | `mokapi://lib/mocking/types/http`     | `import {...} from 'mokapi/http'`     |
| **Kafka**    | `mokapi://lib/mocking/types/kafka`    | `import {...} from 'mokapi/kafka'`    |
| **Faker**    | `mokapi://lib/mocking/types/faker`    | `import {...} from 'mokapi/faker'`    |
| **Mustache** | `mokapi://lib/mocking/types/mustache` | `import {...} from 'mokapi/mustache'` |
| **Yaml**     | `mokapi://lib/mocking/types/yaml`     | `import {...} from 'mokapi/yaml'`     |