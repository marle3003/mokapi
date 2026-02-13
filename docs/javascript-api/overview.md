---
title: JavaScript API Documentation
description: Extend and customize Mokapi using JavaScript. Learn how to structure scripts, use built-in modules, and handle runtime limitations.
---

# Mokapi JavaScript API

Mokapi allows you to extend and customize its behavior using JavaScript.
JavaScript is used to implement dynamic logic such as request handling, data generation,
scheduled jobs, and event-driven workflows.

Mokapi executes JavaScript in a **simple and explicit way**:
only files that export a **default function** are executed.
All other JavaScript files are treated as **modules**.

This documentation explains:
- How JavaScript is executed in Mokapi
- Script vs module structure
- Runtime environment limitations
- Built-in modules

## How JavaScript Is Used in Mokapi

JavaScript can be used for tasks such as:

- Handling HTTP, Kafka, and other protocol events
- Generating dynamic mock data
- Scheduling background jobs
- Transforming or enriching request and response data

Each JavaScript file is either:
- an **executable script**, or
- a **module** imported by other scripts

## Executable Scripts (Default Function)

A JavaScript file is executed **only if it exports a default function**.

```js
export default function (ctx) {
  // executed by Mokapi
}
```

- The default function is the entry point of the script
- Mokapi invokes this function directly 
- All execution logic must be placed inside this function

If a file does not export a default function, Mokapi **will not execute it**.

## Runtime Environment

Mokapi executes JavaScript in its own runtime.
It is **not** a Node.js environment and **not** a browser environment.

As a result, many APIs that are commonly available in Node.js or browsers are **not available** in Mokapi.

### ⚠️ Important Limitations

- Node.js built-ins such as `fs`, `path`, `os`, or `net` are not available
- Browser APIs such as `window`, `document`, or `fetch` are not available
- Third-party packages that rely on Node.js or browser APIs may fail

### Provided Alternatives

Mokapi provides its own APIs for common tasks:

- **HTTP requests**  
  Use the `fetch` function from `mokapi/http`:
  ```js
  import { fetch } from "mokapi/http"
  ```
- **File access**  
  Use the global open() function to read files:
  ```js
  const text = open("data/example.json")
  ```
- **Environment variables**  
  Use env() from the core API:
  ```js
  import { env } from "mokapi"
  ```

### Improve Startup Performance

Mokapi scans all configured directories at startup to discover API specifications and JavaScript files.
Starting a JavaScript runtime is memory-intensive, especially in projects with many JavaScript files.

To improve startup time and reduce memory usage, follow these best practices:

#### Use a Single Entry File per Script

Structure your JavaScript so that only entry files export a default function.

```
mocks/
└─ api/
   └─ users/
      ├─ index.js        # executable script (default export)
      ├─ handlers.js     # module
      └─ utils.js        # module
```

Only index.js should export a default function. All other files should be imported as modules.

#### Configure an Include Filter for JavaScript Files

If a directory contains many JavaScript files (for example helpers or shared modules),
configure an include filter so Mokapi only considers entry files:

```yaml
providers:
  file:
    directory:
      - path: ./mocks
        include:
          - "**/index.js"
```

This prevents Mokapi from inspecting every JavaScript file and significantly reduces
startup time and memory usage.

## Modules

JavaScript files without a **default export** are treated as **modules**.
Modules allow you to organize and reuse code.

```js
export function normalizeUser(user) {
  return { ...user, active: true }
}
```

- Modules are not executed by Mokapi 
- Can be imported by executable scripts or other modules 
- Used to organize shared logic

For more details on the different types of modules (built-in, local filesystem, JSON/YAML), see the dedicated [JavaScript Modules guide](/docs/javascript-api/modules.md).

## Module Resolution

Mokapi resolves imports using the same algorithm as Node.js:

1. The directory of the importing file 
2. Any node_modules directory in the same folder 
3. Parent directories up to the nearest package.json

> Module resolution only determines where Mokapi looks for the file.
  Runtime limitations still apply: modules that rely on Node.js or browser APIs may fail.

## TypeScript Support

``` box=tip url=[@types/mokapi on npm](https://www.npmjs.com/package/@types/mokapi)
Mokapi provides TypeScript type definitions for its JavaScript API.
Install them using:
`npm install @types/mokapi --save-dev` 
```

## Built-in Modules

Mokapi provides a set of built-in JavaScript modules that can be used from executable scripts
and custom modules.

### mokapi (Core API)

