---
title: Failing Stories with Mokapi
description: Fail Storybook stories and CI runs when MSW mocks don't match your OpenAPI contract, with test modes, wait limit, sidebar status and Vitest.
---

# Failing Stories

By default, contract errors only show up in the panel. In CI nobody looks at the
panel, so a story can fail instead. The test mode decides what happens when a
story has a failing result:

| Mode               | Report status | Story                                                                 |
|--------------------|---------------|-----------------------------------------------------------------------|
| `'error'`          | `failed`      | In the Storybook UI the story finishes with status `error` and gets an error icon in the sidebar. Under `@storybook/addon-vitest` the test fails |
| `'warning'` (default) | `warning`  | The story stays green and gets a warning icon in the sidebar          |
| `'off'`            | no report     | The story stays green                                                 |

In the Storybook UI the addon doesn't throw an error for a failed story. There
is no error display, no entry under **Interactions** and no red error screen in
the preview. The story finishes with status `error` because of its `failed`
report, and the **Mokapi** panel shows the `FAIL` badge, the headline and the details.

Set the default for all stories in `.storybook/main.ts`:

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

`parameters.mokapi.test` overrides it for a story, a component or the whole
project:

```ts
export const Legacy = {
  parameters: {
    mokapi: { test: 'warning' },
  },
};
```

