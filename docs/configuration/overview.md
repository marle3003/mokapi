---
title: Introduction to Mokapi Configuration | Static & Dynamic Setup Explained
description: Discover how to configure Mokapi using static files or dynamic updates. Learn startup options, hot-reloading, and flexible setup for your mocked APIs.
---
# Configuration

Mokapi has two types of configuration:
- The startup configuration, referred to as the static configuration
- The dynamic configuration (e.g. an OpenAPI specification)

The static configuration sets up the providers, which read the dynamic
configuration. The dynamic configuration contains the mocked services,
like a REST API or a Kafka topic. It can change without restarting Mokapi.

## Static Configuration

Elements in the static configuration don't often change and changes require a restart of the application.

There are three ways to define static configuration options:
1. Command-line flags
2. Environment variables
3. Configuration file

### Naming

Every option has the same name in all three sources, only the notation differs.
The [CLI Flags](/docs/configuration/static/cli-flags) page lists all options with
their flag, environment variable and configuration file name.

| CLI                                  | Env                                       | Config file                         |
|--------------------------------------|-------------------------------------------|-------------------------------------|
| `--providers-file-directory ./mocks` | `MOKAPI_PROVIDERS_FILE_DIRECTORY=./mocks` | `providers.file.directory: ./mocks` |

The environment variable name is `MOKAPI_` followed by the flag name in upper case,
with each `-` replaced by `_`. In the configuration file, the options are nested YAML
(or JSON) keys:

```yaml
providers:
  file:
    directory: ./mocks
```

### Precedence

If the same option is set in more than one place, the highest source wins:
command-line flag, then environment variable, then configuration file.

```yaml tab=File (mokapi.yaml)
api:
  port: 8081
```
```bash tab=Env
export MOKAPI_API_PORT=8082
```
```bash tab=CLI
mokapi --api-port 8083
```

With all three set, the dashboard listens on port 8083. Without the flag it listens on
8082, and without the environment variable on 8081.

A list option is replaced as a whole, not merged. If `mokapi.yaml` sets
`providers.file.filenames` and you start Mokapi with
`--providers-file-filename b.yaml`, only `b.yaml` is used. Repeating a list
flag on the command line collects all values.

### Configuration File

Mokapi looks for a configuration file named `mokapi` in these locations and
uses the first file it finds:

1. ./mokapi.yaml
2. ./mokapi.yml
3. ./mokapi.json
4. /etc/mokapi/mokapi.yaml
5. /etc/mokapi/mokapi.yml
6. /etc/mokapi/mokapi.json

You can override this:

```bash tab=CLI
mokapi --config-file=/foo/mokapi.yaml
```
```bash tab=Env
MOKAPI_CONFIG_FILE=/foo/mokapi.yaml
```

To start a configuration file, let Mokapi print all options with their default
values and remove what you don't need:

```bash
mokapi --generate-cli-skeleton > mokapi.yaml
```

The flag takes an optional section, for example `--generate-cli-skeleton providers`
prints only the provider options.

### Next Steps

- [Running Mokapi](/docs/configuration/running-mokapi.md) shows the common commands.
- [CLI Syntax](/docs/configuration/static/cli-usage.md) explains how to write lists, booleans and JSON values.
- [CLI Flags](/docs/configuration/static/cli-flags) lists every option.

## Dynamic Configuration

Mokapi reads dynamic configuration and script using [providers](/docs/configuration/dynamic/overview.md).

