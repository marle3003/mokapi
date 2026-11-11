---
title: "Storybook Stories That Check Your OpenAPI Contract with Mokapi"
description: Validate the MSW responses of your Storybook stories against your OpenAPI contract with mokapi-storybook-addon. Switch validation on with a toolbar toggle and see contract errors in a panel.
subtitle: Your stories already run on MSW handlers. Add one toolbar toggle, and Mokapi checks every mocked response against your OpenAPI contract. You see in the Storybook panel when a mock no longer matches the API.
image:
  url: "/mokapi-storybook-addon-panel.png"
  alt: Storybook with the Validate with Mokapi toggle switched on and the Mokapi panel showing one invalid response, because the required property email is missing.
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

# Storybook Stories That Check Your OpenAPI Contract with Mokapi

## Green Stories, Broken Contract

Storybook lets you build and review UI components without a backend. Most teams fill the gap with
[MSW](https://mswjs.io) handlers: each story gets a handler that returns the JSON the component needs.

The handlers are handwritten, and nothing compares them with the OpenAPI specification. The backend team marks `email` as
required, a story mock forgets it, and the story stays green. The component renders fine against a response the real
API would never send. You find out in an integration environment, or in production.

[mokapi-storybook-addon](https://github.com/marle3003/mokapi-storybook-addon) closes that gap. This article shows
**how to switch contract validation on with a toolbar button, and how to see the result next to your component.**

---

## The Idea: Keep MSW, Let Mokapi Check the Result

Your stories keep running on MSW. The component gets the response from your handler, unchanged and without delay. The
addon listens to the MSW worker's `response:mocked` event, copies the request and the response, and sends the copy to
Mokapi's dashboard API (`POST /api/services/http/validate`). Mokapi looks up the operation in your OpenAPI contract and
checks both request and response. The result shows up in the **Mokapi** panel.

![Diagram: MSW answers the story's request in the browser, the addon copies request and response to Mokapi's validate endpoint, and the Mokapi panel shows the result](/mokapi-storybook-addon-flow.png "MSW answers, Mokapi checks")

Nothing Mokapi-specific goes into your stories or into `.storybook/preview.ts`. The addon does not start MSW or Mokapi and
does not import anything from `msw`. You keep your MSW worker setup and start Mokapi yourself.

Only responses from your handlers are validated. Requests that you pass through with `passthrough()` or that no handler
matches are not sent to Mokapi.

---

## What You Need

| Requirement | Version                                                           |
|-------------|-------------------------------------------------------------------|
| Mokapi      | 0.58.0 or newer, with your OpenAPI contract loaded                |
| Storybook   | 10.6.4 or newer 10.x, Vite builder, CSF3 stories                  |
| Vite        | 6 or newer                                                        |
| MSW         | 2.15.0 or newer 2.x                                               |
| React       | 18 or newer (used for the addon panel; React projects are tested) |
| Node.js     | 22.12.0 or newer                                                  |

Older Mokapi versions have no validation endpoint. The panel then tells you to update Mokapi.

---

## Set It Up

The example validates one endpoint, `GET /api/profile`, of an API whose OpenAPI `info.title` is `Profile API`.
Its contract marks `name` and `email` as required.

### 1. Install

```sh
npm install --save-dev mokapi-storybook-addon msw msw-storybook-addon @storybook/addon-docs go-mokapi
npx msw init public
```

`go-mokapi` installs the Mokapi binary through npm. Skip it if you already have Mokapi.
`npx msw init public` generates MSW's `mockServiceWorker.js`, which Storybook has to serve.
`msw-storybook-addon` is optional: any loader that sets `context.msw` works. `@storybook/addon-docs` provides the Docs pages that `tags: ['autodocs']` generates.

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
        dashboardUrl: 'http://localhost:8080', // default
        defaultValidation: false,              // default
      },
    },
  ],
};
```

Both options are optional. `dashboardUrl` is the address of the Mokapi dashboard, which must be reachable from the browser.
`defaultValidation` sets the initial state of the toolbar toggle.

### 3. Keep Your MSW Loader

Your `.storybook/preview.ts` needs no addon code:

```ts
// .storybook/preview.ts
import { setupWorker } from 'msw/browser';
import { mswLoader } from 'msw-storybook-addon/csf3';

export default {
  loaders: [
    mswLoader(async () => {
      const worker = setupWorker();
      await worker.start({ onUnhandledRequest: 'bypass', quiet: true });
      return worker;
    }),
  ],
  parameters: { docs: { story: { inline: false, iframeHeight: '300px' } } },
};
```

If you use your own loader instead of `mswLoader`, it must set `context.msw` to the worker that answers your requests.
The addon reads the worker from there.

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

The stories stay plain MSW. The `Invalid` story breaks the contract on purpose:

```tsx
// src/Profile.stories.tsx
import { http, HttpResponse } from 'msw';
import { Profile } from './Profile';

export default { title: 'Profile', component: Profile, tags: ['autodocs'] };

export const Valid = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/profile', () =>
          HttpResponse.json({ name: 'Jane Doe', email: 'jane.doe@example.com' }),
        ),
      ],
    },
  },
};