Provides core functions for scheduling jobs, handling events, and accessing environment variables.

| Functions                                                                    | Description                                       |
|------------------------------------------------------------------------------|---------------------------------------------------|
| [cron( expression, handler, \[args\] )](/docs/javascript-api/mokapi/cron.md) | Schedules a periodic job using a cron expression. |
| [date( \[args\] )](/docs/javascript-api/mokapi/date.md)                      | Returns a formatted date string.                  |
| [env( name )](/docs/javascript-api/mokapi/env.md)                            | Gets the value of an environment variable.        |
| [every( interval, handler, \[args\] )](/docs/javascript-api/mokapi/every.md) | Runs a periodic job at a fixed interval.          |
| [on( event, handler, \[args\]](/docs/javascript-api/mokapi/on.md) )          | Registers an event handler.                       |
| [sleep( time )](/docs/javascript-api/mokapi/sleep.md)                        | Pauses execution                                  |
| [marshal( value, \[encoding\] )](/docs/javascript-api/mokapi/marshal.md)     | Converts a value to a marshaled string.           |

### mokapi/http (HTTP Requests)

Functions to send HTTP requests within Mokapi scripts.

| Functions                                                                         | Description                   |
|-----------------------------------------------------------------------------------|-------------------------------|
| [get( url, \[args\] )](/docs/javascript-api/mokapi-http/get.md)                   | Sends an HTTP GET request.    |
| [post( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/post.md)       | Sends an HTTP POST request    |
| [put( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/put.md)         | Sends an HTTP PUT request     |
| [head( url, \[args\] )](/docs/javascript-api/mokapi-http/head.md)                 | Sends an HTTP HEAD request    |
| [patch( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/patch.md)     | Sends an HTTP PATCH request   |
| [delete( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/delete.md)   | Sends an HTTP DELETE request  |
| [options( url, \[body\], \[args\] )](/docs/javascript-api/mokapi-http/options.md) | Sends an HTTP OPTIONS request |
| [fetch( url, \[opts\] )](/docs/javascript-api/mokapi-http/fetch.md)               | Fetch using Mokapi's API      |

### mokapi/faker (Mock Data Generator)

Generates random test data based on a schema.

| Functions                                                   | Description                             |
|-------------------------------------------------------------|-----------------------------------------|
| [fake( schema )](/docs/javascript-api/mokapi-faker/fake.md) | Generates mock data based on a schema.  |

### mokapi/kafka (Kafka Messaging)

Functions for interacting with Kafka topics.

| Functions                                                           | Description                            |
|---------------------------------------------------------------------|----------------------------------------|
| [produce( \[args\] )](/docs/javascript-api/mokapi-kafka/produce.md) | Publishes a message to a Kafka topic.  |

### mokapi/mustache (Template Engine)

Processes Mustache templates with dynamic data.

| Functions                                                                   | Description                                      |
|-----------------------------------------------------------------------------|--------------------------------------------------|
| [render( template, scope )](/docs/javascript-api/mokapi-mustache/render.md) | Renders a Mustache template with provided data.  |

### mokapi/yaml (YAML Parsing & Serialization)

Handles YAML data parsing and conversion.

| Functions                                                           | Description                          |
|---------------------------------------------------------------------|--------------------------------------|
| [parse( text )](/docs/javascript-api/mokapi-yaml/parse.md)          | Parses a YAML string into an object. |
| [stringify( value )](/docs/javascript-api/mokapi-yaml/stringify.md) | Converts an object to YAML.          |

### mokapi/encoding (Encoding Utilities)

Functions for encoding and decoding data.

| Functions                                                                       | Description                 |
|---------------------------------------------------------------------------------|-----------------------------|
| [base64.encode( input )](/docs/javascript-api/mokapi-encoding/base64-encode.md) | Encodes a string to Base64. |
| [base64.decode( input )](/docs/javascript-api/mokapi-encoding/base64-decode.md) | Decodes a Base64 string.    |

### mokapi/file

Functions for working with files

| Functions                                                                    | Description                                    |
|------------------------------------------------------------------------------|------------------------------------------------|
| [read( path )](/docs/javascript-api/mokapi-file/read.md)                     | Reads the contents of a file.                  |
| [writeString( path, s )](/docs/javascript-api/mokapi-file/write-string.md)   | Writes a string to a file at the given path.   |
| [appendString( path, s )](/docs/javascript-api/mokapi-file/append-string.md) | Appends a string to a file at the given path.  |




