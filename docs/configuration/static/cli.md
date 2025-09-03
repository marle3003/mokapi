---
title: Use Mokapi CLI
description: This page provides information on how to configure Mokapi using CLI parameters.
---
# Use Mokapi CLI

This page provides information on how to configure Mokapi using CLI parameters. A list of available parameters can be found [here](/docs/configuration/reference.md)

## Command structure

Mokapi has a simple command structure that contains

1. mokapi as the command name
2. optional a list of options and parameters

```bash
mokapi [options and parameters]
```

Parameter and option names are marked with a double hyphen and are not case-sensitive, e.g. "--Providers.File" = "--providers.file"
Parameters can take various types of input values, such as strings, numbers, booleans and JSON structure.

## Specify parameter values

Many parameters are simple string or numeric values, such as the `providers-file-directory`. An equal sign (=) between 
parameter and value is optional.

```bash
mokapi --providers-file-directory /foo
mokapi --providers-file-directory=/foo # equal sign is optional
mokapi --providers.file.directory=/foo # Separation by dot is also possible as old style (previous v0.10)
```

### List

One or more value separated by spaces. If any value contain a space, you must put quotation marks around that item.
Using index operator is also possible which can overwrite values. The include list in the last example only contains `*.yaml`

```bash
mokapi --providers-file-include *.json *.yaml
mokapi --providers-file-include *.json --providers.file.include *.yaml
mokapi --providers-file-include "C:\Documents and Settings\" C:\Work
mokapi --providers-file-include *.json --providers.file.include[0] *.yaml
```

### Boolean

Binary flag that turns an option on or off if no value is specified. It can also be used with a value.

```bash
mokapi --dashboard
mokapi --dashboard true
mokapi --dashboard 1
mokapi --dashboard false
mokapi --no-dashboard
mokapi --no-dashboard true
```

Enabling dashboard is not necessary as this is the default behavior. The last example turns dashboard off.

### Integer

There is nothing special about using integer values.

```bash
mokapi --providers-git-repositories[0]-auth-github-appId 12345
```

## Parameters from file

For some parameters the file name can be specified directly, for others a file URL is required.
The parameter `--configfile` provides the ability to define all parameters in a file.

```bash
mokapi --providers-file file:///tmp/file.json
mokapi --providers-git-rootCaCert=/path/to/caCert.pem
mokapi --cli-input=/path/to/config.json
```

## Positional Arguments

Mokapi CLI supports positional arguments, which are used specifically to provide:
- a configuration URL
- a directory path
- a file path

These arguments must be placed at the end, after all options.

```bash
mokapi [OPTIONS] [config-url|directory|file]
```

- Positional arguments are optional (0-N).
- You can specify multiple files, directories, or URLs.

Using a single configuration file:
```bash
mokapi config.yaml
```
Using multiple configuration files:
```bash
mokapi config1.yaml config2.yaml
```
Using remote configurations:
```bash
mokapi https://example.com/config.yaml
```
Using GIT configurations:
```bash
mokapi git+https://github.com/foo/bar.git
```

## Shorthand Syntax

Mokapi's parameters can accept values in JSON format to simplify the configuration.
However, entering large JSON lists or structures into the command line can be tedious and difficult to read.
Therefore, Mokapi supports a shorthand syntax that allows a simpler representation of your configuration.

### Structure parameters

The shorthand syntax for flat (non-nested) structures makes it easier for you to define your inputs.

```bash
--parameter key1=value1,key2,value2,key3=value3
```

This is equivalent to the following JSON example.

```bash
--parameter '{"key1":"value1","key2","value2","key3"="value3"}'
```

``` box=warning title=PowerShell
When using PowerShell, you must place the stop-parsing symbol (--%) before any arguments and escape quotation marks.
--parameter --% "{\"key1\":\"value1\",\"key2\",\"value2\",\"key3\":\"value3\"}" 
```

This corresponds to the following example, where each parameter is defined separately.

```bash
--parameter-key1 value1 --parameter.key2 value2 --parameter.key3 value3
```

### List parameters

Lists can also be defined as JSON or in short form.

```bash
--parameter value1 value2 value3
--parameter '[value1,value2,value3]'
--parameter value1 --parameter value2 --parameter value3
```