export const Invalid = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/profile', () =>
          HttpResponse.json({ name: 'Jane Doe' }), // email is required by the contract
        ),
      ],
    },
  },
};
```

### 5. Start Mokapi and Storybook

Start Mokapi with your OpenAPI contract, then Storybook:

```sh
npx mokapi ./openapi.json
npx storybook dev
```

---

## See It Fail: The Toggle and the Mokapi Panel

Open the `Invalid` story. Validation is off, so the component renders happily and nothing is sent to Mokapi. Click the
**Validate with Mokapi** icon button in the toolbar. The button is highlighted while validation is on, and its tooltip shows
the state. Switching does not reload the preview. The addon remounts the current story, so a component that fetches on
mount requests its data again, and that request is validated right away.

The toggle sets the global `mokapiValidation` to `on` or `off`. It only appears in story view, not on Docs pages. To
start with validation on, set `defaultValidation: true` or open a story with `globals=mokapiValidation:on` in the URL.

![Storybook toolbar with the highlighted Validate with Mokapi toggle and its tooltip "Mokapi validation is on. Click to turn off." while the Invalid story is open, the Mokapi tab shows badge 1](/mokapi-storybook-addon-toolbar.png "The toggle is highlighted while validation is on")

Open the **Mokapi** panel to see the result. The panel lists the responses of the selected story since its last render.
Switching to another story or rendering the story again clears the list. Each entry shows the method, URL, HTTP status and
time, and one of these results:

| Result               | Meaning                                                                                  |
|----------------------|------------------------------------------------------------------------------------------|
| valid                | Request and response fit the contract                                                    |
| invalid              | The contract is violated. The entry lists each error with its location, for example `response.body`, and message |
| not in contract      | Mokapi found no operation for the request, or several APIs match                         |
| Mokapi not reachable | The request to Mokapi failed. The entry shows the URL and a hint to `dashboardUrl`       |

For the `Invalid` story, the entry is invalid and the error points to `response.body` with the missing property `email`.
The tab title shows a badge with the number of invalid responses. Responses that are not in the contract and
not-reachable entries are not counted.

![The Validate with Mokapi toggle in the toolbar and the Mokapi panel with badge 1, listing GET /api/profile as invalid because email is missing](/mokapi-storybook-addon-panel.png "The panel shows the contract error next to the story")

For valid and invalid entries, the request links to the page of the matched operation in the Mokapi dashboard and opens in
a new tab. Hover the link to see the operation from the contract, for example `GET /api/users/{id}` for a request to `/api/users/42`. The panel also links to the Mokapi
dashboard itself.

Notices appear at the top of the panel when the addon can't validate: Mokapi is not reachable, your Mokapi is older than
0.58.0, Mokapi reports a configuration error such as an unknown `api`, or the MSW worker is missing from `context.msw`.
They never block the story, unless you configure test mode `'error'` (see below).

---

## Choosing the API

Mokapi can serve several APIs. Without further configuration, Mokapi searches all of them for an operation that matches
the request's method and path. If more than one API matches, Mokapi can't choose, and the panel shows the request as
**not in contract** with the names of the matching APIs. Set the API's `info.title` to fix that:

```ts
// .storybook/preview.ts, a story file or a single story
export default {
  parameters: {
    mokapi: { api: 'Profile API' },
  },
};
```

`parameters.mokapi` is inherited from preview, component and story, like other Storybook parameters. A story can overwrite
the value of the project. If `api` names an API that Mokapi doesn't know, the panel shows Mokapi's message.

A story that has no contract, for example a mock for a third-party API, can opt out. The addon then sends nothing to Mokapi,
also when the toggle is on:

```ts
export const ThirdParty = {
  parameters: {
    mokapi: { disable: true },
    msw: { handlers: [/* ... */] },
  },
};
```

---

## Failing Stories in CI

By default, contract errors only show up in the panel. In CI nobody looks at the panel, so a story can fail instead. The
test mode decides what happens when a story has a failing result:

| Mode               | Story                                                                       |
|--------------------|-----------------------------------------------------------------------------|
| `'error'`          | Storybook shows its error display and finishes the story with status `error` |
| `'todo'` (default) | Contract errors are reported as a warning, the story stays green            |
| `'off'`            | Contract errors are shown in the panel only, the story stays green          |

Set the default for all stories in `.storybook/main.ts`, and override it with `parameters.mokapi.test` for a story, a
component or the whole project:

```ts
// .storybook/main.ts
export default {
  addons: [
    {
      name: 'mokapi-storybook-addon',
      options: {
        defaultValidation: true,
        defaultTest: 'error',
      },
    },
  ],
};
```

```ts
export const Legacy = {
  parameters: {
    mokapi: { test: 'todo' },
  },
};
```

Test mode only applies while validation is on, it never turns validation on by itself. While validation is on, the panel
shows the test mode of the selected story, so you can see what a CI run does with it. You can't change the mode in the
panel.

A failing result is an invalid response, an unreachable Mokapi (or one without the validation endpoint), a
configuration error such as an unknown `api`, or a validation that did not finish within 4 seconds. Responses that are
not in the contract and a missing MSW worker only show in the panel, so a project with partial contracts doesn't get a
warning on every story. A story without requests passes.

Before it reports, the story waits for the requests that its handlers answered and for their validation. In `'error'` mode
the story throws an error like this:

```
Mokapi: 1 response does not match the contract.

