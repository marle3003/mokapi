---
title: Use Mokapi CLI
description: This page provides information on how to configure Mokapi using CLI parameters.
---
# Introduction

This page provides information on how to configure Mokapi using CLI parameters. A list of available parameters can be found [here](/docs/configuration/reference.md)

## Command structure

Mokapi has a simple command structure that contains

1. mokapi as the command name
2. optional a list of parameters

```
mokapi [--key value]
```

Parameter names are marked with a double hyphen and are not case-sensitive, e.g. "--Providers.File" = "--providers.file"
Parameters can take various types of input values, such as strings, numbers, booleans and JSON structure.

## Specify parameter values

Many parameters are simple string or numeric values, such as the `providers.file.directory`. An equal sign (=) between 
parameter and value is optional.

```
mokapi --providers.file.directory /foo
mokapi --providers.file.directory=/foo
```

### List

One or more value separated by spaces. If any value contain a space, you must put quotation marks around that item.

```
mokapi --providers.file.include *.json *.yaml
mokapi --providers.file.include *.json --providers.file.include *.yaml
mokapi --providers.file.include "C:\Documents and Settings\" C:\Work
```

### Boolean

Binary flag that turns an option on or off if no value is specified. It can also be used with a value.

```
mokapi --dashboard
mokapi --dashboard true
mokapi --dashboard 1
mokapi --dashboard false
mokapi --no-dashboard
mokapi --no-dashboard true
```

Enabling dashboard is not necessary as this is the default behavior. The last example turns dashboard off.