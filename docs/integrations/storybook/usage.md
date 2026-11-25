---
title: Using the Storybook Addon
description: Switch Mokapi validation on and off, choose the API, disable stories, validate partial responses and read the Mokapi panel in Storybook.
---

# Using the Storybook Addon

How to work with the Mokapi addon day to day: the toolbar toggle, the API a story belongs to,
partial responses, the **Mokapi** panel and Docs pages. If you haven't installed the addon yet, start with the
[overview](/docs/integrations/storybook).

## Toggle

The toolbar has one icon button, **Validate with Mokapi**. One click switches
validation on or off. The button is highlighted while validation is on, and its
tooltip shows the state:

- `Mokapi validation is on. Click to turn off.`
- `Mokapi validation is off. Click to turn on.`

The button sets the global `mokapiValidation` to `on` or `off`. The initial value
comes from `defaultValidation`. With `defaultValidation: true`, the button starts
highlighted. Assistive technology reads the button as a switch named "Validate
with Mokapi" that is checked while validation is on.

The button only appears in story view, not on Docs pages, because the story
remount below only works there.

Switching does not reload the preview. The addon remounts the current story, so
a component that fetches on mount requests its data again. When you switch
validation on, those requests are validated right away, and the panel list
starts empty.

## Choosing the API

Mokapi can serve several APIs. Without `parameters.mokapi.api`, Mokapi searches
all of them for an operation that matches the request's method and path.

If more than one API matches, Mokapi can't choose. The addon then shows the
request as **not in contract**, and the entry contains Mokapi's message with the
names of the matching APIs. To fix it, set the API's `info.title`:

```ts
// .storybook/preview.ts, a story file or a single story
export default {
  parameters: {
    mokapi: { api: 'Profile API' },
  },
};
```

`parameters.mokapi` is inherited from preview, component and story, as other
Storybook parameters. A story can overwrite the value of the project.

If `api` names an API that Mokapi doesn't know, the panel shows a notice with
Mokapi's message, see [Errors](/docs/integrations/storybook/reference.md#errors).

## Disabling a story

A story that has no contract, for example a mock for a third-party API, can opt
out:

```ts
export const ThirdParty = {
  parameters: {
    mokapi: { disable: true },
    msw: { handlers: [/* ... */] },
  },
};
```

The addon sends nothing to Mokapi for this story, also when the toggle is on.
The Mokapi panel shows the hint
`Validation is disabled for this story with parameters.mokapi.disable.` for
such a story.


## Partial responses

A story often mocks only the data it needs, for example a long `name` to test a
layout. Against a contract with required properties, such a mock is invalid, and
every story with partial data shows a red entry. With partial responses, Mokapi
doesn't report missing properties in the response body, but still checks
everything the story delivers.

Switch it on for all stories with `defaultPartial` in `.storybook/main.ts`:

```ts
// .storybook/main.ts
export default {
  addons: [
    {
      name: 'mokapi-storybook-addon',
      options: {
        defaultPartial: true, // default false
      },
    },
  ],
};
```

`parameters.mokapi.partial` overrides it for a story, a component or the whole
project, like `test` overrides `defaultTest`:

```ts
// A story that must deliver the complete response
export const Complete = {
  parameters: {
    mokapi: { partial: false },
  },
};
```

For a story with partial responses, the addon sends `"partialResponse": true` to
Mokapi with every validation. Otherwise it leaves the field out.

Mokapi skips these checks in the response body, on every nesting level:

- `required`
- `dependentRequired`
- `oneOf`: one matching subschema is enough. If several match, the body is valid,
  because without the required properties the variants often can't be told apart

Everything else stays strict: types, formats, enums and `additionalProperties`
for the properties the body contains. The request is validated strictly too.
Required response headers and a missing response body are still reported.

