---
title: Use Mokapi CLI
description: This page provides information on how to configure Mokapi using CLI parameters.
---
# Use Mokapi CLI

This page provides information on how to configure Mokapi using CLI parameters. A list of available parameters can be found [here](/docs/configuration/reference.md)

## Command structure

Mokapi has a simple command structure that contains

1. mokapi as the command name
2. optional a list of parameters

```shell
mokapi [options and parameters]
```

Parameter and option names are marked with a double hyphen and are not case-sensitive, e.g. "--Providers.File" = "--providers.file"
Parameters can take various types of input values, such as strings, numbers, booleans and JSON structure.

## Specify parameter values

Many parameters are simple string or numeric values, such as the `providers.file.directory`. An equal sign (=) between 
parameter and value is optional.

```shell
mokapi --providers.file.directory /foo
mokapi --providers.file.directory=/foo
```

### List

One or more value separated by spaces. If any value contain a space, you must put quotation marks around that item.
Using index operator is also possible which can overwrite values. The include list in the last example only contains `*.yaml`

```shell
mokapi --providers.file.include *.json *.yaml
mokapi --providers.file.include *.json --providers.file.include *.yaml
mokapi --providers.file.include "C:\Documents and Settings\" C:\Work
mokapi --providers.file.include *.json --providers.file.include[0] *.yaml
```

### Boolean

Binary flag that turns an option on or off if no value is specified. It can also be used with a value.

```shell
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

```shell
mokapi --providers.git.repositories[0].auth.github.appId 12345
```

## Parameters from file

For some parameters the file name can be specified directly, for others a file URL is required.
The parameter `--configfile` provides the ability to define all parameters in a file.

```shell
mokapi --providers.file file:///tmp/file.json
mokapi --providers.git.rootCaCert=/path/to/caCert.pem
mokapi ---configfile=/path/to/config.json
```

## Shorthand Syntax

Mokapi's parameters can accept values in JSON format to simplify the configuration.
However, entering large JSON lists or structures into the command line can be tedious and difficult to read.
Therefore, Mokapi supports a shorthand syntax that allows a simpler representation of your configuration.

### Structure parameters

The shorthand syntax for flat (non-nested) structures makes it easier for you to define your inputs.

```shell
--parameter = key1=value1,key2,value2,key3=value3
```

This is equivalent to the following JSON example.

```shell
--parameter = '{"key1":"value1","key2","value2","key3"="value3"}'
```

This corresponds to the following example, where each parameter is defined separately.

```shell
--parameter.key1 value1 --parameter.key2 value2 --parameter.key3 value3
```

### List parameters

Lists can also be defined as JSON or in short form.

```shell
--parameter value1 value2 value3
--parameter '[value1,value2,value3]'
--parameter value1 --parameter value2 --parameter value3
```