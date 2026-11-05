---
title: "Mokapi Meets MSW: Reuse Your Browser Mocks as a Contract-Checked Mock Server"
description: Run your existing MSW handlers in Mokapi with mokapi-msw-adapter. Keep one set of mocks and validate every response against your OpenAPI contract.
subtitle: Your Storybook and browser mocks already describe your API. Serve the same MSW handlers over real HTTP with Mokapi and find out when they drift from the contract.
image:
  url: "/mokapi-msw-adapter-flow.png"
  alt: Diagram showing a shared handler module that imports msw. In the browser msw resolves to the real MSW package, in Mokapi a module alias resolves msw to mokapi-msw-adapter.
tech: http
tags: ['HTTP', 'Testing', 'Development']
links:
  items:
    - title: Module Aliases
      description: Map an import path such as msw to another module
      href: /docs/javascript-api/modules#module-aliases
    - title: Catch Breaking API Changes Before They Reach Production
      description: Validate traffic against your OpenAPI contract at runtime
      href: /resources/blogs/guard-your-api-contracts
    - title: Get Started with Mokapi
      description: Up and running in seconds
      href: /docs/get-started/installation
---

# Mokapi Meets MSW: Reuse Your Browser Mocks as a Contract-Checked Mock Server

## Two Sets of Mocks for One API

