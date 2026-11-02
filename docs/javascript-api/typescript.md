---
title: TypeScript & IntelliSense
description: Get autocompletion and type checking for Mokapi scripts in your editor.
---

# TypeScript & IntelliSense

Mokapi provides type definitions for its JavaScript API. With them, editors like
VS Code offer autocompletion, inline documentation and type checking,
for TypeScript **and** plain JavaScript scripts.

There are two ways to get the types:

|                     | `@types/mokapi`              | `go-mokapi` + `tsconfig.json`          |
|---------------------|------------------------------|----------------------------------------|
| Setup               | One command                  | Install package and add configuration  |
| Up to date          | May lag behind new releases  | Matches the installed Mokapi version   |
| Best for            | Getting started quickly      | Using the latest API features          |

## Option 1: @types/mokapi (quickest)

```bash
npm install --save-dev @types/mokapi
```

No further configuration is needed — TypeScript and VS Code pick up packages
under `@types` automatically.

## Option 2: Types from go-mokapi (always current)

The `go-mokapi` npm package contains the type definitions for exactly the Mokapi
version you install. If you already installed Mokapi via npm, you only need the
configuration below.

```bash
npm install --save-dev go-mokapi
```

Add a `tsconfig.json` (or `jsconfig.json` for JavaScript-only projects) next to your scripts:

```jsonc
{
  "compilerOptions": {
    "lib": ["es2017"],
    "paths": {
      "mokapi": ["./node_modules/go-mokapi/types/index"],
      "mokapi/*": ["./node_modules/go-mokapi/types/*"]
    },
    "types": ["go-mokapi/types/global"]
  }
}
```

## Check that it works

```typescript
import { on } from 'mokapi'

export default function () {
  on('http', (request, response) => {
    // hover over `request` — your editor should show its type
  })
}
```