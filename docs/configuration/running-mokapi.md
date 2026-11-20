---
title: Running Mokapi | Common Commands
description: The commands you use most to start Mokapi with a spec file, a directory, a URL, a Git repository or a script, and how to change the dashboard port and log level.
---
# Running Mokapi

Mokapi starts without arguments, but without dynamic configuration there is nothing to mock.
Pass the files, directories or URLs that contain your OpenAPI or AsyncAPI specs and scripts as
positional arguments at the end of the command. Everything else is a [flag](/docs/configuration/static/cli-flags).

```bash
mokapi [flags] [CONFIG-URL|DIRECTORY|FILE]...
```

## Mock from a single file

```bash
mokapi petstore.yaml
```

Starts Mokapi with the specification in `petstore.yaml`.

## Mock from several files

```bash
mokapi petstore.yaml orders.yaml
```

You can pass as many files, directories and URLs as you need.

## Mock from a directory

```bash
mokapi ./mocks
```

Reads all files in the directory and watches it for changes. Use `--providers-file-include` and
`--providers-file-exclude` to filter files, see the [file provider](/docs/configuration/dynamic/file.md).

## Mock from a URL

```bash
mokapi https://petstore3.swagger.io/api/v3/openapi.json
```

Downloads the specification and polls it for changes, see the [HTTP provider](/docs/configuration/dynamic/http.md).

## Mock from a Git repository

```bash
mokapi git+https://github.com/foo/bar.git
```

Clones the repository and pulls it regularly, see the [Git provider](/docs/configuration/dynamic/git.md).

## Specification and script

```bash
mokapi petstore.yaml petstore.js
```

Mokapi treats a script like any other dynamic configuration, so you pass it next to the
specification it belongs to.

## Change the dashboard port

```bash
mokapi --api-port 9090 petstore.yaml
```

The dashboard and the REST API of Mokapi listen on port 8080 by default. See [`--api-port`](/docs/configuration/static/cli-flags#api-port).

## Show debug logs

```bash
mokapi --log-level debug petstore.yaml
```

Sets the log level, see [`--log-level`](/docs/configuration/static/cli-flags#log-level) for the available values.

## Positional arguments

Mokapi decides by the scheme of an argument which provider reads it:

| Argument                      | Provider                                  |
|-------------------------------|-------------------------------------------|
| `http://` or `https://`       | HTTP, same as `--providers-http-url`      |
| `git+https://`                | Git, same as `--providers-git-url`        |
| no scheme (file or directory) | File, same as `--providers-file-filename` |

Positional arguments must come after all flags. A flag that takes a single value or a boolean
flag can be followed directly by them. A list flag, such as `--providers-file-include`, takes all
following values, so separate it from the positional arguments with `--`:

```bash
mokapi --log-level debug petstore.yaml
mokapi --api-dashboard petstore.yaml
mokapi --providers-file-include '*.json' '*.yaml' -- petstore.yaml
```

See [CLI Syntax](/docs/configuration/static/cli-usage.md) for the equals form and other ways to write list values.

## Next steps

- [Overview](/docs/configuration/overview.md): where options come from and which one wins
- [CLI Syntax](/docs/configuration/static/cli-usage.md): how to write lists, booleans and JSON values
- [CLI Flags](/docs/configuration/static/cli-flags): every flag, environment variable and config file option
- [Docker and Docker Compose](/docs/configuration/use-cases/docker.md): run the official image with volumes, ports and environment variables