If you build frontends, you probably know [MSW](https://mswjs.io) (Mock Service Worker). You describe
responses with `http.get(...)` and `HttpResponse.json(...)`, and MSW answers requests in the browser,
in Storybook or in your unit tests. It is a great way to develop a UI without a running backend.

The limit is the browser. A service worker only answers requests made by the browser that registered it. Your Playwright
suite running against a deployed frontend, a second service, or a quick `curl` can't reach those handlers.
So a team adds a second mock, this time a real mock server, and now there are two descriptions of the same API.

Two sets of mocks have a second problem: both are handwritten. Nobody checks them against the OpenAPI
contract. A field gets renamed in the spec, the handlers still return the old shape, the stories stay green,
and the mismatch shows up later in an integration environment.

What you want is one set of handlers that works in the browser and as a real HTTP server, and that fails when it
no longer matches the contract.

---

## The Idea: Change What `msw` Means

[mokapi-msw-adapter](https://github.com/marle3003/mokapi-msw-adapter) is a small npm package that makes this
possible without touching your handlers. Your shared module keeps its normal import:

```js
import { http, HttpResponse } from 'msw';
```

What differs is how that import is resolved:

- In the browser (Storybook, Vite, ...), `msw` resolves to the real MSW package.
- In Mokapi, a [module alias](/docs/javascript-api/modules#module-aliases) resolves `msw` to
  `mokapi-msw-adapter`. The adapter provides `http` and `HttpResponse` with the same shape as MSW.

![Diagram showing a shared handler module that imports msw, resolved to the real MSW package in the browser and to mokapi-msw-adapter in Mokapi](/mokapi-msw-adapter-flow.png "One handler module, two runtimes: the import of msw is resolved differently")

The module's default export calls `register()`. Mokapi runs the default export when it loads the file and
turns your handlers into native Mokapi routes. A browser bundler only uses the named exports, so the
default export never runs there.

---

## Walkthrough: One Profile Endpoint

Requirements: Mokapi 0.56.0 or newer and Node.js 22.12.0 or newer.

Install the adapter. If you don't have Mokapi yet, `go-mokapi` installs the Mokapi binary through npm:

```sh
npm install --save-dev mokapi-msw-adapter go-mokapi
```

### Write Handlers as Scenarios

Group your handlers into named scenarios. Here, `success` is the normal case and `error` simulates a failing backend:

```js
// src/mocks/profile.js
import { http, HttpResponse } from 'msw';
import { register } from 'mokapi-msw-adapter/register';

export const scenarios = {
  success: [
    http.get('/api/profile', () =>
      HttpResponse.json({ name: 'Jane Doe', email: 'jane.doe@example.com' }),
    ),
  ],
  error: [
    http.get('/api/profile', () =>
      HttpResponse.json({ message: 'Internal error' }, { status: 500 }),
    ),
  ],
};

// Executed by Mokapi only. Storybook and the browser only use `scenarios`.
export default function () {
  register(scenarios);
}
```

Your browser tooling imports `scenarios` and hands them to MSW as before. The adapter's `register` import
is only relevant to Mokapi: in a browser bundle `mokapi-msw-adapter/register` resolves to a stub that throws when called.

### Describe the Contract

The OpenAPI specification is what Mokapi validates against:

```yaml
# src/mocks/openapi.yaml
openapi: 3.1.0
info:
  title: My API
  version: 1.0.0
servers:
  - url: http://localhost:3000
paths:
  /api/profile:
    get:
      responses:
        '200':
          description: Profile
          content:
            application/json:
              schema:
                type: object
                required: [name, email]
                properties:
                  name: { type: string }
                  email: { type: string }
        '500':
          description: Error
          content:
            application/json:
              schema:
                type: object
                properties:
                  message: { type: string }
```

### Configure Mokapi

Create `mokapi.yaml` in your project root. The alias is what makes `import ... from 'msw'` load the adapter,
and the file provider loads the script and the contract from `src/mocks`:

```yaml
javascript:
  modules:
    aliases:
      msw: mokapi-msw-adapter
providers:
  file:
    directories:
      - src/mocks
```

See the [file provider](/docs/configuration/dynamic/file) if the folder contains files you want to exclude.

### Start Mokapi and Call the API

```sh
npx mokapi
```

Wait until the log shows `executing script ...profile.js`, then call the endpoint:

```sh
curl http://localhost:3000/api/profile
# {"name":"Jane Doe","email":"jane.doe@example.com"}

curl "http://localhost:3000/api/profile?__mockScenario=error"
# HTTP 500 {"message":"Internal error"}
```

The query parameter `__mockScenario` selects the scenario for that single request. Without it, the adapter uses
`success`. You don't need to declare the parameter in the OpenAPI contract.

---

## What You Get Beyond Reuse

### Handlers That Are Checked Against the Contract

The adapter hands the result to Mokapi as response data, so Mokapi validates it against the schema of your
OpenAPI specification. Suppose the API renames `name` to `fullName` in the browser handler, but not in the spec:

```js
drift: [
  http.get('/api/profile', () =>
    HttpResponse.json({ fullName: 'Jane Doe', email: 'jane.doe@example.com' }),
  ),
]
```

Calling `/api/profile?__mockScenario=drift` returns HTTP 500 with a message that names the problem:

```
HTTP body marshalling failed.

Body: {
  "email": "jane.doe@example.com",
  "fullName": "Jane Doe"
}

Validation error count 1:
	- #/required: required properties are missing: name
```

The same handler would pass silently in the browser. In Mokapi, the drift becomes visible.

### Scenarios Per Request

Because the scenario travels with the request and not in a global switch, parallel requests don't interfere.
Two stories, or two Playwright tests running at the same time, can use `success` and `error` against the same mock
server. If you prefer another parameter name or default, `register()` accepts options:

| Option                      | Default            | Description                                                  |
|-----------------------------|--------------------|--------------------------------------------------------------|
| `options.scenarioParameter` | `'__mockScenario'` | Query parameter that selects the scenario                    |
| `options.defaultScenario`   | `'success'`        | Scenario used when the parameter is missing                  |
| `options.api`               | all APIs           | `info.title` of the OpenAPI contract to restrict handlers to |

### More Than One API

By default, `register()` applies the handlers to every API that Mokapi mocks. If several contracts share a path,
pass the `info.title` of the contract:

```js
export default function () {
  register(scenarios, { api: 'My API' });
}
```

Requests to other APIs keep Mokapi's usual behavior.

### A Real Mock Server for Everyone

Once Mokapi runs, it isn't only your browser that sees the mock. Playwright tests, a backend service under
development, or a teammate with `curl` use the same handlers. Combine this with the dashboard and you can inspect
every request and response. See
[End to End Testing with Mocked APIs](/resources/blogs/end-to-end-testing-with-mocked-apis).

---

## What Is Supported

The adapter implements a subset of MSW's API. Before you move an existing mock suite, check it against this list:

| Supported                                                                                       | Not supported                                                            |
|-------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| `http.get`, `post`, `put`, `patch`, `delete`, `head`, `options`, `all`                          | Path patterns such as `/users/:id` or `/files/*`, absolute URLs          |
| `new HttpHandler(method, path, resolver)`, including custom methods                             | Responses other than `HttpResponse.json()`                               |
| Synchronous and asynchronous resolvers                                                          | Status codes without a body (204, 205, 304)                              |
| `HttpResponse.json(body, { status, headers })`                                                  | Cookies, streaming, network errors, `delay()`, `passthrough()`           |
| In resolvers: `params`, `request.url`, `method`, `headers.get()`, `json()`, `text()`           | Complete `Request`, `Response`, `Headers` and `URL` implementations      |

Only exact relative paths work. Unsupported usage throws an error that names the problem, and Mokapi logs it.

### When You Get Random Data

If the adapter throws, for example because you asked for a scenario that doesn't exist (`?__mockScenario=nope`)
or the selected scenario has no handler for that method and path, Mokapi logs the error and answers with data
generated from your OpenAPI contract. That is intended Mokapi behavior, but it can look like your mock returned
random values. If that happens, check the log first:

```
level=error msg="unable to execute event handler: Error: Unknown mock scenario: nope"
```

The same can happen for requests that arrive right after startup, before Mokapi has finished running the script.
Wait for the `executing script` log line before you start tests.

---

## Conclusion

If your frontend already uses MSW, you can reuse those handlers in Mokapi without rewriting them: add the alias,
export a default function that calls `register()`, and you have a mock server that checks every answer against
your OpenAPI contract. It covers JSON APIs with exact paths; for anything beyond that, Mokapi's own
[JavaScript API](/resources/blogs/bring-your-mock-apis-to-life-with-javascript) gives you full control.

The adapter is open source (MIT). Find the code, the full README and the issue tracker on
[GitHub](https://github.com/marle3003/mokapi-msw-adapter), and the package as
[mokapi-msw-adapter](https://www.npmjs.com/package/mokapi-msw-adapter) on npm.
