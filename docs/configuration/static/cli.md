---
title: Use Mokapi CLI
description: This page provides information on how to configure Mokapi using CLI flags.
---
# CLI Syntax

This page explains how to write flag values on the command line: lists, booleans, JSON and the shorthand syntax.
For the common commands see [Running Mokapi](/docs/configuration/running-mokapi.md), for where options come from
and which source wins see the [Overview](/docs/configuration/overview.md). All flags are listed on the
[CLI Flags](/docs/configuration/static/cli-flags) page.

## Command structure

Mokapi has a simple command structure that contains

1. mokapi as the command name
2. optional a list of flags
3. optional a list of positional arguments, see [Running Mokapi](/docs/configuration/running-mokapi.md)

```bash
mokapi [flags] [CONFIG-URL|DIRECTORY|FILE]...
```

Flag names start with a double hyphen and are not case-sensitive, e.g. `--Providers-File-Directory` is the same as
`--providers-file-directory`. Flags take various types of input values, such as strings, numbers, booleans and JSON structures.

## Specify flag values

Many flags are simple string or numeric values, such as `--providers-file-directory`. An equal sign (=) between
flag and value is optional.

```bash
mokapi --providers-file-directory /foo
mokapi --providers-file-directory=/foo # equal sign is optional
```

``` box=info title="Dot syntax"
Before v0.10, flag names used dots, e.g. `--providers.file.directory`. Mokapi still accepts this for compatibility,
but it is not recommended and may be removed in a future release. The documentation uses dashes.
```

### List

One or more values separated by spaces. If a value contains a space, you must put quotation marks around that item.
You can also repeat the flag, or use the index operator to overwrite a value. In the last example the include list
contains only `*.yaml`.

```bash
mokapi --providers-file-include '*.json' '*.yaml'
mokapi --providers-file-include '*.json' --providers-file-include '*.yaml'
mokapi --providers-file-include "C:\Documents and Settings\" C:\Work
mokapi --providers-file-include '*.json' --providers-file-include[0] '*.yaml'
```

A list flag takes all following values until the next argument that starts with `-`. To end the list and pass
positional arguments afterwards, use `--` as separator. With the equal sign form
(`--providers-file-include=*.json`) the flag takes only that one value, so following values are positional arguments.

```bash
mokapi --providers-file-include '*.json' '*.yaml' -- petstore.yaml
mokapi --providers-file-include='*.json' petstore.yaml
```

### Boolean

Binary flag that turns an option on or off if no value is specified. It can also be used with the value `true` or `false`.

```bash
mokapi --api-dashboard
mokapi --api-dashboard true
mokapi --api-dashboard false
mokapi --no-api-dashboard
mokapi --no-api-dashboard true
```

Enabling the dashboard is not necessary as this is the default behavior. The last three examples turn the dashboard off.

### Integer

There is nothing special about using integer values.

```bash
mokapi --api-port 9090
mokapi --providers-git-repositories[0]-auth-github-appId 12345
```

## Values from a file

For some flags the file name can be specified directly, for others a file URL is required.
The flag `--config-file` (alias `--cli-input`) reads all options from a file, see the
[Overview](/docs/configuration/overview.md) for the format.

```bash
mokapi --config-file=/path/to/mokapi.yaml
mokapi --providers-file file:/tmp/file.json
mokapi --root-ca-cert=/path/to/caCert.pem
```

## Shorthand Syntax

Flags with structured values accept JSON. However, entering large JSON lists or structures into the command line can
be tedious and difficult to read. Therefore, Mokapi supports a shorthand syntax that allows a simpler representation
of your configuration.

### Structure values

The shorthand syntax for flat (non-nested) structures makes it easier for you to define your inputs.
This example configures the first Git repository:

```bash
--providers-git-repositories[0] url=https://github.com/foo/foo.git,include=*.json
```

This is equivalent to the following JSON example.

```bash
--providers-git-repositories[0] '{"url":"https://github.com/foo/foo.git","include":["*.json"]}'
```

``` box=warning title=PowerShell
When using PowerShell, you must place the stop-parsing symbol (--%) before any arguments and escape quotation marks.
--providers-git-repositories[0] --% "{\"url\":\"https://github.com/foo/foo.git\",\"include\":[\"*.json\"]}"
```

This corresponds to the following example, where each field is defined separately.

```bash
--providers-git-repositories[0]-url https://github.com/foo/foo.git --providers-git-repositories[0]-include '*.json'
```

Some flags use a name instead of an index, such as the event store. Here `foo` is the name of the API:

```bash
--event-store 'foo={"size": 250}'
--event-store-foo-size 250
```

### List values

For a list of structures, repeat the singular flag or pass several shorthand values to the plural flag.

```bash
--providers-git-repository url=https://github.com/foo/foo.git --providers-git-repository url=https://github.com/bar/bar.git
--providers-git-repositories url=https://github.com/foo/foo.git url=https://github.com/bar/bar.git
```
