---
title: JavaScript Modules
description: Organize Mokapi scripts with built-in, local, Node.js, JSON/YAML and remote modules, and map import paths with module aliases.
---
# JavaScript Modules

When writing scripts, it is common to split code into multiple files or to use third-party
packages. Mokapi supports the following kinds of modules:

- **Built-in modules**: provided by Mokapi, such as `mokapi/faker` or `mokapi/http`
- **Local modules**: your own scripts, imported by relative path
- **Node.js packages**: packages installed in `node_modules`
- **JSON & YAML modules**: data files converted into JavaScript objects
- **Remote modules**: scripts hosted on a web server, GitHub or a CDN

You can also define [module aliases](#module-aliases) to map an import path to another module.

``` box=tip
Mokapi watches all imported modules. If a module changes, every script that imports it and
exports a `default` function is reloaded automatically.
```

## Built-in Modules

Built-in modules start with `mokapi`. For example, `mokapi/faker` generates realistic test
data from JSON schemas. See the [API overview](/docs/javascript-api/overview.md) for a complete list.

```javascript
import { fake } from 'mokapi/faker'
```

## Local Modules

Import your own files using relative paths. The file extension is optional for JavaScript
and TypeScript files.

```javascript
import { someFunc } from './helpers.js'
import { otherFunc } from '../lib'
```

## Node.js Packages

Imports without a relative path are resolved like in Node.js: Mokapi looks for the package in
`node_modules` next to the importing file and then in each parent directory, up to the nearest
`package.json`.

```javascript
import dateTime from 'date-time' // requires: npm install date-time
```

``` box=warning
Mokapi does not run on Node.js. Packages that depend on Node.js built-ins such as `fs` or
`path` will not work. See [Runtime limits](/docs/javascript-api/overview.md) for alternatives.
```

## JSON & YAML Modules

JSON and YAML files can be imported and are converted to objects automatically:

```javascript tab=JavaScript
import users from './users.json'
import envs from './environments.yaml'

console.log(users[0].name)
console.log(envs[0])
```
```json tab=JSON
[
    { "name": "Alice", "email": "alice@foo.bar" },
    { "name": "Bob", "email": "bob@foo.bar" }
]
```
```yaml tab=YAML
- test
- integration
- production
```

## Remote Modules

Modules can also be hosted remotely on public web servers, GitHub or CDNs.

```javascript
import { helper } from 'https://example.com/mokapi-helpers.js'
```

Mokapi loads remote modules with its [HTTP provider](/docs/configuration/dynamic/http.md),
so changes are picked up without restarting Mokapi.

## Module Aliases

A module alias replaces an import path with another one before Mokapi resolves the module.
Use aliases to:

- Run existing code that imports a package under a different name, for example an adapter
  that replaces `msw` with a Mokapi-compatible implementation
- Give long or deeply nested paths a short, stable name
- Switch an implementation in one place without changing every import

Aliases are defined in the [static configuration](/docs/configuration/reference.md#javascript-modules):

```yaml tab=File (YAML)
javascript:
  modules:
    aliases:
      msw: mokapi-msw-adapter
      "msw/*": mokapi-msw-adapter/*
```
```json tab=File (JSON)
{
  "javascript": {
    "modules": {
      "aliases": {
        "msw": "mokapi-msw-adapter",
        "msw/*": "mokapi-msw-adapter/*"
      }
    }
  }
}
```

With this configuration, the imports in your scripts resolve as follows:

| Import          | Resolved to                     |
|-----------------|---------------------------------|
| `msw`           | `mokapi-msw-adapter`            |
| `msw/browser`   | `mokapi-msw-adapter/browser`    |
| `msw/core/http` | `mokapi-msw-adapter/core/http`  |
| `mswx`          | `mswx` (no alias matches)       |

The resolved path is then loaded like any other import, for example from `node_modules`.

### Wildcards

A key may contain **one** `*`. It matches any sequence of characters, including `/`. The
matched text replaces the `*` in the target. A target without `*` maps every match to the
same module.

``` box=info
Aliases are not glob patterns. Only a single `*` is supported; `**`, `?`, `[...]` and `{a,b}`
have no special meaning.
```

### Which Alias Wins

If several keys match an import, the most specific one is used:

1. A key without `*` that matches exactly
2. Otherwise, the wildcard key with the longest part before the `*`
3. On a tie, the wildcard key with the longest part after the `*`

```yaml
javascript:
  modules:
    aliases:
      "msw/*": mokapi-msw-adapter/*
      "msw/node": mokapi-msw-adapter/runtime  # wins for exactly 'msw/node'
```

### Aliases and TypeScript

The alias rules match `paths` in `tsconfig.json`. Add the same mapping there, so your editor
resolves the imports the same way Mokapi does:

```jsonc
{
  "compilerOptions": {
    "paths": {
      "msw": ["./node_modules/mokapi-msw-adapter"],
      "msw/*": ["./node_modules/mokapi-msw-adapter/*"]
    }
  }
}
```
