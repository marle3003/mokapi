---
title: NPM Provider for dynamic configuration
description: The NPM provider reads dynamic configuration files, such as OpenAPI or AsyncAPI specs and scripts, from packages in a node_modules folder.
---
# NPM

The NPM provider reads dynamic configuration from files inside an NPM package. This lets you ship
specs and scripts as a package and reuse them across projects.

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
        include: ['dist/**/*.json']
```
```bash tab=CLI
--providers-npm-package name=name,file=dist/api.json
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGE=name=name,file=dist/api.json
```

## Provider Configuration
A list of all options that can be used with the npm provider, refer to
the [CLI Flags](/docs/configuration/static/cli-flags#providers-npm) page.

### Packages
Defines the NPM packages that Mokapi reads. Use the package name as you would in `npm install`,
including the scope for scoped packages (`@scope/name`).

Mokapi looks for the package folder `node_modules/<name>` starting in the working directory
(the directory you start Mokapi from, not the location of the executable) and then checks each
parent directory until it finds the package. If it doesn't find it, Mokapi checks the
[global folders](#global-folders). If the package is still missing, Mokapi logs an error
(`module <name> not found`) and continues without it.

Mokapi resolves the package folder once at startup. A package that you install later is not
picked up until you restart Mokapi. Mokapi does not read `package.json`, so the `main` field has
no effect. By default Mokapi reads all files in the package folder, including nested folders.

Mokapi watches the package folder. If a file in the package changes, for example after you run
`npm install` to update the package, Mokapi reloads that file. The package folder itself must
exist at startup.

Files whose name starts with an underscore (`_`) are skipped, and a `.mokapiignore` file inside
the package is honored, as in the [File Provider](/docs/configuration/dynamic/file.md).

``` box=tip
To read only the files you need, restrict them with `files` or `include`. Otherwise Mokapi
reads everything in the package, including a nested `node_modules` folder.
```

#### Files
Specifies an allow list of files. A file is read only if its path relative to the package root
exactly equals one of the entries. Use forward slashes and no leading `./`. Patterns are not
supported here, use `include` for that.

```yaml tab=File (YAML)
providers:
  npm:
    packages:
      - name: name
        files: ['dist/api.json']
```
```bash tab=CLI
--providers-npm-package name=name,file=dist/api.json
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_PACKAGE=name=name,file=dist/api.json
```

#### Include
Specifies a list of glob patterns. A file is read if its path relative to the package root
matches at least one pattern. The patterns follow the same syntax as the `include` option of the
[File Provider](/docs/configuration/dynamic/file.md), including `**` to match nested folders.

```yaml tab=File (YAML)
providers:
  npm:
    packages:
      - name: name
        include: ['dist/**/*.json']
```

#### Combining Files and Include
A file is read if it is listed in `files` or matches a pattern in `include`. If you set neither,
Mokapi reads all files of the package. Files that match neither option are ignored.

```yaml tab=File (YAML)
providers:
  npm:
    packages:
      - name: name
        files: ['openapi.yaml']
        include: ['schemas/**/*.json']
```

### Global Folders
Defines additional folders where Mokapi looks for packages. Mokapi checks them only if it
doesn't find the package in a `node_modules` folder of the working directory or its parents.

Mokapi checks each folder in the given order and looks for `<folder>/<name>`. Unlike the
default search, it does not append `node_modules`, so point to the `node_modules` folder itself,
for example the output of `npm root -g`. A relative path is resolved against the working directory.

```yaml tab=File (YAML)
providers:
  npm:
    globalFolders: [/path/to/specific/node_modules]
```
```bash tab=CLI
--providers-npm-global-folder /path/to/specific/node_modules
```
```bash tab=Env
MOKAPI_PROVIDERS_NPM_GLOBAL_FOLDER=/path/to/specific/node_modules
```

## Referencing Package Files

Files from a package get an `npm://` URL: `npm://<name>/<path>`, for example
`npm://name/dist/api.json`. For a scoped package, pass the scope as query parameter:
`npm://name/dist/api.json?scope=@scope`. You can use these URLs as `$ref` in a spec. A `$ref`
reads the file directly, so `files` and `include` don't restrict it.

## NPM Provider and JavaScript Modules

The provider and the `import` of npm packages in a script are two separate mechanisms.
The provider only delivers files to Mokapi as dynamic configuration. When a script
imports a package, Mokapi resolves it on its own: first relative to the script, then in the
`node_modules` folder of the script's directory and each parent directory. For a package folder
it reads the `main` field of `package.json`, then falls back to `index.js`, `index.ts` or
`index.json`. A script therefore doesn't need the package to be configured as an NPM provider
package, and listing a package here doesn't make it importable.
