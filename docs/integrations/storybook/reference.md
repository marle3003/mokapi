---
title: Storybook Addon Reference
description: Options, parameters, global, error messages, limitations and deployment notes of the Mokapi Storybook addon.
---

# Storybook Addon Reference

## Addon options

Set in `.storybook/main.ts`. Restart Storybook after changing them.

| Option              | Type      | Default                   | Description                                         |
|---------------------|-----------|---------------------------|-----------------------------------------------------|
| `dashboardUrl`      | `string`  | `'http://localhost:8080'` | Mokapi dashboard URL, must be reachable from the browser. A trailing slash is removed |
| `defaultValidation` | `boolean` | `false`                   | Initial state of the toolbar toggle. `true` starts with validation on and the button highlighted |
| `defaultTest`       | `'error' \| 'warning' \| 'off'` | `'warning'` | Test mode of all stories, see [Failing stories](/docs/integrations/storybook/failing-stories.md) |
| `defaultPartial`    | `boolean` | `false`                   | Validate the responses of all stories as partial responses, see [Partial responses](/docs/integrations/storybook/usage.md#partial-responses). Needs Mokapi 0.60.0 |
| `defaultTimeout`    | `number`  | `4000`                    | Milliseconds a story waits for its responses and their validation, quiet window included. Must be a finite number greater than 0, see [Wait limit](/docs/integrations/storybook/failing-stories.md#wait-limit) |

## `parameters.mokapi`

| Field     | Type      | Description                                                                 |
|-----------|-----------|-----------------------------------------------------------------------------|
| `api`     | `string`  | OpenAPI `info.title` of the story's API. Optional, see [Choosing the API](/docs/integrations/storybook/usage.md#choosing-the-api) |
| `disable` | `boolean` | Sends nothing to Mokapi for this story                                      |
| `test`    | `'error' \| 'warning' \| 'off'` | Overrides `defaultTest` for this story, see [Failing stories](/docs/integrations/storybook/failing-stories.md) |
| `partial` | `boolean` | Overrides `defaultPartial` for this story, see [Partial responses](/docs/integrations/storybook/usage.md#partial-responses) |
| `timeout` | `number`  | Overrides `defaultTimeout` for this story, in milliseconds, see [Wait limit](/docs/integrations/storybook/failing-stories.md#wait-limit) |

## Global `mokapiValidation`

| Value   | Description                                     |
|---------|-------------------------------------------------|
| `'on'`  | Validate the responses of the mocked requests   |
| `'off'` | Validate nothing (default)                      |

The toolbar button switches the global of the current preview only. You can set
the initial value with `defaultValidation`, with the URL parameter
`globals=mokapiValidation:on`, or for component tests with
`storybookTest({ initialGlobals: { mokapiValidation: 'on' } })`, see
[Vitest](/docs/integrations/storybook/failing-stories.md#vitest).

## TypeScript

Type definitions are included. Importing `mokapi-storybook-addon` adds
`mokapi` to Storybook's `Parameters` type. The package also exports the type
`TestMode`.


## Errors

Invalid options throw an error when Storybook starts:

- `Mokapi: dashboardUrl must be an HTTP(S) URL.`
- `Mokapi: defaultValidation must be true or false.`
- `Mokapi: defaultTest must be 'error', 'warning' or 'off'.`
- `Mokapi: defaultPartial must be true or false.`
- `Mokapi: defaultTimeout must be a positive number.`

The panel shows these notices. None of them counts in the badge. They don't stop
the story either, unless the test mode is `'error'` and the notice is a failing
result, see [Failing stories](/docs/integrations/storybook/failing-stories.md):

| Notice | Cause and fix |
|--------|---------------|
| `Mokapi is not reachable at <url>. Check the dashboardUrl option and that Mokapi is running.` | The request failed. Start Mokapi or fix `dashboardUrl` |
| `Mokapi at <url> does not support /api/services/http/validate. Update Mokapi to 0.58.0.` | Your Mokapi is older than 0.58.0 |
| Mokapi's message, for example for an unknown API | `parameters.mokapi.api` names an API that Mokapi doesn't serve. Use the `info.title` of your contract |
| `Mokapi answered HTTP <status>: <message>` | Any other unexpected answer from Mokapi. Mokapi's message is added if it sent one |
| `Mokapi: parameters.mokapi.test must be 'error', 'warning' or 'off'.` | A story, component or preview sets an invalid test mode. The story uses `defaultTest` |
| `Mokapi: parameters.mokapi.partial must be true or false.` | A story, component or preview sets a `partial` that isn't a boolean. Counts as a configuration error. The story uses `defaultPartial` |
| `Mokapi: parameters.mokapi.timeout must be a positive number.` | A story, component or preview sets a `timeout` that isn't a finite number greater than 0. Counts as a configuration error. The story uses `defaultTimeout` |
| `The Mokapi at <dashboardUrl> doesn't support partial responses and must be updated to 0.60.0.` | A story with partial responses met a Mokapi older than 0.60.0. It ignores the option and validates strictly. Update Mokapi or set `partial: false` |
| `Mokapi: could not read the mocked request or response: <error message>` | The addon could not copy the mocked request or response, or could not show the result. Nothing was validated for that request. The error message tells why |
| `Mokapi validation did not finish within <s> s (<method> <path>). Increase parameters.mokapi.timeout or defaultTimeout.` | MSW answered, but Mokapi did not answer in time. Failing result. One notice per unfinished request. Increase the limit or check that Mokapi is responsive |
| `No response from the MSW handler within <s> s (<method> <path>), not validated. If the response is delayed on purpose, increase parameters.mokapi.timeout or defaultTimeout.` | The handler did not answer within the limit. Not a failing result. Increase the limit if the delay is intended |
| ``MSW worker not found: `context.msw` is missing. Use `mswLoader` or set `context.msw` in your loader.`` | The story has no MSW worker. Use `mswLoader` or set `context.msw` in your own loader |

In test mode `'error'`, the story fails with the message described in
[Error message](/docs/integrations/storybook/failing-stories.md#error-message): `Mokapi contract validation failed: ` and the
headline `<N> response(s) do(es) not match the contract.`, followed by one line
per invalid response with its errors, and the notice texts of the other failing
results.

The notice about a missing worker only appears while validation is on. With
validation off, a story without a worker runs without any message.

## Limitations

- Inline Docs stories share one MSW worker, so their results can't be attributed
  to a story. Use `inline: false`, see [Docs pages](/docs/integrations/storybook/usage.md#docs-pages). If you keep inline
  stories that share one preview, results may be attributed to the wrong story.
- The test mode waits until no new request has started for 100 ms (at most 1 s),
  then for the requests that started before. A request that starts later, for
  example from polling or a timer, is not counted. It still appears in the panel.
- The wait limit (4 s by default, see [Wait limit](/docs/integrations/storybook/failing-stories.md#wait-limit)) includes that
  quiet window. A limit below the quiet window still waits for it.
- Opaque responses emit no MSW event, so a request that gets one only ends at the
  limit, as a `handler-timeout` notice that doesn't fail the story.
- A story with a loading skeleton, whose handler never answers, waits the whole
  limit before it ends as `SKIP`. Set `disable: true` or a small `timeout` for it.
- A response that MSW delays beyond the limit isn't validated, so an invalid
  response stays green. Raise `timeout` for such a story.
- A response that arrives after the report appears in the panel, but doesn't
  change the badge, the headline or the report.
- The Mokapi `afterEach` runs after the `afterEach` hooks of other addons and
  after the one in your own `.storybook/preview.ts`, because the addon puts its
  preview annotations first. Another addon that does the same can still run
  after it.
- A test run result of a story in `'error'` mode has no entries, and results of test runs are gone after a browser reload, see [Results of Run tests](/docs/integrations/storybook/failing-stories.md#results-of-run-tests).
- The **Contract** row of the context menu works only with `storybook dev`. Statuses of a static build are gone after a reload. With `storybook dev` they survive a reload and stay until the dev server stops, see [Sidebar status](/docs/integrations/storybook/failing-stories.md#sidebar-status).
- The addon registers a test provider, so the **Run tests** widget shows the row **Contract** in every project that uses the addon. Without `@storybook/addon-vitest`, **Run tests** does nothing for Mokapi, see
  [Run tests widget](/docs/integrations/storybook/failing-stories.md#run-tests-widget).
- The **Contract** row of the widget removes Storybook's divider with a CSS rule on `data-module-id`. A Storybook upgrade can bring the divider back.
- Only `@storybook/addon-vitest` (Vitest browser mode) fails a story on a
  contract error, see [Vitest](/docs/integrations/storybook/failing-stories.md#vitest).
- If a handler throws, MSW answers with its own 500. The addon doesn't validate
  that response.
- Only responses from MSW handlers are validated, see [How it works](/docs/integrations/storybook#how-it-works). `HttpResponse.error()` has no HTTP response to validate.
- The addon needs a worker (or server) with MSW's `events.on` in `context.msw`.
- Mokapi validates the request and response pair against the contract. It does
  not run your handlers.
- The panel shows a binary body as text, so it can look garbled.
- Storybook with Vite and CSF3 only. Webpack and CSF Next are not supported.
  Only React projects are tested.
- A Storybook served over HTTPS can't reach Mokapi on a plain `http://` address, see
  [Static builds and deployment](#static-builds-and-deployment).
- Requires Mokapi 0.58.0 or newer, and 0.60.0 for partial responses, see [Requirements](/docs/integrations/storybook#requirements).

## Static builds and deployment

The browser calls Mokapi directly. The addon sets up no proxy, so `storybook dev`
and `storybook build` use the same `dashboardUrl`. It must be an address that the
browser of every viewer can reach, not only the machine that builds Storybook.

Mokapi answers the request with `Access-Control-Allow-Origin: *`, so no CORS
setup is needed.

A Storybook served over HTTPS can't call Mokapi on a plain `http://` address other
than `localhost`: browsers block this as mixed content. Chrome and Firefox allow
`http://localhost`, but Chrome may ask for permission to access the local network,
and other browsers may still block it. Put Mokapi behind an HTTPS reverse proxy
and use its address as `dashboardUrl`, or run Storybook over HTTP while you
validate.

Only expose the dashboard API to trusted users. Don't put credentials in
`dashboardUrl`, because the option is visible in the browser.

