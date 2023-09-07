---
title: File Provider for dynamic configuration
description: The file provider reads dynamic configuration from a single file or multiple files.
---
# File

The file provider reads dynamic configuration from a single file or 
multiple files.

``` box=tip
Other providers such as Git provider also use the file provider
```

## Configuration
A list of all options that can be used with the file provider, refer to
the [reference page](/docs/configuration/reference.md).

``` box=limitation
Mokapi uses fsnotify to listen to file system notification. There
are issues with if Mokapi runs in a linux docker container on windows
WSL2 host system.
```

### Filename
Defines the path to the configuration file.
```bash tab=CLI
--providers.file.filename=foobar.yaml
```
```bash tab=Env
MOKAPI_Providers_File_Filename=foobar.yaml
```
```yaml tab=File (YAML)
providers:
  file:
    filename: foobar.yaml
```

### Directory
Defines the path to the directory that contains the configuration files.
You can also organize your configuration files in subdirectories. 
`Filename` and `Directory` are mutually exclusive and `Directory` is weighted higher.

```bash tab=CLI
--providers.file.directory=/foobar
```
```bash tab=Env
MOKAPI_Providers_File_Directory=/foobar
```
```yaml tab=File (YAML)
providers:
  file:
    directory: /foobar
```

``` box=tip
You can define multiple file names or directory by separating them with system's path separator
(Unix=':', Windows=';')
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