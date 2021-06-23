# Configuration

Mokapi has two types of configuration:
- The startup configuration (referred as the static configuration)
- The dynamic configuration (e.g. OpenApi configuration)

## Dynamic Configuration

The dynamic configuration contains everything that defines a service like an OpenApi configuration. This configuration can change during runtime and is seamlessly hot-reloaded.

Mokapi gets its dynamic configuration from providers. Each service has its own configuration format.

- [OpenApi](https://swagger.io/docs/specification/about/) Version 3.0
- [AsyncApi](https://www.asyncapi.com/docs/specifications/v2.0.0)
- LDAP
- [Mokapi](#/docs/actions/intro)

### File Provider
The file provider lets you define the dynamic configuration in a YAML or JSON file. Depending on service you can split your configuration in multiple files

`filename` defines the path to the configuration file

`directory` Defines the path to de directory that contains the configuration files. Mokapi includes all subdirectories

### GIT Provider
Provide your configuration files via a GIT repository for a example a github repository.

`url` URL to the GIT repository

`pullInterval` Defines the pull interval, default '5s'

### HTTP Provider
Provide your configuration via HTTP URL

`url` URL to your configuration file

`pollInterval` Defines the poll interval, default '5s'

## Static Configuration

Elements in the static configuration don't often change and changes require a restart of the application.

There are two different ways to define a static configuration options.
1. Command-line arguments
2. Environment variables

These ways are evaluated in the order listed above.

### CLI variables

`--log.level`
Log level: `debug` | `info` | `error` (Default: `error`)

`--log.format`
Log format: `json` | `default`

`--providers.file.filename`
Load dynamic configuration from a file

`--providers.file.directory`
Load dynamic configuration from one or more files in a directory

`--providers.git.url`
URL to the GIT repository

`--providers.git.pullInterval`
Defines the pull interval, default '5s'

`--providers.http.url`
URL to your configuration file

`--providers.http.pollInterval`
Defines the poll interval, default '5s'

`--api.dashboard`
Enables API/dashboard. (Default `true`)

`--api.port`
Dashboard's port (default: `8080`)

### Environment variables

`MOKAPI_Log.Level`
Log level: `debug` | `info` | `error` (Default: `error`)

`MOKAPI_Log.Format`
Log format: `json` | `default`

`MOKAPI_Providers.File.Filename`
Load dynamic configuration from a file

`MOKAPI_Providers.File.Directory`
Load dynamic configuration from one or more files in a directory

`MOKAPI_Providers.Git.Url`
URL to the GIT repository

`MOKAPI_Providers.Git.PullInterval`
Defines the pull interval, default '5s'

`MOKAPI_Providers.Http.Url`
URL to your configuration file

`MOKAPI_Providers.Http.pollInterval`
Defines the poll interval, default '5s'

`MOKAPI_Api.Dashboard`
Enables API/dashboard. (Default `true`)

`MOKAPI_Api.Port`
Dashboard's port (default: `8080`)
