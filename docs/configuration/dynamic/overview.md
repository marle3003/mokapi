---
title: Dynamic Configuration
description: How Mokapi reads specifications, scripts and data files from providers, recognizes their type and reloads them when they change.
---
# Dynamic Configuration

Dynamic configuration is everything Mokapi mocks: OpenAPI, AsyncAPI, LDAP and mail
specifications, the scripts that extend them, and data files those scripts read.
Mokapi reloads it while it runs, so you don't restart it after an edit.

The startup options, such as ports, log level or provider settings, are the static
configuration. They are read once and a change needs a restart. See
[Configuration](/docs/configuration/overview.md) for those and the
[CLI Flags](/docs/configuration/static/cli-flags) for the full list.

## Providers

A provider reads dynamic configuration from a source and reports changes.

- [File](/docs/configuration/dynamic/file.md) reads files and directories and watches them for changes.
- [HTTP](/docs/configuration/dynamic/http.md) downloads files from a URL and checks it again every 3 minutes by default (`pollInterval`).
- [Git](/docs/configuration/dynamic/git.md) clones a repository and pulls it every 3 minutes by default (`pullInterval`).
- [NPM](/docs/configuration/dynamic/npm.md) reads files from a package in a `node_modules` folder.

You can combine providers. Files from different providers describing the same API are
[merged](#several-files-for-one-api).

## How Mokapi recognizes a file

Mokapi first parses the file by its extension. `.yml` and `.yaml` are parsed as YAML,
`.json` as JSON. `.js`, `.cjs`, `.mjs`, `.ts`, `.lua` and `.ldif` are kept as text. A file with any
other extension is tried as JSON, then as YAML, and kept as text if both fail. A trailing
`.tmpl` is removed first and the content is rendered as a template, so `api.yaml.tmpl` is
parsed as YAML.

Parsed JSON and YAML files are then identified by a top-level key:

| Key        | Type                                        |
|------------|---------------------------------------------|
| `openapi`  | OpenAPI                                     |
| `swagger`  | Swagger 2.0                                 |
| `asyncapi` | AsyncAPI, version 2 or 3                    |
| `ldap`     | LDAP                                        |
| `mail`     | Mail (SMTP and IMAP)                        |
| `smtp`     | Mail, deprecated format                     |

The `smtp` format is converted on load and logs a deprecation warning. Use `mail` for new
configuration, see [Mail](/docs/mail/overview.md).

A file without one of these keys is still loaded, but Mokapi doesn't mock anything from it.
It's available as data, for example for a `$ref` or for [`open()`](/docs/javascript-api/open.md) in a script.

Mokapi starts a script when its extension is `.js`, `.cjs`, `.mjs`, `.ts` or `.lua`. Any other
text file is not executed.

## Reloading

The file provider collects file system events for one second and then processes them together.
For a changed file, Mokapi reads it again and skips it if the content is identical. Otherwise it
logs `reloading <file>`, parses the file again and updates the mock. Scripts run again from the
start.

If the new content doesn't parse or fails validation, Mokapi logs the error and keeps
the last working version of the mock. Fix the file and save it again. This applies to
specifications. Mokapi stops a script before it runs the new version, so if the new version
fails, the old one is no longer running either.

If you delete a file, Mokapi removes what it contributed. A server that was defined only in that file
stops, and a script stops with all its event handlers and jobs. The HTTP provider only detects changed
content, so it doesn't report a file that is no longer available.

## References between files

A `$ref` to another file, a script that imports a module with `import` or `require`,
and a file that a script reads with `open()` all link the two files. Mokapi watches the
referenced file too, even when it's outside the directory you passed to the file provider.
When it changes, Mokapi parses the referencing file again, resolves all references
against the new content and updates the mock. If a script imports the file, the script runs again.

A reference can be a relative path or a URL. Mokapi reads it with the provider for its scheme
(`file`, `http`, `https`, `git` or `npm`).

If the referenced file is deleted, the referencing file is parsed again as well.

## Several files for one API

If several files describe the same API, Mokapi merges them into one mock. Files are applied in
alphabetical order of their file names, so a patch can add servers or override single values
without changing the original. See [Patching](/docs/configuration/patching.md).

## Example

A directory with an OpenAPI specification and a script:

``` box=tree title="mocks"
mocks/
├── petstore.yaml
└── petstore.js
```

```bash
mokapi ./mocks
```

Mokapi reads both files and starts the script. Change a response in `petstore.yaml` and save.
After about a second, the log shows `reloading` for that file, and the next request returns the
new response. If you save `petstore.js`, the script runs again with its new code.

By default, Mokapi skips files and directories whose name starts with `_`, see
[`--providers-file-skip-prefix`](/docs/configuration/static/cli-flags#providers-file-skip-prefix).

## Dashboard

The **Configs** tab in the [dashboard](/docs/get-started/dashboard.md) lists all loaded files. The
**Configs** tab of an API shows the files that Mokapi merged into it.
