---
title: File Provider
description: The file provider reads dynamic configuration from a single file or multiple files.
---
# File

The file provider reads dynamic configuration from a single file or 
multiple files.

## Configuration Example

```yaml tab=File (YAML)
providers:
  file:
    directory: /path/to/dynamic/config
```
```bash tab=CLI
--providers-file-directory /path/to/dynamic/config
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORY=/path/to/dynamic/config
```

``` box=tip
You also can use CLI JSON or shorthand syntax to, see [CLI](/docs/configuration/static/cli.md)
```

## Provider Configuration
A list of all options that can be used with the file provider, refer to
the [reference page](/docs/configuration/reference.md).

``` box=warning title=Limitation
Mokapi uses fsnotify to listen to file system notification. There
are issues with if Mokapi runs in a Linux Docker container on Windows
WSL2 host system.
```

### Filename
Defines the path to the configuration file.

```yaml tab=File (YAML)
providers:
  file:
    filename: foobar.yaml
```
```bash tab=CLI
--providers-file-filename foobar.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_FILENAME=foobar.yaml
```

### Directory
Defines the path to the directory that contains the configuration files.
You can also organize your configuration files in subdirectories.

```yaml tab=File (YAML)
providers:
  file:
    directory: /foobar
```
```bash tab=CLI
--providers-file-directory /foobar
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_DIRECTORY=/foobar
```

``` box=tip
You can define multiple file names or directory using CLI JSON or shorthand syntax, see [CLI](/docs/configuration/static/cli.md)
```

``` box=tip
You can also define multiple file names or directory by separating them with system's path separator
(Unix=':', Windows=';')
```

### Include
One or more patterns that a file must match, except when empty. The filter is only applied to files.

```yaml tab=File (YAML)
providers:
  file:
    include: ["*.json", "*.yaml"]
```
```bash tab=CLI
--providers-file-include *.json *.yaml
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_INCLUDE="*.json *.yaml"
```

### Ignoring Files and Directories
You can create a `.mokapiignore` file in your directory to tell
Mokapi which files and directories to ignore. The structure of this
file follows the [gitignore specification](https://git-scm.com/docs/gitignore)

Example to exclude everything except a specific directory `foo`
```
/*
!/foo
```