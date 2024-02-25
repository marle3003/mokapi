---
title: NPM Provider for dynamic configuration
description: The NPM provider reads dynamic configuration from a NPM module
---
# File

The NPM provider reads dynamic configuration from a NPM module.

``` box=warning noTitle
Mokapi does not install NPM packages itself. You need to make sure the packages are available.
```

## Configuration Example

```yaml tab=File (YAML)
providers:
  npm:
    packages:
      - name: name
        files: ['dist/api.json']
      - name: '@scope/name'
        include: [dist/**/*.json]
```

## Provider Configuration
A list of all options that can be used with the npm provider, refer to
the [reference page](/docs/configuration/reference.md).

### Packages
Defines NPM packages that Mokapi looks for.

#### Files
Specifies an allow list of files. Mokapi will only read these files.

```yaml tab=File (YAML)
providers:
  npm:
    packages:
      - name: name
        files: ['dist/api.json']
```

#### Include
Specifies an array of filenames or patterns. Mokapi will only read these files.

```yaml tab=File (YAML)
providers:
  npm:
    packages:
      - name: name
        include: [dist/**/*.json]
```

### GlobalFolders
By default, Mokapi begins to search for a directory named "node_modules".
It will start in the current directory (where executable file of Mokapi is)
and then work its way up the folder hierarchy, checking each level for a node_modules folder.
Once Mokapi finds the node_modules folder and that contains the module, it will read all files
contained in the NPM module. GlobalFolders allows you to define additional node_modules folders,
which are in a different hierarchy.

```yaml tab=File (YAML)
providers:
  npm:
    globalFolders: [/path/to/specific/node_modules]
```