The test mode only applies while validation is on. It never turns validation on
by itself. In CI, set `defaultValidation: true`, set
`initialGlobals: { mokapiValidation: 'on' }` in `storybookTest` (see
[Vitest](#vitest)) or open the story with `globals=mokapiValidation:on`.

## What fails

A failing result is:

- an invalid response
- Mokapi is not reachable, or the Mokapi at `dashboardUrl` has no validation endpoint
- a story with partial responses and a Mokapi older than 0.60.0, which ignores the option
- a configuration error, for example an unknown `api`, an invalid `parameters.mokapi.test`, `parameters.mokapi.partial` or `parameters.mokapi.timeout`
- a validation that did not finish within the [wait limit](#wait-limit), although MSW had answered the request

These don't count, they only show in the panel: responses that are not in the
contract (a project with partial contracts would get a warning on every story),
a missing MSW worker and a handler that did not answer within the wait limit
(a loading skeleton delays on purpose). A story without requests passes.

Before it reports, the story waits for the requests that its handlers answer, and
for their validation. Requests that MSW passes through or that no handler
matches are dropped, so they don't block the story. The wait is capped at 4 s by
default, including the short quiet window described in
[Limitations](/docs/integrations/storybook/reference.md#limitations). If MSW had answered a request and its validation is
still open then, the story fails with
`Mokapi validation did not finish within 4 s (GET /api/profile/1). Increase parameters.mokapi.timeout or defaultTimeout.`,
one request per line if several.

## Wait limit

Two things can still be open when the limit is reached, and they end differently:

- MSW answered, Mokapi did not validate in time: the story fails, as described
  above.
- The handler did not answer in time: nothing was validated and nothing failed.
  The story passes, the panel shows `SKIP` with the notice
  `No response from the MSW handler within 4 s (GET /api/profile), not validated. If the response is delayed on purpose, increase parameters.mokapi.timeout or defaultTimeout.`

Change the limit for all stories with `defaultTimeout` and for a story, a
component or the whole project with `parameters.mokapi.timeout`, both in
milliseconds:

```ts
// .storybook/main.ts
export default {
  addons: [
    {
      name: 'mokapi-storybook-addon',
      options: {
        defaultTimeout: 4000, // default
      },
    },
  ],
};
```

```ts
// A story whose handler delays the response on purpose
export const SlowProfile = {
  parameters: {
    mokapi: { timeout: 10000 },
    msw: { handlers: [/* ... */] },
  },
};
```

The value must be a finite number greater than 0. The notices show it in
seconds, for example `1.5 s` for `1500`.

## Sidebar status

When a story with validation on has its report, the addon gives it a status in
Storybook's status store:

| Report | Sidebar status |
|--------|----------------|
| `failed` (mode `'error'`) | error |
| `warning` (mode `'warning'`) | warning |
| `passed` with at least one valid response | success |
| skipped (`passed`, but no response was checked), validation off | none, an earlier status stays |
| `disable: true`, mode `'off'` | none, an earlier status is removed when the story is mounted (validation on or off) |

The status has the title `Mokapi`. For an error or a warning, its description is
the headline of the [error message](#error-message), for example
`1 response does not match the contract.`. For a pass it is
`All responses match the contract.`. A story with a success status gets
Storybook's green icon in the sidebar and shows up in the filter
**Stories with passing tests**. A skipped story, one that passed without a single
response checked against the contract, gets no status, like the gray `SKIP` badge
in the panel.

A new result of the story replaces its status. A mount that checks no response
(skipped) neither replaces nor removes it: the last result stays, in the panel as
well, until a newer result exists, also after you switch to another story and
back. Mounting a story with `disable: true` or test mode `'off'` (also through
`defaultTest`) is different: it removes the remembered panel result and a test
run result as well. Switching the validation toggle removes nothing. A story that
has never checked a response has no status.
Switching to another story keeps the status. **Clear all statuses** in the
testing module removes all Mokapi statuses, the success statuses included.
Without a new report, a status stays when you switch validation off.

### Context menu

Open the context menu of a story, a component or a group in the sidebar. The row
**Contract** sits next to the reports of other test providers such as
Interactions and Accessibility. It is always there, so you see at a glance
whether Mokapi checked the story. A docs entry has no row.

| State | Dot | Number | Button |
|-------|-----|--------|--------|
| No result | gray | none | disabled, `Mokapi: No result` |
| Passed | green | none | `Mokapi: Passed` |
| Warning | yellow | invalid responses | `Mokapi: Warning (1 invalid response)` |
| Error | red | invalid responses | `Mokapi: Error (2 invalid responses)` |

The row says "No result" for a story without a Mokapi status. This covers a
skipped story, mode `'off'`, validation off and `disable: true`. The tooltip of a
story is `Validate with Mokapi or run tests to see results`,
`All responses match the contract.` or the headline of the error or warning.

A component or a group shows the worst status of the stories below it, in the
order error, warning, passed, no result. The number is the sum of their invalid
responses. Its tooltip is `Contract validation failed`,
`Contract validation has warnings`, `All responses match the contract.` or, without
a result, the same text as for a story.

A click on the button selects a story and opens the **Mokapi** panel. For a
component or a group that is the first story with an error, otherwise the first
with a warning, otherwise the first that passed. Without a result the button is
disabled.

**Run tests** sets Mokapi statuses as well, see [Vitest](#vitest). A test run
replaces the status of every story that took part in it. Stories outside the run
keep theirs. The status of a test run stays until you mount the story with
validation on, mount it with `disable: true` or test mode `'off'`, or the preview
reports a newer result.

### Run tests widget

The testing module below the sidebar has the same row **Contract**. It sits
directly below the rows of `@storybook/addon-vitest` (Interactions, Coverage
and, with `@storybook/addon-a11y`, Accessibility) without a divider or heading, so it reads
as one more row of that block.
The widget is part of the dev server only, like the context menu.

The checkbox shows whether **Run tests** validates contracts. It is checked when
`defaultValidation` is `true` and unchecked otherwise, and it is disabled, because
you can't switch it here. Hover it to see the text. Set `defaultValidation: true`
to have the tests check contracts, see [Vitest](#vitest).

The button has the same states, dot, number and texts as the
[context menu](#context-menu), but it covers all stories of the Storybook, as a
component or group row does: the worst status wins and the number is the sum of
the invalid responses. Its tooltips are always the fixed texts, such as
`Contract validation failed`, also when a single story fails. A click selects a
story and opens the **Mokapi** panel. The story is the first one with an error,
otherwise the first with a warning, otherwise the first that passed, in the order
of the sidebar.

The row updates when a test run ends and when you click **Clear all statuses**,
which turns it gray. It shows no spinner while a run is in progress and keeps the
previous result until the new one arrives.

Storybook draws a divider below each box of the testing module. The addon removes
the divider between addon-vitest's rows and **Contract** with a CSS rule on the
box's `data-module-id`. If the addon is registered before `@storybook/addon-vitest`
in `.storybook/main.ts`, Mokapi's box comes first and Storybook's divider stays.

## Error message

The error message looks like this. Under Vitest a failed story throws it as a
`MokapiContractError`. The text after `Mokapi contract validation failed: ` is
the headline. The panel and the description of the sidebar status show the
headline alone:

```
MokapiContractError: Mokapi contract validation failed: 1 response does not match the contract.

GET /api/profile/1 -> GET /profile/{id} (http://localhost:8080/dashboard/http/services/Profile%20API/profile/%7Bid%7D/get)
  response.body: missing required property 'name'
```

With several invalid responses, the headline reads
`N responses do not match the contract.` The headline counts invalid
responses only. Other failing results (not reachable, endpoint missing,
configuration error, timeout) follow as lines with the same text as in the
panel. Without an invalid response, the headline is the text of the first
failing notice, for example
`Mokapi is not reachable at http://localhost:8080. Check the dashboardUrl option and that Mokapi is running.`
and the details don't repeat it.

## Test with Playwright

The Storybook UI shows no error display for a contract error, so check the
`storyFinished` event that Storybook sends on its channel. Its payload has
`status: 'error'` for a failed story, and `reporters` contains the report of type
`mokapi` with the status `failed`:

```ts
import { test, expect } from '@playwright/test';

test('a drifted mock fails the story', async ({ page }) => {
  await page.addInitScript(() => {
    window.__storyFinished = [];
    let channel;
    Object.defineProperty(window, '__STORYBOOK_ADDONS_CHANNEL__', {
      get: () => channel,
      set(next) {
        next?.on('storyFinished', (event) => window.__storyFinished.push(event));
        channel = next;
      },
    });
  });
  await page.goto('http://localhost:6006/iframe.html?id=profile--invalid&viewMode=story&globals=mokapiValidation:on');

  await expect
    .poll(() => page.evaluate(() => window.__storyFinished.at(-1)?.status))
    .toBe('error');
});
```

The story is a story with an invalid mock under `defaultTest: 'error'`. Adjust
the story id to your project. For a CI run, `@storybook/addon-vitest` is the
simpler way, see below.

## Vitest

With `@storybook/addon-vitest`, a story in `'error'` mode fails its test with a
`MokapiContractError`, see [Error message](#error-message). This applies to **Run
tests** in the Storybook UI and to `vitest run`. The addon throws only when it
detects a Vitest browser run, because addon-vitest fails a test only on a thrown
error and ignores a `failed` report. Make sure that Mokapi runs with your
contract while Vitest runs. Stories in `'warning'` and `'off'` pass.

Component tests don't take the toolbar state from the Storybook UI. The toolbar
button changes the globals of the preview you're looking at, but **Run tests**
and `vitest run` start every story in their own Vitest context, with the initial
globals. So if you turn validation on in the toolbar while `defaultValidation`
is `false`, the panel shows the contract errors and the component test still
passes, because nothing was validated in it.

To validate in the tests, set `defaultValidation: true`, or turn validation on
for the Vitest project only with `initialGlobals` of `storybookTest`. Then the
Storybook UI starts with validation off and the tests always validate:

```ts
// vitest.config.ts
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';
import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';

export default defineConfig({
  test: {
    projects: [
      {
        extends: true,
        plugins: [
          storybookTest({
            configDir: fileURLToPath(new URL('./.storybook', import.meta.url)),
            initialGlobals: { mokapiValidation: 'on' },
          }),
        ],
        // test: { name: 'storybook', browser: { ... } }
      },
    ],
  },
});
```

```ts
// .storybook/main.ts
export default {
  addons: [
    {
      name: 'mokapi-storybook-addon',
      options: {
        defaultValidation: false,
        defaultTest: 'error',
      },
    },
  ],
};
```

With this setup a contract error fails the component test in `'error'` mode,
whatever the toolbar shows.

The wait limit counts inside Vitest's own limit. If you set a large
`parameters.mokapi.timeout` or `defaultTimeout`, raise `testTimeout` of the
Vitest project as well (Vitest's default is 5 s), otherwise Vitest ends the test
first:

```ts
// vitest.config.ts, in the project with storybookTest
test: { testTimeout: 20000 },
```

The addon adds its result with Storybook's reporting API as a report of type
`mokapi`, with the status `passed`, `warning` or `failed`.

## Results of Run tests

When a run started with **Run tests** (the testing module below the sidebar)
is over, the **Mokapi** manager reads its results from the channel event
`storybook/test/channel` that `@storybook/addon-vitest` sends. It then updates
the Mokapi statuses in the sidebar:

- A story in `'warning'` mode with a failing result gets a yellow status.
- A story in `'error'` mode with a failing result gets a red status. The story
  is also listed as failed in addon-vitest's own **Component tests** status.
- A story that passed with at least one valid response gets the green success
  status.
- A story that took part in the run without one of these results loses its Mokapi
  status. This includes skipped stories (passed without a valid response),
  stories without a Mokapi report, validation off, `disable: true` and mode
  `'off'`. Skipped stories count as part of the run.
- Stories that didn't take part in the run keep their status.

The description of the status is the headline (`All responses match the contract.`
for a pass). The **Contract** row in the context menu and in the **Run tests**
widget opens the **Mokapi** panel, see [Run tests widget](#run-tests-widget). The
widget row shows the result right after the run.

The panel shows the result of the story as it does after a validation in the
preview: badge, headline, counts, notices and entries with their Request /
Response details. This includes a passed run with its valid entries. A line above the notices says `Result of the last test run.`
A story with [partial responses](/docs/integrations/storybook/usage.md#partial-responses) shows `PARTIAL` and the
partial line there as well.
The hint to turn validation on and the line
`No responses validated since this story was mounted.` don't appear for it. This
works with validation off in the Storybook UI too, so with the setup above
(`defaultValidation: false` and `initialGlobals: { mokapiValidation: 'on' }`) you
see the contract errors of the test run in the panel.

For a story in `'error'` mode, the addon only knows the text of the error, not
the entries. The panel then shows the `FAIL` badge and the headline, but no
entries. Switch validation on in the toolbar to see them.

The result of a test run stays in the panel and in the sidebar until a newer
result replaces it: a mount of the story with validation on, or a report of the
preview from a mount that started after the test run or from a re-render with
changed args or globals. A report from a mount that started before the test run
and arrives only after it doesn't replace the result. A newer test run replaces
a result from the preview. **Clear all statuses** removes the Mokapi statuses and
the stored results of test runs. The results of test runs are kept in the
browser only, so they are gone after you reload the page.

The Storybook test-runner (the Playwright based one, not Vitest) doesn't fail on
contract errors anymore. Use `@storybook/addon-vitest` to fail stories in CI.

