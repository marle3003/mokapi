# Configuration

Mokapi has two types of configuration:
- The startup configuration (referred as the static configuration)
- The dynamic configuration (e.g. OpenApi configuration)

## Dynamic Configuration

The dynamic configuration contains everything that defines a service like an
OpenApi configuration. This configuration can change during runtime and is
seamlessly hot-reloaded.

Mokapi gets its dynamic configuration from providers. Currently, only the 
file provider is supported. Each service has its own configuration format.

- OpenApi
- AsyncApi  
- LDAP

## Static Configuration

Elements in the static configuration don't often change and changes require
a restart of the application.

There are two different ways to define a static configuration options.
1. Command-line arguments
2. Environment variables

These ways are evaluated in the order listed above.

### CLI variables

`--log.level`:
Log level: `debug` | `info` | `error` (Default: `error`)

`--log.format`:
Log format: `json` | `default`

`--log.providers.file.filename`:
Load dynamic configuration from a file

`--log.providers.file.directory`:
Load dynamic configuration from one or more files in a directory

`--api.dashboard`:
Enables API/dashboard. (Default `true`)

`--api.port`:
Dashboard's port (default: `8080`)

### Environment variables

`MOKAPI_LOG_LEVEL`:
Log level: `debug` | `info` | `error` (Default: `error`)

`MOKAPI_LOG_FORMAT`:
Log format: `json` | `default`

`MOKAPI_PROVIDERS_FILE`:
Load dynamic configuration from a file

`MOKAPI_PROVIDERS_DIRECTORY`:
Load dynamic configuration from one or more files in a directory

`MOKAPI_API_DASHBOARD`:
Enables API/dashboard. (Default `true`)

`MOKAPI_API_PORT`:
Dashboard's port (default: `8080`)