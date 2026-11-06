---
title: "Storybook Stories That Check Your OpenAPI Contract: Switch Between MSW and Mokapi"
description: Switch Storybook stories between MSW and Mokapi with mokapi-storybook-addon. Reuse your handlers and validate every response against your OpenAPI contract.
subtitle: Your stories already run on MSW handlers. Add a toolbar switch, send the same requests to Mokapi, and see in the Storybook panel when a mock no longer matches the API.
image:
  url: "/mokapi-storybook-addon-panel.png"
  alt: The Mokapi panel in Storybook showing an HTTP 500 response for the Invalid scenario with the validation error that the required property email is missing.
tech: http
tags: ['HTTP', 'Testing', 'Development']
links:
  items:
    - title: Mokapi Meets MSW
      description: Reuse your MSW handlers as a contract-checked mock server
      href: /resources/blogs/mokapi-meets-msw
    - title: Catch Breaking API Changes Before They Reach Production
      description: Validate traffic against your OpenAPI contract at runtime
      href: /resources/blogs/guard-your-api-contracts
    - title: End to End Testing with Mocked APIs
      description: Use the same mock server in your Playwright tests
      href: /resources/blogs/end-to-end-testing-with-mocked-apis
    - title: Get Started with Mokapi
      description: Up and running in seconds
      href: /docs/get-started/installation
---

# Storybook Stories That Check Your OpenAPI Contract: Switch Between MSW and Mokapi

## Green Stories, Broken Contract

