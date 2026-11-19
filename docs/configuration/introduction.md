---
title: Introduction to Mokapi Configuration | Static & Dynamic Setup Explained
description: Discover how to configure Mokapi using static files or dynamic updates. Learn startup options, hot-reloading, and flexible setup for your mocked APIs.
---
# Configuration

Mokapi has two types of configuration:
- The startup configuration (referred as the static configuration)
- The dynamic configuration (e.g. OpenApi configuration)

Static configuration sets up the providers which reads dynamic 
configurations. Dynamic configurations contains mock service like 
a REST API or a Kafka Topic. This configuration can change without 
restarting of Mokapi

## Static Configuration

Elements in the static configuration don't often change and changes require a restart of the application.

There are three ways to define static configuration options:
1. Configuration file
2. Command-line arguments
3. Environment variables

If the same option is set in more than one place, the highest source wins:
command-line argument, then environment variable, then configuration file.

```bash
MOKAPI_LOG_LEVEL=info mokapi --log-level debug
```

This example runs with log level `debug`.

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

### CLI Arguments

A list of available arguments can be found [here](/docs/configuration/reference.md)

### Environment Variables

A list of available environment variables can be found [here](/docs/configuration/reference.md)

## Dynamic Configuration

Mokapi reads dynamic configuration and script using [providers](/docs/configuration/dynamic/overview.md).