The panel shows that a story validates with partial responses, so a green result
isn't read as "complete", see [Toolbar](#toolbar). Mokapi describes the checks in
[Partial responses](/docs/configuration/dashboard-api.md#partial-responses).
Requires Mokapi 0.60.0. With an older Mokapi, the panel shows a notice, see
[Notices](#notices).

## Panel and badge

The **Mokapi** panel shows the responses of the selected story since it was
mounted. A re-render that keeps the component, for example a change in Controls
or in a toolbar, keeps the list and adds the responses of the new requests.
Switching to another story, **Reload Story** and switching validation start a
new list. Until the first response arrives, the panel shows
`No responses validated since this story was mounted.`

The tab title shows a badge with the number of invalid responses. Responses that
are not in the contract and not-reachable entries are not counted.

### Toolbar

While validation is on, the panel starts with a toolbar. From left to right it
has:

- A report badge: `PASS`, `PARTIAL`, `WARN`, `FAIL`, `SKIP` or `WAIT`. Hover it to see the
  test mode of the story. See the table below.
- A headline in bold, only for `WARN`, `FAIL` and `SKIP`. For `WARN` and `FAIL` it
  is the headline of the [error message](/docs/integrations/storybook/failing-stories.md#error-message), for example
  `1 response does not match the contract.` For `SKIP` it is
  `No response was checked against the contract.`, or the message about the
  missing MSW worker if `context.msw` is not set, or the message about a handler
  that did not answer in time.
- The counts in one line, for example
  `2 responses validated, 1 valid, 1 invalid`. The line counts all entries since
  the story was mounted and updates as results come in. Kinds that don't occur
  are left out (`valid`, `invalid`, `not in contract`, `Mokapi not reachable`).
- The line `Partial responses: missing properties are not checked.`, only while
  the story validates with [partial responses](#partial-responses). It is shown
  in every state, also next to `WARN` and `FAIL`.
- On the right: the **Rerun** icon button, a help icon that opens the
  [addon documentation](https://mokapi.io/docs/integrations/storybook), and a
  link to the Mokapi dashboard.

The badge depends on the test mode of the story, see
[Failing stories](/docs/integrations/storybook/failing-stories.md):

| Badge  | When |
|--------|------|
| `WAIT` | The story is still waiting for its requests and their validation |
| `FAIL` | Test mode `'error'` and a failing result |
| `WARN` | Test mode `'warning'` and a failing result |
| `PASS` | No failing result and at least one valid response |
| `PARTIAL` | Like `PASS`, for a story with [partial responses](#partial-responses). Green as well |
| `SKIP` | No failing result and no valid response: every response is not in the contract, the story sent no request, the MSW worker is missing, or a handler did not answer within the [wait limit](/docs/integrations/storybook/failing-stories.md#wait-limit) |

`SKIP` is shown in the panel only. Storybook still reports the story as passed.
A story that is mounted again and checks no response (for example because your
app answers from a cache) keeps the result of its last check in the panel and
the sidebar. `SKIP` only shows for a story that has no earlier result.
The badge follows the report of the story, so a response that arrives after the
report appears in the list but doesn't change the badge.
`PARTIAL` changes the badge only: the report status stays `passed`, so the
sidebar status, Vitest and the **Run tests** widget treat the story as a plain
pass.

In test mode `'off'` and for a story with `parameters.mokapi.disable` there is no
badge. If a story is aborted before it reports, for example because you switch
stories, the badge stays at `WAIT` until the story is mounted again. The details
of a `FAIL` or `WARN` are in the entries and notices below the toolbar.

The tooltip of the badge shows the test mode, so you can see what a CI run does
with the story, see [Failing stories](/docs/integrations/storybook/failing-stories.md):

- `Test mode: error. Contract errors make this story fail.`
- `Test mode: warning. Contract errors are reported as warnings, the story stays green.`

In test mode `'off'` there is no badge to hover. The toolbar shows
`Test mode: off. Contract errors are shown here only.` as text instead. You can't
change the mode in the panel, it is set in code only.

**Rerun** remounts the story, like switching validation does. The component
requests its data again, the list starts empty and the story gets a new report.
The preview is not reloaded. The button is disabled while the story renders,
because remounting a story that is still pending makes Storybook reload the
preview.

### Notices

Notices appear below the toolbar when the addon can't validate:

- Mokapi is not reachable.
- The Mokapi at `dashboardUrl` has no validation endpoint and must be updated to 0.58.0.
- Mokapi reports a configuration error, for example an unknown `api`.
- `parameters.mokapi.test` has an invalid value.
- `parameters.mokapi.partial` has an invalid value.
- `parameters.mokapi.timeout` has an invalid value.
- The Mokapi at `dashboardUrl` doesn't support partial responses and must be
  updated to 0.60.0. The entries show what Mokapi returned, validated strictly.
- Mokapi did not answer in time: MSW answered, but the validation did not finish
  within the wait limit, see [Failing stories](/docs/integrations/storybook/failing-stories.md).
- The MSW handler did not answer in time. This is not a failure, nothing was
  validated, see [Wait limit](/docs/integrations/storybook/failing-stories.md#wait-limit).
- The MSW worker is missing from `context.msw`.

Notices that don't fail a story (a missing worker, a handler that did not answer)
are shown in the warning color, the others in red.

The exact texts are listed in [Errors](/docs/integrations/storybook/reference.md#errors). With validation off, the panel
shows only a hint how to turn it on, without toolbar.

### Entries

Each entry is a row with an icon, the result, the method of the request as a
badge, the URL of the request, and the HTTP status of the response and the time
on the right. The result is one of these:

| Result                 | Icon | Meaning                                                                          |
|------------------------|------|----------------------------------------------------------------------------------|
| valid                  | check mark | Request and response fit the contract. The URL links to the matched operation |
| invalid                | red stop sign | The contract is violated. The row has a red background and lists each error with its `location` (for example `response.body`) and message below it |
| not in contract        | empty circle | Mokapi found no operation for the request, or several APIs match. The entry shows Mokapi's message |
| Mokapi not reachable   | empty circle | The request to Mokapi failed. The entry shows the URL and a hint to `dashboardUrl` |

In valid and invalid entries, the URL of the request (for example
`/api/profile/1`) links to the page of the operation Mokapi matched in the Mokapi
dashboard and opens in a new tab. The method badge in front of it is not part of
the link. Hover the link to see the operation, for example `GET /profile/{id}`.
Entries that are not in the contract or not reachable have no link.

The method is a badge with the method in upper case, colored by method:

| Method        | Color  |
|---------------|--------|
| GET, OPTIONS  | blue   |
| POST          | green  |
| PUT, PATCH    | orange |
| DELETE        | red    |
| HEAD, TRACE   | gold   |
| other         | gray   |

The status is a badge with the status code, colored by its class:

| Status  | Color  |
|---------|--------|
| 2xx     | green  |
| 3xx     | blue   |
| 4xx     | orange |
| 5xx     | red    |
| other   | gray   |

Hover the badge to see the status text, for example `Not Found` for 404. A code
without a known text shows its class instead: Informational, Success, Redirect,
Client error or Server error. For a code outside 100-599 (for example `0` for
`HttpResponse.error()`), the tooltip shows only the code. A method and a status
can have the same color, for example `GET` and a 3xx status are both blue. The
time is `HH:mm:ss` in the time zone of your browser.

Under each entry, **Request / Response** is collapsed. Its **Request** heading
shows the method badge and its **Response** heading shows the status badge. Open
it to see the headers and the body of the request and of the response as the
addon sent them to Mokapi. A JSON body is pretty-printed, and an empty body
shows `(empty)`. The panel shows at most 10,000 characters of a body and adds
`(truncated to 10,000 characters)` if it cut one. Mokapi always gets the full
body.


## Docs pages

Use Storybook's Docs page with stories that are not inline:

```ts
// .storybook/preview.ts
export default {
  parameters: {
    docs: { story: { inline: false, iframeHeight: '300px' } },
  },
};
```

Each story then runs in its own iframe with its own MSW worker. Inline stories
share one worker, so the addon can't attribute their results to a story. See
[Limitations](/docs/integrations/storybook/reference.md#limitations).

The toolbar button is not shown on Docs pages. Switch validation on in story
view first, or set `defaultValidation: true`.

The story frame on a Docs page shows no contract result, also not for a story in
`'error'` mode. The story still gets its status in the sidebar, as in story view.