Storybook lets you build and review UI components without a backend. Most teams fill the gap with
[MSW](https://mswjs.io) handlers: each story gets a handler that returns the JSON the component needs.

The handlers are handwritten, and nothing compares them with the OpenAPI specification. The backend team marks `email` as
required, a story mock forgets it, and the story stays green. The component renders fine against a response the real
API would never send. You find out in an integration environment, or in production.

In [Mokapi Meets MSW](/resources/blogs/mokapi-meets-msw) we showed how Mokapi can serve your MSW handlers as a real
HTTP mock server that validates every response against the contract. This article covers the Storybook side:
**how to switch a story between MSW and Mokapi with a toolbar button, and how to see the validation result next to your component.**

---

## The Idea: One Handler Module, Two Backends

[mokapi-storybook-addon](https://github.com/marle3003/mokapi-storybook-addon) adds a `mockBackend` switch to the
Storybook toolbar with two values:

- **MSW (browser):** nothing changes. Your story's handlers run in the browser's MSW service worker.
- **Mokapi (OpenAPI validation):** the addon replaces the story's handlers with forwarding handlers. They send each
  request through a Vite development proxy (`/mokapi` by default) to Mokapi and add the story's scenario as the query
  parameter `__mockScenario`.

On the Mokapi side, [mokapi-msw-adapter](https://github.com/marle3003/mokapi-msw-adapter) serves the same handler module and
picks the handler list by that scenario name. Mokapi validates the result against your OpenAPI specification. Its answer,
including a validation error, goes back to your component unchanged.

![Diagram: the Storybook toolbar switches between browser MSW handlers and forwarding to Mokapi, which answers with the same handlers through mokapi-msw-adapter and validates against OpenAPI](/mokapi-storybook-addon-flow.png "The toolbar decides who answers the request")

The addon does not start MSW or Mokapi and does not import your mock modules. You keep your MSW worker setup and
start Mokapi yourself.

---

## What You Need

| Requirement | Version                                                           |
|-------------|-------------------------------------------------------------------|
| Mokapi      | 0.56.0 or newer (with mokapi-msw-adapter)                         |
| Storybook   | 10.6.4 or newer 10.x, Vite builder, CSF3 stories                  |
| Vite        | 6 or newer                                                        |
| MSW         | 2.15.0 or newer 2.x                                               |
| React       | 18 or newer (used for the addon panel; React projects are tested) |
| Node.js     | 22.12.0 or newer                                                  |

---

## Set It Up

The example mocks one endpoint, `GET /api/profile`. Your handler module, `mokapi.yaml` (with the `msw` alias)
and OpenAPI file are exactly those from [Mokapi Meets MSW](/resources/blogs/mokapi-meets-msw), so they are not repeated in full here.
For the Storybook example we add one scenario that breaks the contract on purpose:

```js
// src/mocks/profile.js
import { http, HttpResponse } from 'msw';
import { register } from 'mokapi-msw-adapter/register';

export const scenarios = {
  success: [http.get('/api/profile', () =>
    HttpResponse.json({ name: 'Jane Doe', email: 'jane.doe@example.com' }),
  )],
  error: [http.get('/api/profile', () =>
    HttpResponse.json({ message: 'Internal error' }, { status: 500 }),
  )],
  invalid: [http.get('/api/profile', () =>
    HttpResponse.json({ name: 'Jane Doe' }), // email is required by the contract
  )],
};

export default function () {
  register(scenarios, { api: 'Profile API' });
}
```

`api` is the `info.title` of your OpenAPI contract (`Profile API` here). Keep that value in mind, because the addon needs it, too.

### 1. Install

```sh
npm install --save-dev mokapi-storybook-addon mokapi-msw-adapter msw msw-storybook-addon @storybook/addon-docs go-mokapi
npx msw init public
```

`go-mokapi` installs the Mokapi binary through npm. Skip it if you already have Mokapi.
`npx msw init public` generates MSW's `mockServiceWorker.js`, which Storybook has to serve.
`msw-storybook-addon` is optional: any loader that reads `parameters.msw` works. `@storybook/addon-docs` provides the Docs pages that `tags: ['autodocs']` generates.

### 2. Register the Addon

```ts
// .storybook/main.ts
export default {
  framework: '@storybook/react-vite',
  stories: ['../src/**/*.stories.tsx'],
  staticDirs: ['../public'],
  addons: [
    '@storybook/addon-docs',
    {
      name: 'mokapi-storybook-addon',
      options: {
        target: 'http://localhost:3000', // the server URL of your API in Mokapi
      },
    },
  ],
};
```

`target` is required. Use the server URL from your OpenAPI contract, because Mokapi takes the port of the mocked API
from there. The addon adds the toolbar switch, the **Mokapi** panel and the Vite proxy from `/mokapi` to `target`. The
panel reads requests from Mokapi's dashboard at `http://localhost:8080`. If your dashboard runs elsewhere, set `dashboardUrl`.

### 3. Wrap Your MSW Loader

```ts
// .storybook/preview.ts
import { setupWorker } from 'msw/browser';
import { mswLoader } from 'msw-storybook-addon/csf3';
import { createMokapiLoader } from 'mokapi-storybook-addon/msw';

const loadMsw = mswLoader(async () => {
  const worker = setupWorker();
  await worker.start({ onUnhandledRequest: 'bypass', quiet: true });
  return worker;
});

export default {
  loaders: [createMokapiLoader(loadMsw)],
  parameters: { docs: { story: { inline: false, iframeHeight: '300px' } } },
};
```

``` box=info title="Note"
Register only the wrapped loader. If you also register <code>loadMsw</code> directly, both reset the MSW handlers.
```

Your loader keeps creating, starting and resetting the worker. `createMokapiLoader` only swaps the handlers before it runs.

### 4. Write the Stories

The component is a small profile that fetches `/api/profile` and prints the name, or the HTTP status if the request fails:

```tsx
// src/Profile.tsx
import { useEffect, useState } from 'react';

export function Profile() {
  const [text, setText] = useState('Loading...');

  useEffect(() => {
    void fetch('/api/profile').then(async (response) => {
      setText(response.ok ? (await response.json()).name : `HTTP ${response.status}`);
    });
  }, []);

  return <p>{text}</p>;
}
```


```tsx
// src/Profile.stories.tsx
import { createStoryScenarios } from 'mokapi-storybook-addon/msw';
import { Profile } from './Profile';
import { scenarios as profileScenarios } from './mocks/profile';

const scenarios = createStoryScenarios(profileScenarios, { api: 'Profile API' });

export default { title: 'Profile', component: Profile, tags: ['autodocs'] };

export const Success = { parameters: scenarios.success };
export const ServerError = { parameters: scenarios.error };
export const Invalid = { parameters: scenarios.invalid };
```

`createStoryScenarios` turns every named handler list into story parameters: `parameters.msw` with your handlers and
`parameters.mokapi` with the scenario name, the API and the routes. It reads method and path from the handlers and
does not run the resolvers. Scenario names stay typed, so `scenarios.sucess` is a TypeScript error.

### 5. Start Mokapi and Storybook

```sh
npx mokapi
npx storybook dev
```

Wait until Mokapi logs that it executed your script, then open a story.

---

## See It Fail: The Mokapi Panel

Open the `Invalid` story. With the toolbar on **MSW (browser)**, the component renders happily. Switch it to
**Mokapi (OpenAPI validation)**. The story reloads and the request goes to Mokapi, which answers with HTTP 500 because
`email` is missing. The component shows exactly what a real client would see.

![Storybook showing the Invalid story with the toolbar set to Mokapi (OpenAPI validation)](/mokapi-storybook-addon-toolbar.png "Switch the toolbar and the same story talks to Mokapi")

Open the **Mokapi** panel to see why. For each request it shows:

- status, method, path, duration and time
- the response body, which contains the validation error, such as `required properties are missing: email`
- the request and response metadata
- the executed scripts with console logs, parameters and script errors

![The Mokapi panel listing a request with status 500 and the response body with the validation error that email is missing](/mokapi-storybook-addon-panel.png "The panel shows the contract error next to the story")

The panel separates two kinds of errors. A response with a status of 400 or higher gets a note saying that a mocked error
status is not necessarily a contract violation. The `ServerError` story returns an intended 500 and is fine; `Invalid` returns a 500
only because the schema check failed. Script errors are shown on their own under the executed script.

How the panel works:

- It loads requests from Mokapi's dashboard API (`/api/events`) for the selected operation. If the story has several routes, pick one in the **Operation** select.
- It refreshes every two seconds while it is open in Mokapi mode. **Refresh** reloads immediately.
- Pages hold 25 requests and are filtered by scenario after loading, so a page can be empty while another page contains matches.
- It shows the operation's history for the scenario, **not** one story execution. Requests without the scenario parameter are not shown.
- Errors, such as an unreachable dashboard, appear in the panel and never block the story.
- It needs `api` in `createStoryScenarios`. Without it, the panel tells you to configure it.

---

## Good to Know

### Scenarios Work Per Request

The scenario travels with each request, not in a global switch. On a Docs page, every story can show a different scenario at the
same time, if you render stories in their own iframes (`docs.story.inline: false`, as in the `preview.ts` above). Inline stories share one MSW worker and
cannot show different scenarios. When you change the toolbar, the addon reloads all preview iframes, including those on a Docs page.

### Keep Some Handlers Local

By default, Mokapi mode replaces all of a story's handlers. If some should stay in the browser, for example for
authentication, put the Mokapi-managed ones in an MSW handler group:

```ts
const scenarios = createStoryScenarios(profileScenarios, {
  api: 'Profile API',
  handlerGroup: 'profile',
});

export const Success = {
  parameters: {
    ...scenarios.success,
    msw: { handlers: { ...scenarios.success.msw.handlers, auth: authHandlers } },
  },
};
```

`authHandlers` is your own `HttpHandler[]`. In Mokapi mode, only the `profile` group is replaced. The loader throws if `parameters.msw.handlers` has no group with that name.

### Configuration

| Option              | Default                   | Description                                                                         |
|---------------------|---------------------------|-------------------------------------------------------------------------------------|
| `target`            | required                  | Mocked API in Mokapi, target of the proxy                                           |
| `dashboardUrl`      | `'http://localhost:8080'` | Mokapi dashboard used by the panel                                                  |
| `proxyPath`         | `'/mokapi'`               | Local proxy path                                                                    |
| `defaultBackend`    | `'msw'`                   | Initial toolbar selection, `'msw'` or `'mokapi'`                                    |
| `scenarioParameter` | `'__mockScenario'`        | Query parameter that selects the scenario                                           |
| `browserUrl`        | not set                   | Mokapi address the browser calls directly instead of the proxy                      |

If you rename `scenarioParameter`, use the same name in `register(scenarios, { scenarioParameter })`. The adapter reads it on the Mokapi side.

### Errors Are Never Hidden

If Mokapi can't be reached, the addon logs the error and the story gets an HTTP 502. It never falls back to the local mocks,
because that would defeat the purpose. Invalid options or story parameters throw an error that starts with `Mokapi:`.

### Static Builds

`storybook build` has no Vite server, so the `/mokapi` proxy does not exist. Either provide a reverse proxy for `/mokapi`
(to `target`) and `/mokapi-dashboard` (to `dashboardUrl`), or set `browserUrl` at build time and enable CORS in Mokapi.

``` box=info title="Note"
The dashboard API returns request history, which can include headers, bodies and script data. Expose it only to
trusted users, and never put credentials into addon options: they are visible in the browser.
```

---

## Limitations

| Limitation                                  | Detail                                                                                  |
|---------------------------------------------|-----------------------------------------------------------------------------------------|
| Exact relative paths only                   | No `/users/:id` or `/files/*`, no query strings in routes, no absolute URLs             |
| JSON responses                              | mokapi-msw-adapter supports `HttpResponse.json()` only                                  |
| HTTP handlers only                          | Keep GraphQL or WebSocket handlers local with a handler group                           |
| Storybook with Vite and CSF3                | Webpack and CSF Next are not supported, only React is tested                            |
| Same origin                                 | Preview iframes must have the same origin as the Storybook manager                      |

See the section "What Is Supported" in [Mokapi Meets MSW](/resources/blogs/mokapi-meets-msw) for the handler subset the adapter supports.
Check your existing mocks against it before you switch a story.

---

## Beyond Storybook

Because Mokapi runs as a real HTTP server, the same mocks serve more than Storybook. Playwright tests, a teammate
with `curl` or a service under development use them as well, and you can inspect every request in the Mokapi
dashboard. See [End to End Testing with Mocked APIs](/resources/blogs/end-to-end-testing-with-mocked-apis). If you need more than the adapter
supports, Mokapi's own [JavaScript API](/resources/blogs/bring-your-mock-apis-to-life-with-javascript) gives you full control.

---

## Conclusion

If your stories already use MSW, you don't need a second set of mocks to check them against your API. Keep your handlers, add
the addon, and one toolbar switch tells you whether a story's data still matches the OpenAPI contract. The result appears next to
your component, in the **Mokapi** panel.

Both projects are open source (MIT): [mokapi-storybook-addon](https://github.com/marle3003/mokapi-storybook-addon)
([npm](https://www.npmjs.com/package/mokapi-storybook-addon)) and
[mokapi-msw-adapter](https://github.com/marle3003/mokapi-msw-adapter)
([npm](https://www.npmjs.com/package/mokapi-msw-adapter)). New to Mokapi? Start with the
[installation guide](/docs/get-started/installation).
