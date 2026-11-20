---
title: mokapi CLI Flags
description: A complete list of all mokapi flags, with descriptions, defaults, and examples of how to set the option in config file, environment variables, or CLI.
---

# mokapi

Start Mokapi and serve mocked APIs

## Description

Mokapi is an easy, modern and flexible API mocking tool using Go and Javascript.

## Usage

```bash tab=Bash
mokapi [flags] [CONFIG-URL|DIRECTORY|FILE]...
```

## Commands

| Command | Description |
|---------|-------------|
| [sample-data](#sample-data) | Generate sample data from a schema |

<div class="flags">

## Flags

### Log

| Flag | Description |
|------|-------------|
| --[log-level](#log-level) | Set log level (debug\|info\|warn\|error) |
| --[log-format](#log-format) | Set log output format (text\|json) |

### API & Dashboard

| Flag | Description |
|------|-------------|
| --[api-port](#api-port) | Port for the API server |
| --[api-path](#api-path) | Path prefix for the API and dashboard |
| --[api-base](#api-base) | Base path used when the API is behind a reverse proxy |
| --[api-dashboard](#api-dashboard) | Enable the web dashboard |
| --[api-search-enabled](#api-search-enabled) | Enable search functionality in the dashboard |
| --[api-search-index-path](#api-search-index-path) | Set storage location for the dashboard search index |
| --[api-search-in-memory](#api-search-in-memory) | Store the dashboard search index entirely in memory |

### MCP Server

| Flag | Description |
|------|-------------|
| --[mcp-server-enabled](#mcp-server-enabled) | Enable the MCP server |
| --[mcp-server-port](#mcp-server-port) | Port for the MCP server |
| --[mcp-server-path](#mcp-server-path) | Path for the MCP server endpoint |

### Health

| Flag | Description |
|------|-------------|
| --[health-enabled](#health-enabled) | Enables or disables the health check endpoint entirely. |
| --[health-port](#health-port) | The port on which the health endpoint is exposed. |
| --[health-path](#health-path) | The HTTP path for the health endpoint (default: /health). |
| --[health-log](#health-log) | Controls whether HTTP requests to the health endpoint are logged. |

### File Provider

| Flag | Description |
|------|-------------|
| --[providers-file](#providers-file) | Configure a File-based provider using shorthand syntax |
| --[providers-file-filename](#providers-file-filename) | Load dynamic configuration from a file |
| --[providers-file-filenames](#providers-file-filenames) | Load dynamic configuration from a file |
| --[providers-file-directory](#providers-file-directory) | Load the dynamic configuration from directories |
| --[providers-file-directories](#providers-file-directories) | Load the dynamic configuration from directories |
| --[providers-file-skip-prefix](#providers-file-skip-prefix) | One or more prefixes that indicate whether a file or directory should be skipped. |
| --[providers-file-include](#providers-file-include) | One or more patterns that a file must match, except when empty |
| --[providers-file-exclude](#providers-file-exclude) | Exclude files or directories matching patterns |

### HTTP Provider

| Flag | Description |
|------|-------------|
| --[providers-http](#providers-http) | Configure an HTTP-based provider using shorthand syntax |
| --[providers-http-url](#providers-http-url) | Fetch configuration from an HTTP endpoint |
| --[providers-http-urls](#providers-http-urls) | Fetch configurations from HTTP endpoints |
| --[providers-http-poll-interval](#providers-http-poll-interval) | Polling interval for HTTP endpoints |
| --[providers-http-poll-timeout](#providers-http-poll-timeout) | Timeout for HTTP polling requests |
| --[providers-http-proxy](#providers-http-proxy) | HTTP proxy URL |
| --[providers-http-tls-skip-verify](#providers-http-tls-skip-verify) | Skip TLS certificate verification |
| --[providers-http-ca](#providers-http-ca) | Custom certificate authority file (default: system certification pool) |

### Git Provider

| Flag | Description |
|------|-------------|
| --[providers-git](#providers-git) | Configure a Git-based provider using shorthand syntax |
| --[providers-git-url](#providers-git-url) | Clone configuration from a Git repository |
| --[providers-git-urls](#providers-git-urls) | Clone configuration from Git repositories |
| --[providers-git-pull-interval](#providers-git-pull-interval) | Interval for pulling updates from Git repositories |
| --[providers-git-temp-dir](#providers-git-temp-dir) | Temporary directory used for Git checkouts |
| --[providers-git-repository](#providers-git-repository) | Configure a Git repository using shorthand syntax |
| --[providers-git-repositories](#providers-git-repositories) | Configure Git repositories using shorthand syntax |

### NPM Provider

| Flag | Description |
|------|-------------|
| --[providers-npm](#providers-npm) | Configure an npm-based provider using shorthand syntax |
| --[providers-npm-global-folder](#providers-npm-global-folder) | Load configuration from a global npm folder |
| --[providers-npm-global-folders](#providers-npm-global-folders) | Load configuration from a global npm folder |
| --[providers-npm-package](#providers-npm-package) | Configure an npm package using shorthand syntax |
| --[providers-npm-packages](#providers-npm-packages) | Configure npm packages using shorthand syntax |

### Certificates

| Flag | Description |
|------|-------------|
| --[root-ca-cert](#root-ca-cert) | Root CA certificate used for signing generated certificates |
| --[root-ca-key](#root-ca-key) | Private key of the root CA |

### Events Store

| Flag | Description |
|------|-------------|
| --[event-store-default-size](#event-store-default-size) | Default maximum number of stored events per API |
| --[event-store](#event-store) | Configure event store using shorthand syntax |

### Data Generator

| Flag | Description |
|------|-------------|
| --[data-gen-optional-properties](#data-gen-optional-properties) | Probability for generating optional properties |

### General

| Flag | Description |
|------|-------------|
| --[config](#config) | Provide inline configuration data |
| --[configs](#configs) | Provide inline configuration data |
| --[config-file](#config-file) | Read configuration from a file |
| --[generate-cli-skeleton](#generate-cli-skeleton) | Generate a configuration skeleton and exit |
| --[help](#help) | Show help information and exit |
| --[version](#version) | Show version information and exit |

### Config file only

| Option | Description |
|--------|-------------|
| [javascript.modules](#javascript-modules) | Module aliases for JavaScript imports |
| [certificates](#certificates) | Static TLS certificates |

</div>

## <a id="log-level" name="log-level"></a>log-level

The default level of log messages is info. You can set the log level to one of the following, listed in order of least to most information. The level is cumulative: for the debug level, the log file also includes messages at the info, warn, and error levels.
- Debug
- Info
- Warn
- Error


| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --log-level | MOKAPI_LOG_LEVEL | log.level | string | info |

```bash tab=CLI
--log-level warn
```
```bash tab=Env
MOKAPI_LOG_LEVEL=warn
```
```yaml tab=File
log:
  level: warn
```

## <a id="log-format" name="log-format"></a>log-format

Defines how Mokapi writes log messages. Use text for human-readable output or json to get one JSON object per line, which is easier to process with log collectors.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --log-format | MOKAPI_LOG_FORMAT | log.format | string | text |

```bash tab=CLI
--log-format json
```
```bash tab=Env
MOKAPI_LOG_FORMAT=json
```
```yaml tab=File
log:
  format: json
```

## <a id="api-port" name="api-port"></a>api-port

Specifies the TCP port on which the Mokapi API server listens.
The API server is the central entry point where developers can access all mocked services, events, and configuration data exposed by Mokapi. It also serves the web dashboard when enabled.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-port | MOKAPI_API_PORT | api.port | int | 8080 |

```bash tab=CLI
--api-port 5000
```
```bash tab=Env
MOKAPI_API_PORT=5000
```
```yaml tab=File
api:
  port: 5000
```

## <a id="api-path" name="api-path"></a>api-path

Defines a path prefix under which the API and web dashboard are served.
This is useful when Mokapi is hosted behind a reverse proxy and needs to be accessible under a specific URL path.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-path | MOKAPI_API_PATH | api.path | string | - |

```bash tab=CLI
--api-path /mokapi/dashboard
```
```bash tab=Env
MOKAPI_API_PATH=/mokapi/dashboard
```
```yaml tab=File
api:
  path: /mokapi/dashboard
```

## <a id="api-base" name="api-base"></a>api-base

Specifies the external base path used to access the API when Mokapi is running behind a reverse proxy.
This value is used to generate correct URLs in responses and in the web dashboard, and may differ from the internal api-path.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-base | MOKAPI_API_BASE | api.base | string | - |

```bash tab=CLI
--api-base /mokapi/dashboard
```
```bash tab=Env
MOKAPI_API_BASE=/mokapi/dashboard
```
```yaml tab=File
api:
  base: /mokapi/dashboard
```

## <a id="api-dashboard" name="api-dashboard"></a>api-dashboard

Enables or disables the Mokapi web dashboard.
When disabled, the API server continues to run, but the dashboard UI is not exposed.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-dashboard | MOKAPI_API_DASHBOARD | api.dashboard | bool | true |

```bash tab=CLI
--api-dashboard true
--api-dashboard
--no-api-dashboard
```
```bash tab=Env
MOKAPI_API_DASHBOARD=false
```
```yaml tab=File
api:
  dashboard: false
```

## <a id="api-search-enabled" name="api-search-enabled"></a>api-search-enabled

Enables search functionality in the web dashboard.
When enabled, users can search through mocked APIs, resources, and requests directly from the dashboard UI.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-search-enabled | MOKAPI_API_SEARCH_ENABLED | api.search.enabled | bool | true |

```bash tab=CLI
--api-search-enabled true
--api-search-enabled
--no-api-search-enabled
```
```bash tab=Env
MOKAPI_API_SEARCH_ENABLED=true
```
```yaml tab=File
api:
  search:
    enabled: true
```

## <a id="api-search-index-path" name="api-search-index-path"></a>api-search-index-path

Defines where Mokapi stores the search index used by the web dashboard.

By default, Mokapi stores the search index in a temporary directory on disk.
This reduces memory usage and provides predictable memory behavior.

If a file system path is specified, the index will be stored at that location during Mokapi's runtime.
The index is rebuilt at startup and does not block application startup.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-search-index-path | MOKAPI_API_SEARCH_INDEX_PATH | api.search.indexPath | string | - |

```bash tab=CLI
--api-search-index-path /var/lib/mokapi/search
```
```bash tab=Env
MOKAPI_API_SEARCH_INDEX_PATH=/var/lib/mokapi/search
```
```yaml tab=File
api:
  search:
    indexPath: /var/lib/mokapi/search
```

## <a id="api-search-in-memory" name="api-search-in-memory"></a>api-search-in-memory

Forces Mokapi to keep the search index entirely in memory instead of using disk storage.

This provides fast indexing and search performance but increases Go heap memory usage.

This option is recommended only for small projects or development environments.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --api-search-in-memory | MOKAPI_API_SEARCH_IN_MEMORY | api.search.inMemory | bool | false |

```bash tab=CLI
--api-search-in-memory
```
```bash tab=Env
MOKAPI_API_SEARCH_IN_MEMORY=true
```
```yaml tab=File
api:
  search:
    inMemory: true
```

## <a id="mcp-server-enabled" name="mcp-server-enabled"></a>mcp-server-enabled

Enables the MCP (Model Context Protocol) server.
When enabled, Mokapi exposes an MCP-compatible endpoint that allows external tools and AI clients to interact with mocked APIs, events, and configuration through a standardized protocol.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --mcp-server-enabled | MOKAPI_MCP_SERVER_ENABLED | mcp.server.enabled | bool | false |

```bash tab=CLI
--mcp-server-enabled
```
```bash tab=Env
MOKAPI_MCP_SERVER_ENABLED=true
```
```yaml tab=File
mcp:
  server:
    enabled: true
```

## <a id="mcp-server-port" name="mcp-server-port"></a>mcp-server-port

Specifies the TCP port on which the MCP server listens.
The MCP server provides access to Mokapi resources via the Model Context Protocol and can be consumed by compatible clients and tools.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --mcp-server-port | MOKAPI_MCP_SERVER_PORT | mcp.server.port | int | 8080 |

```bash tab=CLI
--mcp-server-port 5000
```
```bash tab=Env
MOKAPI_MCP_SERVER_PORT=5000
```
```yaml tab=File
mcp:
  server:
    port: 5000
```

## <a id="mcp-server-path" name="mcp-server-path"></a>mcp-server-path

Defines the HTTP path under which the MCP server is exposed.
This is useful when Mokapi is running behind a reverse proxy or when the MCP endpoint should be served under a specific path.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --mcp-server-path | MOKAPI_MCP_SERVER_PATH | mcp.server.path | string | /mcp |

```bash tab=CLI
--mcp-server-path /foo/mcp
```
```bash tab=Env
MOKAPI_MCP_SERVER_PATH=/foo/mcp
```
```yaml tab=File
mcp:
  server:
    path: /foo/mcp
```

## <a id="health-enabled" name="health-enabled"></a>health-enabled

Enables or disables the health check endpoint entirely.
When set to false, Mokapi will not expose any health endpoint.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --health-enabled | MOKAPI_HEALTH_ENABLED | health.enabled | bool | true |

```bash tab=CLI
--health-enabled false
```
```bash tab=Env
MOKAPI_HEALTH_ENABLED=false
```
```yaml tab=File
health:
  enabled: false
```

## <a id="health-port" name="health-port"></a>health-port

The port on which the health endpoint is exposed.
- If the value matches the dashboard/API port, the health endpoint is served by the same HTTP server.
- If a different port is specified, Mokapi starts a separate HTTP listener dedicated to health checks.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --health-port | MOKAPI_HEALTH_PORT | health.port | int | 8080 |

```bash tab=CLI
--health-port 8081
```
```bash tab=Env
MOKAPI_HEALTH_PORT=8081
```
```yaml tab=File
health:
  port: 8081
```

## <a id="health-path" name="health-path"></a>health-path

The HTTP path for the health endpoint.
- If empty, the default path /health is used.
- The value must be an absolute path (starting with /).

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --health-path | MOKAPI_HEALTH_PATH | health.path | string | /health |

```bash tab=CLI
--health-path /health/live
```
```bash tab=Env
MOKAPI_HEALTH_PATH=/health/live
```
```yaml tab=File
health:
  path: /health/live
```

## <a id="health-log" name="health-log"></a>health-log

Controls whether HTTP requests to the health endpoint are logged.
By default, health check requests are not logged to avoid excessive log noise
from load balancers, uptime monitors, and orchestration systems.

Enable this option when debugging health check behavior.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --health-log | MOKAPI_HEALTH_LOG | health.log | bool | false |

```bash tab=CLI
--health-log
```
```bash tab=Env
MOKAPI_HEALTH_LOG=true
```
```yaml tab=File
health:
  log: true
```

## <a id="providers-file" name="providers-file"></a>providers-file

Enables the file provider using a shorthand configuration. This option is useful for quick setups without defining a full configuration block.
When set, Mokapi loads dynamic configuration files from the specified path. You can further control how files are loaded using additional providers-file-* flags such as include rules, directories, and skip prefixes.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file | MOKAPI_PROVIDERS_FILE | providers.file | string | - |

```bash tab=CLI
--providers-file directory=/foo,fileSkipPrefix=skip
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE=directory=/foo,fileSkipPrefix=skip
```
```yaml tab=File
providers:
  file:
    directory: /foo
    fileSkipPrefix: [skip]
```

## <a id="providers-file-filename" name="providers-file-filename"></a>providers-file-filename

Specifies a single configuration file to load using the file provider.
This option can be used multiple times to load additional files.
To pass several files in one argument, use --providers-file-filenames instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-filename | MOKAPI_PROVIDERS_FILE_FILENAME | providers.file.filenames | list | - |

```bash tab=CLI
--providers-file-filename foobar.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_FILENAME=foobar.yaml
```
```yaml tab=File
providers:
  file:
    filenames: [foobar.yaml]
```

## <a id="providers-file-filenames" name="providers-file-filenames"></a>providers-file-filenames

Specifies multiple configuration files to load using the file provider.
This option is equivalent to using --providers-file-filename multiple times, but allows defining all files in a single argument.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-filenames | MOKAPI_PROVIDERS_FILE_FILENAMES | providers.file.filenames | list | - |

```bash tab=CLI
--providers-file-filenames foo.yaml bar.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_FILENAMES=foo.yaml bar.yaml
```
```yaml tab=File
providers:
  file:
    filenames: [foo.yaml, bar.yaml]
```

## <a id="providers-file-directory" name="providers-file-directory"></a>providers-file-directory

Specifies a directory from which configuration files are loaded.
All supported configuration files in the directory are processed. The directory is watched for changes, allowing dynamic reloading when files are added, modified, or removed.
To pass several directories in one argument, use --providers-file-directories instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-directory | MOKAPI_PROVIDERS_FILE_DIRECTORY | providers.file.directories | list | - |

```bash tab=CLI
--providers-file-directory ./configs
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORY=./configs
```
```yaml tab=File
providers:
  file:
    directories: [./configs]
```

## <a id="providers-file-directories" name="providers-file-directories"></a>providers-file-directories

Specifies multiple directories from which configuration files are loaded.
All supported configuration files in each directory are processed. Directories are watched for changes, allowing dynamic reloading across all configured paths.
To add a single directory per flag, use --providers-file-directory instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-directories | MOKAPI_PROVIDERS_FILE_DIRECTORIES | providers.file.directories | list | - |

```bash tab=CLI
--providers-file-directories ./configs ./data
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORIES=./configs ./data
```
```yaml tab=File
providers:
  file:
    directories: [./configs, ./data]
```

### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="providers-file-directories-index" name="providers-file-directories-index"></a>`--providers-file-directories[<index>]` | providers.file.directories[] | string | Configure the directory at the specified index |
| <a id="providers-file-directories-index-path" name="providers-file-directories-index-path"></a>`--providers-file-directories[<index>]-path` | providers.file.directories[].path | string | Set the directory path |
| <a id="providers-file-directories-index-include" name="providers-file-directories-index-include"></a>`--providers-file-directories[<index>]-include` | providers.file.directories[].include | list | Include only matching files or patterns |
| <a id="providers-file-directories-index-exclude" name="providers-file-directories-index-exclude"></a>`--providers-file-directories[<index>]-exclude` | providers.file.directories[].exclude | list | Exclude matching files or patterns |

#### `--providers-file-directories[<index>]`

Configures a directory entry at the specified index in the directories list.
This option allows directory-specific configuration and is mainly intended for advanced or programmatic use cases where precise control over individual directories is required.

```bash tab=CLI
--providers-file-directories[0] ./mocks
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORIES[0]=./mocks
```

#### `--providers-file-directories[<index>]-path`

Specifies the filesystem path of the directory to load configuration files from.
All supported configuration files in this directory are processed, subject to include and exclude rules.


```bash tab=CLI
--providers-file-directories[1]-path ./mocks
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORIES[1]_PATH=./mocks
```

#### `--providers-file-directories[<index>]-include`

Defines include patterns that files in the directory must match to be loaded.
If at least one include pattern is specified, only files matching one of the patterns are processed. When empty, all supported files are included by default.


```bash tab=CLI
--providers-file-directories[1]-include *index.js
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORIES[1]_INCLUDE=*index.js
```

#### `--providers-file-directories[<index>]-exclude`

Defines exclude patterns for files or directories within the specified directory.
Any file or directory matching one of the exclude patterns is ignored, even if it matches an include rule.


```bash tab=CLI
--providers-file-directories[0]-exclude *.tmp
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORIES[0]_EXCLUDE=*.tmp
```

## <a id="providers-file-skip-prefix" name="providers-file-skip-prefix"></a>providers-file-skip-prefix

Defines prefixes that cause files or directories to be ignored by the file provider.
Any file or directory whose name starts with one of the configured prefixes is skipped. This is useful for ignoring temporary files, backups, or disabled configurations.
By default, files and directories starting with "_" are skipped.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-skip-prefix | MOKAPI_PROVIDERS_FILE_SKIP_PREFIX | providers.file.skipPrefix | list | _ |

```bash tab=CLI
--providers-file-skip-prefix skip_
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_SKIP_PREFIX=skip_
```
```yaml tab=File
providers:
  file:
    skipPrefix: [skip_]
```

## <a id="providers-file-include" name="providers-file-include"></a>providers-file-include

Defines include patterns that configuration files must match to be loaded.
If at least one include pattern is specified, only files matching one of the patterns are processed. When empty, all supported files are included by default.
Patterns typically follow glob-style matching.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-include | MOKAPI_PROVIDERS_FILE_INCLUDE | providers.file.include | list | - |

```bash tab=CLI
--providers-file-include *.json *.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_INCLUDE=*.json *.yaml
```
```yaml tab=File
providers:
  file:
    include: ['*.json', '*.yaml']
```

### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="providers-file-include-index" name="providers-file-include-index"></a>`--providers-file-include[<index>]` | providers.file.include[] | string | Set include rule at the specified index |

#### `--providers-file-include[<index>]`

Sets or overrides an include rule at a specific index in the include list.
This option is mainly intended for advanced use cases or programmatic configuration where precise control over individual include rules is required.

```bash tab=CLI
--providers-file-include[1] *.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_INCLUDE[1]=*.yaml
```

## <a id="providers-file-exclude" name="providers-file-exclude"></a>providers-file-exclude

Defines patterns for files or directories that should be excluded when loading configuration.
Any file or directory matching one of the exclude patterns is ignored, even if it would otherwise be included. When empty, no files are explicitly excluded.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-file-exclude | MOKAPI_PROVIDERS_FILE_EXCLUDE | providers.file.exclude | list | - |

```bash tab=CLI
--providers-file-exclude tmp/*
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_EXCLUDE=tmp/*
```
```yaml tab=File
providers:
  file:
    exclude: ['tmp/*']
```

### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="providers-file-exclude-index" name="providers-file-exclude-index"></a>`--providers-file-exclude[<index>]` | providers.file.exclude[] | string | Set exclude rule at the specified index |

#### `--providers-file-exclude[<index>]`

Sets or overrides an exclude rule at the specified index.
This option is useful when exclude rules are defined via environment variables or configuration files but need to be adjusted or overridden using CLI arguments.

```bash tab=CLI
--providers-file-exclude[1] tmp/*
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_EXCLUDE[1]=tmp/*
```

## <a id="providers-http" name="providers-http"></a>providers-http

Enables the HTTP provider using a shorthand configuration.
When enabled, Mokapi fetches dynamic configuration from one or more HTTP endpoints. The provider periodically polls the configured URLs and applies updates automatically when the remote configuration changes.
Additional flags allow you to control polling behavior, timeouts, proxy settings, and TLS verification.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http | MOKAPI_PROVIDERS_HTTP | providers.http | string | - |

```bash tab=CLI
--providers-http url=https://foo.bar/file.yaml,proxy=https://proxy.example.com
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP=url=https://foo.bar/file.yaml,proxy=https://proxy.example.com
```
```yaml tab=File
providers:
  http:
    urls: [https://foo.bar/file.yaml]
    proxy: https://proxy.example.com
```

## <a id="providers-http-url" name="providers-http-url"></a>providers-http-url

Specifies a single HTTP endpoint from which configuration is fetched.
This option can be used multiple times to define additional endpoints. Each endpoint is polled at the configured interval, and changes are applied dynamically.
To pass several endpoints in one argument, use --providers-http-urls instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-url | MOKAPI_PROVIDERS_HTTP_URL | providers.http.urls | list | - |

```bash tab=CLI
--providers-http-url https://foo.bar/file.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URL=https://foo.bar/file.yaml
```
```yaml tab=File
providers:
  http:
    urls: [https://foo.bar/file.yaml]
```

## <a id="providers-http-urls" name="providers-http-urls"></a>providers-http-urls

Specifies multiple HTTP endpoints from which configuration is fetched.
This option is equivalent to using --providers-http-url multiple times, but allows defining all endpoints in a single argument or configuration block.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-urls | MOKAPI_PROVIDERS_HTTP_URLS | providers.http.urls | list | - |

```bash tab=CLI
--providers-http-urls https://foo.bar/file1.yaml https://foo.bar/file2.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_URLS=https://foo.bar/file1.yaml https://foo.bar/file2.yaml
```
```yaml tab=File
providers:
  http:
    urls: [https://foo.bar/file1.yaml, https://foo.bar/file2.yaml]
```

## <a id="providers-http-poll-interval" name="providers-http-poll-interval"></a>providers-http-poll-interval

Defines how often the configured HTTP endpoints are polled for changes.
The value must be a valid duration string, such as "30s", "1m", or "5m". Shorter intervals result in faster updates but may increase network usage.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-poll-interval | MOKAPI_PROVIDERS_HTTP_POLL_INTERVAL | providers.http.pollInterval | string | 3m |

```bash tab=CLI
--providers-http-poll-interval 10s
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_POLL_INTERVAL=10s
```
```yaml tab=File
providers:
  http:
    pollInterval: 10s
```

## <a id="providers-http-poll-timeout" name="providers-http-poll-timeout"></a>providers-http-poll-timeout

Sets the maximum duration allowed for a single HTTP polling request.
If the request does not complete within this time, it is aborted and retried at the next polling interval.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-poll-timeout | MOKAPI_PROVIDERS_HTTP_POLL_TIMEOUT | providers.http.pollTimeout | string | 5s |

```bash tab=CLI
--providers-http-poll-timeout 10s
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_POLL_TIMEOUT=10s
```
```yaml tab=File
providers:
  http:
    pollTimeout: 10s
```

## <a id="providers-http-proxy" name="providers-http-proxy"></a>providers-http-proxy

Configures an HTTP proxy used for all requests made by the HTTP provider.
This is useful in environments where outbound traffic must go through a proxy server.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-proxy | MOKAPI_PROVIDERS_HTTP_PROXY | providers.http.proxy | string | - |

```bash tab=CLI
--providers-http-proxy http://localhost:3128
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_PROXY=http://localhost:3128
```
```yaml tab=File
providers:
  http:
    proxy: http://localhost:3128
```

## <a id="providers-http-tls-skip-verify" name="providers-http-tls-skip-verify"></a>providers-http-tls-skip-verify

Disables TLS certificate verification for HTTPS endpoints.
This option can be useful in development environments or in enterprise setups where HTTPS traffic is intercepted by a proxy using a custom or internal certificate.
However, skipping certificate verification is insecure and should be avoided when possible. A safer alternative is to configure a custom certificate authority using providers-http-ca or to install the proxy certificate into the system certification pool.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-tls-skip-verify | MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY | providers.http.tlsSkipVerify | bool | false |

```bash tab=CLI
--providers-http-tls-skip-verify true
--providers-http-tls-skip-verify
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_TLS_SKIP_VERIFY=true
```
```yaml tab=File
providers:
  http:
    tlsSkipVerify: true
```

## <a id="providers-http-ca" name="providers-http-ca"></a>providers-http-ca

Specifies a custom certificate authority (CA) file used to verify TLS connections.
The value can be a file path or the certificate content itself. PEM (including bundles with several certificates) and DER encoded certificates are accepted.
The CA is added to the system certificate pool, so certificates signed by well-known authorities are still trusted. This is useful when working with internal or self-signed certificates.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-http-ca | MOKAPI_PROVIDERS_HTTP_CA | providers.http.ca | string | - |

```bash tab=CLI
--providers-http-ca /path/to/mycert.pem
```
```bash tab=Env
MOKAPI_PROVIDERS_HTTP_CA=/path/to/mycert.pem
```
```yaml tab=File
providers:
  http:
    ca: /path/to/mycert.pem
```

## <a id="providers-git" name="providers-git"></a>providers-git

Enables the Git provider using a shorthand configuration.
When enabled, Mokapi clones one or more Git repositories and loads configuration files from them. Repositories are periodically pulled to detect changes and apply updates dynamically.
Additional flags allow you to control polling behavior, repository selection, authentication, and file filtering.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git | MOKAPI_PROVIDERS_GIT | providers.git | string | - |

```bash tab=CLI
--providers-git pullInterval=10s,tempDir=/tempdir
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT=pullInterval=10s,tempDir=/tempdir
```
```yaml tab=File
providers:
  git:
    pullInterval: 10s
    tempDir: /tempdir
```

## <a id="providers-git-url" name="providers-git-url"></a>providers-git-url

Specifies a single Git repository from which configuration is cloned.
This option can be used multiple times to define additional repositories. Each repository is cloned and periodically pulled to fetch updates.
To pass several repository URLs in one argument, use --providers-git-urls instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git-url | MOKAPI_PROVIDERS_GIT_URL | providers.git.urls | list | - |

```bash tab=CLI
--providers-git-url https://github.com/foo/foo.git
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_URL=https://github.com/foo/foo.git
```
```yaml tab=File
providers:
  git:
    urls: [https://github.com/foo/foo.git]
```

## <a id="providers-git-urls" name="providers-git-urls"></a>providers-git-urls

Specifies multiple Git repositories from which configuration is cloned.
This option is equivalent to using --providers-git-url multiple times, but allows defining all repositories in a single argument or configuration block.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git-urls | MOKAPI_PROVIDERS_GIT_URLS | providers.git.urls | list | - |

```bash tab=CLI
--providers-git-urls https://github.com/foo/foo.git https://github.com/bar/bar.git
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_URLS=https://github.com/foo/foo.git https://github.com/bar/bar.git
```
```yaml tab=File
providers:
  git:
    urls: [https://github.com/foo/foo.git, https://github.com/bar/bar.git]
```

## <a id="providers-git-pull-interval" name="providers-git-pull-interval"></a>providers-git-pull-interval

Defines how often Git repositories are pulled to check for updates.
The value must be a valid duration string, such as "30s", "1m", or "5m". Shorter intervals result in faster updates but may increase network and Git server load.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git-pull-interval | MOKAPI_PROVIDERS_GIT_PULL_INTERVAL | providers.git.pullInterval | string | 3m |

```bash tab=CLI
--providers-git-pull-interval 10s
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_PULL_INTERVAL=10s
```
```yaml tab=File
providers:
  git:
    pullInterval: 10s
```

## <a id="providers-git-temp-dir" name="providers-git-temp-dir"></a>providers-git-temp-dir

Specifies the directory used for cloning and checking out Git repositories.
If not set, Mokapi uses a default temporary directory. Setting this option can be useful for controlling disk usage or for environments with restricted filesystem access.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git-temp-dir | MOKAPI_PROVIDERS_GIT_TEMP_DIR | providers.git.tempDir | string | - |

```bash tab=CLI
--providers-git-temp-dir /tempdir
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_TEMP_DIR=/tempdir
```
```yaml tab=File
providers:
  git:
    tempDir: /tempdir
```

## <a id="providers-git-repository" name="providers-git-repository"></a>providers-git-repository

Configures a single Git repository using a shorthand syntax.
This option allows defining repository-specific settings such as allowed files, authentication, and pull intervals.
To pass several repositories in one argument, use --providers-git-repositories instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git-repository | MOKAPI_PROVIDERS_GIT_REPOSITORY | providers.git.repositories | list | - |

```bash tab=CLI
--providers-git-repository url=https://github.com/foo/foo.git,include=*.json
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORY=url=https://github.com/foo/foo.git,tempDir=/tempdir
```
```yaml tab=File
providers:
  git:
    repositories:
     - url: https://github.com/foo/foo.git
       include: '*.json'
```

## <a id="providers-git-repositories" name="providers-git-repositories"></a>providers-git-repositories

Configures multiple Git repositories using shorthand syntax.
This option is equivalent to using --providers-git-repository multiple times, but allows defining all repositories in a single argument or configuration block.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-git-repositories | MOKAPI_PROVIDERS_GIT_REPOSITORIES | providers.git.repositories | list | - |

```bash tab=CLI
--providers-git-repositories url=https://github.com/foo/foo.git,include=*.json url=https://github.com/bar/bar.git,include=*.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES=url=https://github.com/foo/foo.git,include=*.json url=https://github.com/bar/bar.git,include=*.yaml
```
```yaml tab=File
providers:
  git:
    repositories:
     - url: https://github.com/foo/foo.git
       include: ['*.json']
     - url: https://github.com/bar/bar.git
       include: ['*.yaml']
```

### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="providers-git-repositories-index" name="providers-git-repositories-index"></a>`--providers-git-repositories[<index>]` | providers.git.repositories[] | string | Configure the repository at the specified index using shorthand syntax |
| <a id="providers-git-repositories-index-url" name="providers-git-repositories-index-url"></a>`--providers-git-repositories[<index>]-url` | providers.git.repositories[].url | string | Set the repository URL |
| <a id="providers-git-repositories-index-file" name="providers-git-repositories-index-file"></a>`--providers-git-repositories[<index>]-file` | providers.git.repositories[].files | list | Allow a single file from the repository (repeatable) |
| <a id="providers-git-repositories-index-files" name="providers-git-repositories-index-files"></a>`--providers-git-repositories[<index>]-files` | providers.git.repositories[].files | list | Allow only specific files from the repository |
| <a id="providers-git-repositories-index-include" name="providers-git-repositories-index-include"></a>`--providers-git-repositories[<index>]-include` | providers.git.repositories[].include | list | Include only matching files or patterns |
| <a id="providers-git-repositories-index-auth-github" name="providers-git-repositories-index-auth-github"></a>`--providers-git-repositories[<index>]-auth-github` | providers.git.repositories[].auth.gitHub | string | Authenticate using GitHub credentials |
| <a id="providers-git-repositories-index-auth-github-appid" name="providers-git-repositories-index-auth-github-appid"></a>`--providers-git-repositories[<index>]-auth-github-appid` | providers.git.repositories[].auth.gitHub.appId | string | GitHub App ID used for authentication |
| <a id="providers-git-repositories-index-auth-github-installationid" name="providers-git-repositories-index-auth-github-installationid"></a>`--providers-git-repositories[<index>]-auth-github-installationid` | providers.git.repositories[].auth.gitHub.installationId | string | GitHub App installation ID |
| <a id="providers-git-repositories-index-auth-github-privatekey" name="providers-git-repositories-index-auth-github-privatekey"></a>`--providers-git-repositories[<index>]-auth-github-privatekey` | providers.git.repositories[].auth.gitHub.privateKey | string | Private key for GitHub App authentication |
| <a id="providers-git-repositories-index-pull-interval" name="providers-git-repositories-index-pull-interval"></a>`--providers-git-repositories[<index>]-pull-interval` | providers.git.repositories[].pullInterval | string | Override pull interval for this repository |

#### `--providers-git-repositories[<index>]`

Configures a Git repository at the specified index in the repositories list.
This option is mainly intended for advanced or programmatic configurations where precise control over individual repositories is required.

```bash tab=CLI
--providers-git-repositories[0] url=https://github.com/foo/foo.git,include=*.json
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]=url=https://github.com/foo/foo.git,include=*.json
```
```yaml tab=File
providers:
  git:
    repositories:
     - url: https://github.com/foo/foo.git
       include: ['*.json']
```

#### `--providers-git-repositories[<index>]-url`

Specifies the Git repository URL.
The repository is cloned and used as a source for configuration files.

```bash tab=CLI
--providers-git-repositories[0]-url https://github.com/foo/foo.git
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_URL=https://github.com/foo/foo.git
```
```yaml tab=File
providers:
  git:
    repositories:
     - url: https://github.com/foo/foo.git
```

#### `--providers-git-repositories[<index>]-file`

Restricts configuration loading to specific files within the repository.
Only the specified files are considered. This option can be used multiple times to allow additional files.
To pass several files in one argument, use --providers-git-repositories[<index>]-files instead.

```bash tab=CLI
--providers-git-repositories[0]-file mokapi/api.json
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_FILE=mokapi/api.json
```
```yaml tab=File
providers:
  git:
    repositories:
     - files: [mokapi/api.json]
```

#### `--providers-git-repositories[<index>]-files`

Specifies multiple files within the repository that are allowed to be loaded.
This option is equivalent to using providers-git-repositories[<index>]-file multiple times.

```bash tab=CLI
--providers-git-repositories[0]-files mokapi/api.json mokapi/handler.js
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_FILES=mokapi/api.json mokapi/handler.js
```
```yaml tab=File
providers:
  git:
    repositories:
     - files: [mokapi/api.json, mokapi/handler.js]
```

#### `--providers-git-repositories[<index>]-include`

Defines include patterns that files must match to be loaded from the repository.
If at least one include pattern is specified, only files matching one of the patterns are processed. When empty, all supported files are included by default.

```bash tab=CLI
--providers-git-repositories[0]-include mokapi/**/*.json
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_INCLUDE=mokapi/**/*.json
```
```yaml tab=File
providers:
  git:
    repositories:
     - include: ['mokapi/**/*.json']
```

#### `--providers-git-repositories[<index>]-auth-github`

Enables authentication using GitHub credentials for the repository.
This option allows accessing private repositories hosted on GitHub by using credentials provided via environment variables or the GitHub CLI configuration.

```bash tab=CLI
--providers-git-repositories[0]-auth-github appId=12345,installationId=123456789,privateKey=2024-2-25.private-key.pem
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_AUTH_GITHUB=appId=12345,installationId=123456789,privateKey=2024-2-25.private-key.pem
```
```yaml tab=File
providers:
  git:
    repositories:
     - auth:
         gitHub:
           appId: 12345
           installationId: 12345
           privateKey: 2024-2-25.private-key.pem
```

#### `--providers-git-repositories[<index>]-auth-github-appid`

Sets or overrides the GitHub App ID used to authenticate access to the repository at the specified index.

```bash tab=CLI
--providers-git-repositories[0]-auth-github-appId 12345
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_AUTH_GITHUB_APPID=12345
```
```yaml tab=File
providers:
  git:
    repositories:
     - auth:
         gitHub:
           appId: 12345
```

#### `--providers-git-repositories[<index>]-auth-github-installationid`

Sets or overrides the GitHub App installation ID used to authenticate access to the repository at the specified index.

```bash tab=CLI
--providers-git-repositories[0]-auth-github-installationid 123456789
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_AUTH_GITHUB_INSTALLATIONID=123456789
```
```yaml tab=File
providers:
  git:
    repositories:
     - auth:
         gitHub:
           installationId: 12345

```

#### `--providers-git-repositories[<index>]-auth-github-privatekey`

Sets or overrides the private key used for GitHub App authentication for the repository at the specified index.

```bash tab=CLI
--providers-git-repositories[0]-auth-github-privatekey 2024-2-25.private-key.pem
```
```bash tab=Env
MOKAPI_PROVIDERS_GIT_REPOSITORIES[0]_AUTH_GITHUB_PRIVATEKEY=2024-2-25.private-key.pem
```
```yaml tab=File
providers:
  git:
    repositories:
     - auth:
         gitHub:
           privateKey: 2024-2-25.private-key.pem
```

#### `--providers-git-repositories[<index>]-pull-interval`

Defines how often this repository is pulled. It overrides providers-git-pull-interval for this repository only.
The value must be a valid duration string, such as "30s" or "5m".

## <a id="providers-npm" name="providers-npm"></a>providers-npm

Enables the npm provider using a shorthand configuration.
When enabled, Mokapi loads configuration files from npm packages or from globally installed npm folders. This allows distributing configuration alongside npm packages and reusing it across projects.
Additional flags allow you to control which packages or files are loaded.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-npm | MOKAPI_PROVIDERS_NPM | providers.npm | string | - |

```bash tab=CLI
--providers-npm package={"name": "foo-api"},globalFolders=/npm
--providers-npm package={\"name\": \"foo-api\"},globalFolders=/npm # Windows

```
```bash tab=Env
MOKAPI_PROVIDERS_NPM=package={"name": "foo-api"},globalFolders=/npm
```
```yaml tab=File
providers:
  npm:
    packages:
     - name: foo-api
    globalFolders: [/npm]
```

## <a id="providers-npm-global-folder" name="providers-npm-global-folder"></a>providers-npm-global-folder

Adds additional folders where Mokapi looks for npm packages.
By default, Mokapi resolves npm packages by searching the current working directory and its parent directories, following the same resolution strategy as npm. This option allows you to extend that search path with one or more additional global npm folders.
To pass several folders in one argument, use --providers-npm-global-folders instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-npm-global-folder | MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDER | providers.npm.globalFolders | list | - |

```bash tab=CLI
--providers-npm-global-folder /npm
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDER=/npm
```
```yaml tab=File
providers:
  npm:
    globalFolders: [/npm]
```

## <a id="providers-npm-global-folders" name="providers-npm-global-folders"></a>providers-npm-global-folders

Specifies multiple global npm folders from which configuration files are loaded.
This option is equivalent to using --providers-npm-global-folder multiple times.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-npm-global-folders | MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDERS | providers.npm.globalFolders | list | - |

```bash tab=CLI
--providers-npm-global-folders /npm /npm2
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDERS=/npm /npm2
```
```yaml tab=File
providers:
  npm:
    globalFolders: [/npm, /npm2]
```

## <a id="providers-npm-package" name="providers-npm-package"></a>providers-npm-package

Configures a single npm package as a source for configuration files.
The package must be available locally or installed globally. Configuration files are loaded from the package contents.
To pass several packages in one argument, use --providers-npm-packages instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-npm-package | MOKAPI_PROVIDERS_NPM_PACKAGE | providers.npm.packages | list | - |

```bash tab=CLI
--providers-npm-package name=foo-api,file=api.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGE=name=foo-api,file=api.yaml
```
```yaml tab=File
providers:
  npm:
    packages:
     - name: foo-api
       files: [api.yaml]
```

## <a id="providers-npm-packages" name="providers-npm-packages"></a>providers-npm-packages

Configures multiple npm packages as sources for configuration files.
This option is equivalent to using --providers-npm-package multiple times.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --providers-npm-packages | MOKAPI_PROVIDERS_NPM_PACKAGES | providers.npm.packages | list | - |

```bash tab=CLI
--providers-npm-packages name=foo-api name=bar-api
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGES=name=foo-api name=bar-api
```
```yaml tab=File
providers:
  npm:
    packages:
     - name: foo-api
       files: [api.yaml]
```

### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="providers-npm-packages-index" name="providers-npm-packages-index"></a>`--providers-npm-packages[<index>]` | providers.npm.packages[] | string | Configure the package at the specified index using shorthand syntax |
| <a id="providers-npm-packages-index-name" name="providers-npm-packages-index-name"></a>`--providers-npm-packages[<index>]-name` | providers.npm.packages[].name | string | Set the name of the npm package |
| <a id="providers-npm-packages-index-file" name="providers-npm-packages-index-file"></a>`--providers-npm-packages[<index>]-file` | providers.npm.packages[].files | list | Allow a single file from the package (repeatable) |
| <a id="providers-npm-packages-index-files" name="providers-npm-packages-index-files"></a>`--providers-npm-packages[<index>]-files` | providers.npm.packages[].files | list | Allow only specific files from the package |
| <a id="providers-npm-packages-index-include" name="providers-npm-packages-index-include"></a>`--providers-npm-packages[<index>]-include` | providers.npm.packages[].include | list | Include only matching files or patterns from the package |

#### `--providers-npm-packages[<index>]`

Configures an npm package at the specified index in the packages list.
This option is useful when packages are defined via environment variables or configuration files, but need to be adjusted or overridden using CLI arguments. It allows precise, index-based modification of existing package definitions.


```bash tab=CLI
--providers-npm-packages[0] file=api.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGES[0]=file=api.yaml
```

#### `--providers-npm-packages[<index>]-name`

Sets or overrides the name of the npm package at the specified index.
This is commonly used to adjust a package definition that was initially provided via environment variables or configuration files.

```bash tab=CLI
--providers-npm-packages[0]-name bar-api
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGES[0]_NAME=bar-api
```

#### `--providers-npm-packages[<index>]-file`

Restricts configuration loading to specific files within the npm package at the specified index.
This option can be used multiple times and is useful for refining or overriding file selection defined via environment variables or configuration files.
To pass several files in one argument, use --providers-npm-packages[<index>]-files instead.

```bash tab=CLI
--providers-npm-packages[0]-file dist/model/openapi/complete-api.json
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGES[0]_FILE=dist/model/openapi/complete-api.json
```

#### `--providers-npm-packages[<index>]-files`

Specifies multiple files within the npm package that are allowed to be loaded.
This option is equivalent to using providers-npm-packages[<index>]-file multiple times and can be used to override existing package configuration.

```bash tab=CLI
--providers-npm-packages[0]-files api.yaml api2.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGES[0]_FILES=api.yaml api2.yaml
```

#### `--providers-npm-packages[<index>]-include`

Defines include patterns that files must match to be loaded from the npm package at the specified index.
This option is useful for narrowing or overriding file selection rules defined via environment variables or configuration files. When empty, all supported files are included by default.

```bash tab=CLI
--providers-npm-packages[0]-include **/api/
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGES[0]_INCLUDE=**/api/
```

## <a id="root-ca-cert" name="root-ca-cert"></a>root-ca-cert

Specifies the root certificate authority (CA) certificate used by Mokapi to sign generated TLS certificates.
This certificate is used when Mokapi dynamically generates certificates for mocked services that require TLS.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --root-ca-cert | MOKAPI_ROOT_CA_CERT | rootCaCert | string | - |

```bash tab=CLI
--root-ca-cert /path/to/caCert.pem
```
```bash tab=Env
MOKAPI_ROOT_CA_CERT=/path/to/caCert.pem
```
```yaml tab=File
rootCaCert: /path/to/caCert.pem
```

## <a id="root-ca-key" name="root-ca-key"></a>root-ca-key

Specifies the private key corresponding to the root CA certificate.
The private key is required to sign generated TLS certificates. It must match the certificate provided via root-ca-cert.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --root-ca-key | MOKAPI_ROOT_CA_KEY | rootCaKey | string | - |

```bash tab=CLI
--root-ca-key /path/to/caKey.pem
```
```bash tab=Env
MOKAPI_ROOT_CA_KEY=/path/to/caKey.pem
```
```yaml tab=File
rootCaKey: /path/to/caKey.pem
```

## <a id="event-store-default-size" name="event-store-default-size"></a>event-store-default-size

Defines the default maximum number of events stored per API.
When the limit is reached, older events are discarded. This helps control memory usage while still allowing recent events to be inspected.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --event-store-default-size | MOKAPI_EVENT_STORE_DEFAULT_SIZE | event.store.default.size | int | 100 |

```bash tab=CLI
--event-store-default-size 500
```
```bash tab=Env
MOKAPI_EVENT_STORE_DEFAULT_SIZE=500
```
```yaml tab=File
event:
  store:
    default:
      size: 500
```

## <a id="event-store" name="event-store"></a>event-store

Configures the event store using a shorthand syntax.
The event store keeps track of events produced by mocked APIs, allowing them to be inspected via the API server or web dashboard.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --event-store | MOKAPI_EVENT_STORE | event.store | string | - |

```bash tab=CLI
--event-store foo={"size":250}
```
```bash tab=Env
MOKAPI_EVENT_STORE=foo={"size":250}
```
```yaml tab=File
event:
  store:
    foo:
      size: 250
```

### Sub-options

| Flag | Config file | Type | Description |
|------|-------------|:----:|-------------|
| <a id="event-store-name-size" name="event-store-name-size"></a>`--event-store-<name>-size` | event.store.&lt;name&gt;.size | int | Sets event store size for a specific API |

#### `--event-store-<name>-size`

Configures the event store.
The event store keeps track of events produced by mocked APIs, allowing them to be inspected via the API server or web dashboard.

```bash tab=CLI
--event-store-foo-size 500
```
```bash tab=Env
MOKAPI_EVENT_STORE_FOO_SIZE=500
```
```yaml tab=File
event:
  store:
    foo:
      size: 500
```

## <a id="data-gen-optional-properties" name="data-gen-optional-properties"></a>data-gen-optional-properties

Controls how often optional properties are included when generating example or mock data.
The value can be specified either as a number between 0 and 1, or as a predefined string:

- always     → 1.0
- often      → 0.85
- sometimes  → 0.5
- rarely     → 0.15
- never      → 0.0

Higher values result in more optional properties being present in generated data.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --data-gen-optional-properties | MOKAPI_DATA_GEN_OPTIONAL_PROPERTIES | data-gen.optionalProperties | string | 0.85 |

```bash tab=CLI
--data-gen-optional-properties 0.9
--data-gen-optional-properties sometimes
```
```bash tab=Env
MOKAPI_DATA_GEN_OPTIONAL_PROPERTIES=0.9
```
```yaml tab=File
data-gen:
  optionalProperties: sometimes
```

## <a id="config" name="config"></a>config

Provides inline configuration data directly via the command line.
This option can be used multiple times to supply additional configuration fragments, which are merged with other configuration sources.
To pass several fragments in one argument, use --configs instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --config | MOKAPI_CONFIG | configs | list | - |

```bash tab=CLI
--config {"openapi": "3.0", "info": { "title": "foo" } }
--config file:/path/to/config.yaml
```

## <a id="configs" name="configs"></a>configs

Provides inline configuration data directly via the command line.
Several fragments are passed in a single argument, separated by spaces.
To add a single fragment per flag, use --config instead.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --configs | MOKAPI_CONFIGS | configs | list | - |

```bash tab=CLI
--configs file:/path/to/config.yaml  file:/path/to/config.yaml
```

## <a id="config-file" name="config-file"></a>config-file

Reads configuration from a file provided via the command line.
The file content is processed as configuration input and merged with other configuration sources such as inline configuration and environment variables.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --config-file | MOKAPI_CONFIG_FILE | - | file | - |

```bash tab=CLI
--config-file=mokapi.yaml
```

### Aliases

- cli-input

## <a id="generate-cli-skeleton" name="generate-cli-skeleton"></a>generate-cli-skeleton

Prints a configuration skeleton and exits without starting Mokapi.
If set without a value, the full skeleton is generated. If a value is provided, only the specified section is generated, e.g. providers.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --generate-cli-skeleton | MOKAPI_GENERATE_CLI_SKELETON | - | string | - |

```bash tab=CLI
--generate-cli-skeleton
--generate-cli-skeleton providers
```

## <a id="help" name="help"></a>help

Show help information and exit

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --help | -h | MOKAPI_HELP | - | bool | false |

## <a id="version" name="version"></a>version

Show version information and exit

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --version | -v | MOKAPI_VERSION | - | bool | false |

## <a id="javascript-modules" name="javascript-modules"></a>javascript.modules

Config file only. Defines module aliases. Before Mokapi resolves an import, it replaces the import path with the target of the matching alias.
A key may contain one `*` wildcard; the matched text replaces the `*` in the target.
See [Module Aliases](/docs/javascript-api/modules.md#module-aliases) for details.

```yaml tab=File (YAML)
javascript:
  modules:
    aliases:
      msw: mokapi-msw-adapter
      "msw/*": mokapi-msw-adapter/*
```
```json tab=File (JSON)
{
  "javascript": {
    "modules": {
      "aliases": {
        "msw": "mokapi-msw-adapter",
        "msw/*": "mokapi-msw-adapter/*"
      }
    }
  }
}
```

## <a id="certificates" name="certificates"></a>certificates

Config file only. Defines static certificates that Mokapi serves for TLS connections instead of generating one with the root CA.
Each entry has a `cert` and a `key`, given either as a file path or as the PEM content. A path starting with `./` is resolved from the directory of the config file.
A certificate is used for its common name and its DNS names. Other host names still get a generated certificate.
The `key` may be omitted if the `cert` file also contains the private key.

```yaml tab=File (YAML)
certificates:
  static:
    - cert: /path/to/cert.pem
      key: /path/to/key.pem
```
```json tab=File (JSON)
{
  "certificates": {
    "static": [
      { "cert": "/path/to/cert.pem", "key": "/path/to/key.pem" }
    ]
  }
}
```

## <a id="sample-data" name="sample-data"></a>sample-data

Generate sample data from a schema

### Description

Generates random example data from a schema definition.

The input can be:
  • A local file path (e.g., ./schema.json)
  • A URL (e.g., https://example.com/schema.yaml)
  • A JSON Schema string passed via stdin or directly as an argument.

Supports JSON Schema, OpenAPI Schema, and Avro Schema formats.

### Usage

```bash tab=Bash
mokapi sample-data [flags] [FILE|URL|SCHEMA]
```

### Example

```bash
# From local file
mokapi sample-data ./schema.json

# From URL
mokapi sample-data https://example.com/schema.yaml

# From stdin (pipe)
cat schema.json | mokapi sample-data

# Direct JSON string
mokapi sample-data '{"type":"object","properties":{"id":{"type":"integer"}}}'
```

<div class="flags">

### Flags

#### General

| Flag | Description |
|------|-------------|
| --[count](#sample-data-count) | Number of samples to generate (default 1) |
| --[input-type](#sample-data-input-type) | Type of the input schema: jsonschema\|openapi\|avro (default: jsonschema) |
| --[output](#sample-data-output) | Output format: json, xml, yaml, binary (or full content type, e.g. application/json) |
| --[help](#sample-data-help) | Show help information and exit |

</div>

### <a id="sample-data-count" name="sample-data-count"></a>count

Defines how many samples are generated. Each sample is generated independently from the schema.

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --count | -n | MOKAPI_COUNT | - | int | 1 |

```bash tab=CLI
mokapi sample-data --count 5 ./schema.json
```

### <a id="sample-data-input-type" name="sample-data-input-type"></a>input-type

Defines the format of the input schema. Supported types are jsonschema, openapi and avro.
Use openapi if the input is an OpenAPI schema object and avro for an Avro schema.

| Flag | Env | Config file | Type | Default |
|------|------|-------------|:----:|:-------:|
| --input-type | MOKAPI_INPUT_TYPE | - | string | jsonschema |

```bash tab=CLI
mokapi sample-data --input-type avro ./schema.avsc
```

### <a id="sample-data-output" name="sample-data-output"></a>output

Defines the format of the generated data. Use one of json, xml, yaml or binary, or a full content type such as application/json.
The default is application/json.

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --output | -o | MOKAPI_OUTPUT | - | string | application/json |

```bash tab=CLI
mokapi sample-data --output yaml ./schema.json
```

### <a id="sample-data-help" name="sample-data-help"></a>help

Show help information and exit

| Flag | Shorthand | Env | Config file | Type | Default |
|------|:---------:|------|-------------|:----:|:-------:|
| --help | -h | MOKAPI_HELP | - | bool | false |