GET /api/profile -> GET /api/profile (http://localhost:8080/dashboard/http/services/Profile%20API/api/profile/get)
  response.body: required properties are missing: email
```

### Test with Playwright

Open the story with validation on and check Storybook's error state. Storybook adds the class `sb-show-errordisplay` to the
body of the preview iframe:

```ts
import { test, expect } from '@playwright/test';

test('a drifted mock fails the story', async ({ page }) => {
  await page.goto('http://localhost:6006/iframe.html?id=profile--invalid&viewMode=story&globals=mokapiValidation:on');

  await expect(page.locator('body.sb-show-errordisplay')).toBeVisible();
  await expect(page.locator('#error-message, #error-stack')).toContainText(
    'Mokapi: 1 response does not match the contract.',
  );
});
```

This test expects the `Invalid` story under `defaultTest: 'error'`. With `@storybook/addon-vitest`, a story in `'error'`
mode fails its test with the same message. Set `defaultValidation: true` and make sure Mokapi runs with your contract
while the tests run.

---

## Good to Know

### Docs Pages

Render stories on a Docs page in their own iframes (`docs.story.inline: false`, as in the `preview.ts` above). Each story
then runs with its own MSW worker. Inline stories share one worker, so the addon can't attribute their results to a
story.

### Static Builds

The browser calls Mokapi directly, and the addon sets up no proxy. `storybook dev` and `storybook build` use the same
`dashboardUrl`, so it must be an address that the browser of every viewer can reach. Mokapi's dashboard API allows
requests from any origin, so you don't need CORS setup.

A Storybook served over HTTPS can't call Mokapi on a plain `http://` address other than `localhost`, because browsers block
mixed content. Put Mokapi behind an HTTPS reverse proxy and use its address as `dashboardUrl`.

``` box=info title="Note"
The dashboard API returns request history, which can include headers, bodies and script data. Expose it only to
trusted users, and never put credentials into `dashboardUrl`: it is visible in the browser.
```

### Configuration

| Option              | Default                   | Description                                                              |
|---------------------|---------------------------|--------------------------------------------------------------------------|
| `dashboardUrl`      | `'http://localhost:8080'` | Mokapi dashboard URL, must be reachable from the browser                 |
| `defaultValidation` | `false`                   | Initial state of the toolbar toggle                                      |
| `defaultTest`       | `'todo'`                  | Test mode of all stories: `'error'`, `'todo'` or `'off'`                 |

Story parameters under `parameters.mokapi`:

| Field     | Description                                                          |
|-----------|----------------------------------------------------------------------|
| `api`     | OpenAPI `info.title` of the story's API. Optional                    |
| `disable` | Sends nothing to Mokapi for this story                               |
| `test`    | Overrides `defaultTest` for this story                               |

Invalid options throw an error that starts with `Mokapi:` when Storybook starts.

---

## Limitations

| Limitation                       | Detail                                                                                          |
|----------------------------------|-------------------------------------------------------------------------------------------------|
| Late requests                    | In test mode, requests that start after a short quiet window (100 ms, at most 1 s) are not counted, but they still appear in the panel |
| Fixed timeout                    | A validation must finish within 4 s. You can't configure it                                     |
| Thrown handlers                  | If a handler throws, MSW answers with its own 500, which the addon doesn't validate             |
| Request and response only        | Mokapi checks the pair against the contract. It does not run your handlers                      |
| Storybook with Vite and CSF3     | Webpack and CSF Next are not supported, only React is tested                                    |

---

## Beyond Storybook

If you also want to serve your MSW handlers over real HTTP, for Playwright tests or a service under development, see
[Mokapi Meets MSW](/resources/blogs/mokapi-meets-msw). That is a separate setup with
[mokapi-msw-adapter](https://github.com/marle3003/mokapi-msw-adapter), and the Storybook addon doesn't need it. For
end-to-end tests, see [End to End Testing with Mocked APIs](/resources/blogs/end-to-end-testing-with-mocked-apis). If you
need more than the adapter supports, Mokapi's own
[JavaScript API](/resources/blogs/bring-your-mock-apis-to-life-with-javascript) gives you full control.

---

## Conclusion

If your stories already use MSW, you don't need a second set of mocks to check them against your API. Keep your handlers and
add the addon. With `defaultTest: 'error'`, drift between a mock and the contract fails the story in CI, before anyone has to
open the panel.

Both projects are open source (MIT): [mokapi-storybook-addon](https://github.com/marle3003/mokapi-storybook-addon)
([npm](https://www.npmjs.com/package/mokapi-storybook-addon)) and
[mokapi-msw-adapter](https://github.com/marle3003/mokapi-msw-adapter)
([npm](https://www.npmjs.com/package/mokapi-msw-adapter)). New to Mokapi? Start with the
[installation guide](/docs/get-started/installation).
