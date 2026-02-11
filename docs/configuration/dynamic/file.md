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
You also can use CLI JSON or shorthand syntax to, see [CLI](/docs/configuration/static/cli-usage.md)
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
You can define multiple file names or directory using CLI JSON or shorthand syntax, see [CLI](/docs/configuration/static/cli-usage.md)
```

``` box=tip
You can also define multiple file names or directory by separating them with system's path separator
(Unix=':', Windows=';')
```

### Include
A list of glob patterns that files must match to be processed.
- Applied only to files (not directories)
- If empty or omitted, all files are included 
- If specified, a file must match at least one pattern to be included

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

### Exclude
A list of glob patterns that files must not match to be processed.
- Applied only to files (not directories)
- If empty or omitted, no files are excluded 
- Exclusion is evaluated after include filtering 
- If a file matches any exclude pattern, it is skipped

```yaml tab=File (YAML)
providers:
  file:
    exclude: ["debug.json"]
```
```bash tab=CLI
--providers-file-include debug.json
```
```bash tab=Env
MOKAPI_PROVIDERS_FILE_INCLUDE="debug.json"
```

### Include + Exclude together

When both are set:
1. Files are first filtered by include 
2. Matching files are then filtered by exclude

```yaml
providers:
  file:
    include: ["*.json"]
    exclude: ["debug.json"]
```

- ✔ config.json → included
- ✖ debug.json → excluded
- ✖ config.yaml → not included

### Ignoring Files and Directories
You can create a `.mokapiignore` file in your directory to tell
Mokapi which files and directories to ignore. The structure of this
file follows the [gitignore specification](https://git-scm.com/docs/gitignore)

Example to exclude everything except a specific directory `foo`
```
/*
!/foo
```