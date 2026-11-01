---
title: Develop and Test Real-Time Collaboration Locally With AsyncAPI, Mokapi, and Playwright
description: Learn how to mock a real-time collaborative editor, using AsyncAPI and Mokapi. Covers presence, shared state, and conflict handling.
subtitle: "This post uses one concrete example to show how Mokapi can support an event-driven application that talks over WebSockets: a small collaborative text editor, similar in spirit to a page in Confluence or Notion."
---

# Develop and Test Real-Time Collaboration Locally With AsyncAPI, Mokapi, and Playwright

## The example

The demo is a single text box that several people can edit at the same time. Each person opens
the page in their own browser and joins a document by its id. From then on:

- everyone sees who else is currently on the document, 
- an edit made by one person appears on the other screens, 
- someone who joins later receives the content that already exists, 
- someone who loses the connection and comes back receives what they missed, 
- an edit based on an outdated version of the document is not applied.

The pieces:

- a React page with a textarea,
- a WebSocket server, which Mokapi simulates from an AsyncAPI file and a short script,
- four Playwright tests that open several browsers and check the behaviors above.

The code is in a [small repository](https://github.com/marle3003/mokapi-websocket-workflow). You can clone
it and run npm test while you read, or just follow along. We'll go through it in order: the contract, the
mock's behavior, the page, running it during development, then the tests, and CI.

### Why a mocked WebSocket server is handy

A WebSocket isn't a single request and response. It's a conversation, and in a collaborative editor, 
the server remembers things during that conversation:

- **Presence:** who is on this page right now?
- **Catch-up:** when someone joins late, they receive what's already been written.
- **Conflicts:** two edits are based on the same version, and one of them is now outdated.
- **Reconnects:** a client drops out, the others carry on, and the client comes back.

Those moments are the heart of the feature. They are also the moments that take effort to stage with a real
system: two sessions, careful timing, a document in just the right state.

With a mock you control, each of them is something you can trigger on demand. That helps in two ways. While
you build, you get a server to talk to from day one. When you test, you get a partner that behaves the same
way every run.

## The demo at a glance

The repo is small on purpose:

```bash
mokapi-websocket-workflow/
├── mock/asyncapi.yaml       # the WebSocket contract
├── mock/collab.js           # Mokapi Script: the stateful "backend"
├── index.html               # React app with a textarea (no build step)
├── server.ts                # tiny static server for index.html
├── playwright.config.js     # starts the servers, runs the tests
├── tests/collab.spec.js     # four end-to-end tests
├── package.json
└── .github/workflows/test.yaml
```

Three files carry the idea: a contract (`asyncapi.yaml`), a brain (`collab.js`) and a proof
(`tests/collab.spec.js). We'll take them in that order.

### Step 1: Describe the conversation with AsyncAPI

[AsyncAPI](https://www.asyncapi.com) is a specification format for event-driven APIs. It plays the role for WebSockets and message streams
that OpenAPI plays for HTTP. Mokapi reads the file and starts a server that follows it.

```yaml
asyncapi: 3.0.0
info:
  title: Collaborative Doc Service
  version: 1.0.0
  description: A shared document editing service, similar to Confluence or Notion.

servers:
  mokapi:
    host: 'localhost:8765'
    protocol: ws

channels:
  doc:
    address: /doc/{docId}
    parameters:
      docId:
        description: Unique identifier for the shared document.
    bindings:
      ws:
        query:
          type: object
          properties:
            userId:
              type: string
          required: [userId]
    messages:
      presence:
        payload:
          type: object
          properties:
            userId: { type: string }
            status: { type: string, enum: [joined, left] }
          required: [userId, status]
      patch:
        payload:
          type: object
          properties:
            userId:      { type: string }
            offset:      { type: integer }
            deleteCount: { type: integer }
            text:        { type: string }
            version:     { type: integer }
          required: [userId, offset, deleteCount, text, version]
      sync:
        payload:
          type: object
          properties:
            content: { type: string }
            version: { type: integer }
            users:   { type: array, items: { type: string } }
          required: [content, version, users]

operations:
  sendPatch:
    action: send
    channel:
      $ref: '#/channels/doc'
    messages:
      - $ref: '#/channels/doc/messages/patch'

  receiveUpdate:
    action: receive
    channel:
      $ref: '#/channels/doc'
    messages:
      - $ref: '#/channels/doc/messages/sync'
      - $ref: '#/channels/doc/messages/presence'
```

#### A quick tour:

**The address is the document.** `/doc/{docId}` gives every document its own room. Clients on `/doc/spec-review` never
see traffic from `/doc/meeting-notes`.

Three message types, three jobs. `presence` announces that someone joined or left. `patch` is a small edit
from a client. `sync` is the full state of the document, sent by the server.

`version` makes conflicts detectable. Each patch says which version of the document it was based on.
You'll see in a moment what the mock does with that.

**The query binding.** `bindings.ws.query` declares that clients connect with `?userId=...`. Mokapi exposes
the query parameters you declare in the binding to your script and validates them.

The contract is also enforced. If your frontend sends a patch without an `offset`, Mokapi rejects it. So the
mock doubles as a check that your client really speaks the agreed language.

### Step 2: Give the mock a memory

A plain relay is enough for a simple chat. A collaborative document needs a state. That's what a Mokapi Script is
for: JavaScript that reacts to WebSocket events.

```typescript
import { on } from 'mokapi'

// one entry per docId: { content, version, users: Set }
const docs = {}

function getDoc(docId) {
  if (!docs[docId]) {
    docs[docId] = { content: '', version: 0, users: new Set() }
  }
  return docs[docId]
}

function syncMessage(doc) {
  return {
    content: doc.content,
    version: doc.version,
    users: Array.from(doc.users)
  }
}

export default function () {
  on('websocket', function (event) {
    const doc = getDoc(event.channel.params['docId'])
    const userId = event.client.query['userId']

    if (event.type === 'connect') {
      doc.users.add(userId)

      // catch the new client up on everything that happened before
      event.client.send(syncMessage(doc))

      // tell everyone that someone showed up
      event.broadcast({ userId, status: 'joined' })
    }

    if (event.type === 'message') {
      const patch = event.message

      // stale patch: the client worked off an old version
      if (patch.version < doc.version) {
        event.client.send(syncMessage(doc))
        return
      }

      const before = doc.content.slice(0, patch.offset)
      const after = doc.content.slice(patch.offset + patch.deleteCount)
      doc.content = before + patch.text + after
      doc.version += 1

      event.broadcast(syncMessage(doc))
    }

    if (event.type === 'close') {
      doc.users.delete(userId)
      event.broadcast({ userId, status: 'left' })
    }
  })
}
```

That's the whole "backend". Each of the three events lines up with one of the moments we care about.

#### Connect: a late joiner sees the document

When Bob connects, the script immediately sends him a sync with the current content, the version, and the list
of users. He doesn't wait for Alice's next keystroke to see anything. Everyone else gets a presence message
that Bob joined.

#### Message: apply the edit or hand back the truth

If a patch is based on the current version, the script splices it into the content (`offset`, `deleteCount`, and
`text` describe the change), bumps the version, and broadcasts the new state.

If the patch is based on an older version, somebody else was faster. The mock doesn't apply it. It answers
that one client with the current `sync`, so the client can show whatever it should show in that situation.

#### Close: tidy up

The user leaves the list and the others are told. That keeps the "who's here" badge honest.

> Scope of the mock. This is last-write-wins on a version number. It's meant to give your client a
> predictable server to react to, not to reimplement operational transform or CRDTs. If your real server
> has smarter merge rules, the script is the place to teach the mock about them.

### Step 3: A small frontend to try it with

`index.html` is a single file with React loaded from a CDN. No bundler, no build step, so anyone can open it
straight from the repo.

Typing turns into a patch. The textarea's `onChange` compares the old and new text, finds the common start and
end, and sends only the part in the middle:

```typescript
function diff(oldStr, newStr) {
  let prefix = 0;
  const maxPrefix = Math.min(oldStr.length, newStr.length);
  while (prefix < maxPrefix && oldStr[prefix] === newStr[prefix]) prefix++;

  let oldEnd = oldStr.length;
  let newEnd = newStr.length;
  while (oldEnd > prefix && newEnd > prefix &&
         oldStr[oldEnd - 1] === newStr[newEnd - 1]) {
    oldEnd--;
    newEnd--;
  }

  return {
    offset: prefix,
    deleteCount: oldEnd - prefix,
    text: newStr.slice(prefix, newEnd)
  };
}
```

Typing, backspace and select-and-replace all end up as the same `offset / deleteCount / text` message.

**Server messages update the screen.** A `sync` replaces the content and the version, and a `presence` message
lands in an activity log.

There's also a button called "Simulate stale patch". It sends an edit that claims to be based on an old
version, so you can see the conflict path with one click. Every interesting element has a `data-testid`
(`editor`, `version`, `users`, `status`, `log`, and the buttons), which the tests will use later.

### Step 4: Use it while you develop

```bash
npm install
npm run mock
```

`npm run mock` starts Mokapi through the `go-mokapi` npm package, so there's no separate binary to install.
Open `index.html in two browser windows, use the same document id, pick two different names, and type.

At the same time, open Mokapi's dashboard at `http://localhost:8080` and go to the WebSocket tab. Every
connection and every message shows up there with direction, payload, and timestamp. It's a handy way to
see what your client really sends.

This is where the mock earns its keep before any test exists:
- You can build the editor UI before the real collaboration service is available.
- The AsyncAPI file is something frontend and backend can agree on early, and the mock keeps your client honest against it.
- Changing the behavior is a quick edit in `collab.js`. Want to see how the UI feels when a `sync` arrives two seconds late? Add a delay and reload.

### Step 5: Turn the same setup into tests

Once the mock behaves the way you want, Playwright can take over the clicking. A collaboration test needs
several users, and Playwright's browser contexts are a natural fit: one context per user, each with its
own connection, like two people on two machines.

First, a small helper and a way to keep tests apart:

```typescript
import { test, expect } from '@playwright/test'

async function joinAs(page, docId, name) {
  await page.goto('/')
  await page.getByTestId('docid-input').fill(docId)
  await page.getByTestId('userid-input').fill(name)
  await page.getByTestId('join-button').click()
  await expect(page.getByTestId('status')).toContainText('Connected', { timeout: 5000 })
}

function uniqueDocId(label) {
  return `${label}-${Date.now()}-${Math.floor(Math.random() * 10000)}`
}
```

The mock keeps document state in memory, so a unique document id per test keeps them independent without
restarting anything.

#### Test 1: A late joiner sees what's already there

```typescript
test('a client joining later immediately sees existing content', async ({ browser }) => {
  const docId = uniqueDocId('late-join')

  const alice = await (await browser.newContext()).newPage()
  const bob = await (await browser.newContext()).newPage()

  await joinAs(alice, docId, 'alice')
  await alice.getByTestId('editor').fill('Hello from Alice')
  await expect(alice.getByTestId('version')).toContainText('version 1', { timeout: 5000 })

  await joinAs(bob, docId, 'bob')

  await expect(bob.getByTestId('editor')).toHaveValue('Hello from Alice')
  await expect(bob.getByTestId('users')).toContainText('alice')
  await expect(bob.getByTestId('users')).toContainText('bob')
})
```

The wait for `version 1` makes sure Alice's edit has made the round trip before Bob joins. Waiting for something
visible in the UI works well for WebSockets, much like waiting for a response does with HTTP.

#### Test 2: Edits show up live on the other screen

```typescript
test('edits from one client sync live to another connected client', async ({ browser }) => {
  const docId = uniqueDocId('live-sync')

  const alice = await (await browser.newContext()).newPage()
  const bob = await (await browser.newContext()).newPage()

  await joinAs(alice, docId, 'alice')
  await joinAs(bob, docId, 'bob')

  await alice.getByTestId('editor').fill('sync test')

  await expect(bob.getByTestId('editor')).toHaveValue('sync test', { timeout: 5000 })
})
```

One user acts, the assertion looks at the other user's screen. Many collaboration tests have that shape.

#### Test 3: A stale patch is turned away

```typescript
test('a stale patch is rejected instead of corrupting the document', async ({ page }) => {
  const docId = uniqueDocId('conflict')

  await joinAs(page, docId, 'alice')
  await page.getByTestId('editor').fill('original content')
  await expect(page.getByTestId('version')).toContainText('version 1', { timeout: 5000 })

  const versionBefore = await page.getByTestId('version').textContent()

  await page.getByTestId('conflict-button').click()
  await page.waitForTimeout(700)

  await expect(page.getByTestId('version')).toHaveText(versionBefore)
  await expect(page.getByTestId('editor')).toHaveValue('original content')
})
```

Conflicts are tricky to catch by hand because they depend on timing. Here the button sends an outdated
version on purpose, the mock refuses it, and the test checks that nothing changed. It runs the same way
every time.

(The `waitForTimeout` is there because "nothing happened" has no event to wait for. For a stricter check,
assert on the log line the client writes when the `sync` comes back.)

#### Test 4: Reconnect and catch up

```typescript
test('a client that disconnects and reconnects catches up on missed edits', async ({ browser }) => {
  const docId = uniqueDocId('reconnect')

  const alice = await (await browser.newContext()).newPage()
  const bob = await (await browser.newContext()).newPage()

  await joinAs(alice, docId, 'alice')
  await joinAs(bob, docId, 'bob')

  await alice.getByTestId('editor').fill('first')
  await expect(bob.getByTestId('editor')).toHaveValue('first', { timeout: 5000 })

  await alice.getByTestId('disconnect-button').click()
  await expect(alice.getByTestId('status')).toContainText('Disconnected')

  await bob.getByTestId('editor').fill('first second')
  await expect(bob.getByTestId('version')).toContainText('version 2', { timeout: 5000 })

  await alice.getByTestId('reconnect-button').click()
  await expect(alice.getByTestId('status')).toContainText('Connected', { timeout: 5000 })

  await expect(alice.getByTestId('editor')).toHaveValue('first second')
})
```

Alice leaves, Bob keeps writing, Alice returns and sees everything she missed. Since the server sends a full `sync` on
every connect, a reconnect is just another join, and the test documents that behavior.

## One command to run everything

Playwright's webServer option can start more than one process. The config starts Mokapi and the
static file server, waits until both respond, then runs the tests:

```typescript
import { defineConfig } from '@playwright/test'

const externalMokapi = process.env.MOKAPI_EXTERNAL === 'true'
const webServer = []

if (!externalMokapi) {
  webServer.push({
    command: 'npx go-mokapi asyncapi.yaml collab.js',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 30000
  })
}

webServer.push({
  command: 'node server.ts',
  url: 'http://localhost:4300',
  reuseExistingServer: !process.env.CI
})

export default defineConfig({
  testDir: './tests',
  timeout: 15000,
  fullyParallel: false,
  workers: 1,
  use: { baseURL: 'http://localhost:4300', trace: 'retain-on-failure' },
  webServer
})
```

```bash
npm test
```

No extra terminal, no manual start. The MOKAPI_EXTERNAL switch is for the next section.

## Running it in CI with Mokapi in Docker

In a pipeline you may want to start Mokapi yourself, for example, as a Docker container, so
you can collect its logs. Set `MOKAPI_EXTERNAL=true` and Playwright leaves Mokapi alone:

```yaml
name: Testing Kafka Workflow

on:
  push:
    branches:
      - main
  pull_request:
  workflow_dispatch:
  schedule:
    - cron: "0 11 * * 5"

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Start Mokapi
        run: |
          docker run -d --rm --name mokapi -p 8080:8080 -p 8765:8765 -v ${{ github.workspace }}/mock:/mock mokapi/mokapi:alpha /mock

      - name: Set Up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 23

      - name: Install Dependencies
        run: |
          npm ci --ignore-scripts
          npx playwright install-deps
          npx playwright install chromium

      - name: Run Tests
        run: npx playwright test
        env:
          MOKAPI_EXTERNAL: 'true'

      - uses: actions/upload-artifact@v4
        if: ${{ !cancelled() }}
        with:
          name: playwright-report
          path: playwright-report/
          retention-days: 30

      - name: Save Mokapi logs
        if: ${{ !cancelled() }}
        run: docker logs mokapi > /var/tmp/mokapi.log 2>&1

      - uses: actions/upload-artifact@v4
        if: ${{ !cancelled() }}
        with:
          name: mokapi-logs
          path: /var/tmp/mokapi.log

      - name: Stop Mokapi
        run: docker stop mokapi
```

`if: always()` is the useful bit. When a run fails, you have Playwright's trace plus what the mock saw on its
side: connections, messages, and validation errors. Together they make it much easier to find out what happened.

## Where this fits

The mock gives you a server that is quick to start, easy to reason about, and the same on every run. That's a
good fit for developing the client and for testing how it reacts to presence changes, late joins, rejected
edits, and reconnects.

## Ideas to try next

- Add a delay to `sync` messages in the script and see how your UI behaves on a slow connection.
- Close a connection from the server side and test your retry and backoff logic.
- Send a message that violates the schema and look at how the client reacts when Mokapi closes the connection, as in the chat example of the WebSocket quick start.
- Extend presence with cursor positions and check that both users see each other's cursor.

All of these become easy once you have the simulation API: add an endpoint, add a case in the script,
call it from a test.

### Trigger edge cases from your tests with a simulation API

So far the mock reacts to what clients send. Some situations need the server to act on its own: another user's
edit arrives while you're looking at the page, the server gets slow, the state has to be reset between tests.
A WebSocket client can't ask for these.

But Mokapi can mock HTTP APIs from OpenAPI files as well, and a script can handle HTTP and WebSocket events
side by side. That gives you a neat trick: describe a small simulation API in OpenAPI, and let its handlers
manipulate the WebSocket server. Your tests then reach the edge cases with plain HTTP calls.

You can read more about a simulation API in the [Mocking a Payment API](/resources/blogs/mocking-payment-service-mokapi-playwright) article

## Wrap-up

A collaborative editor is a lot of fun to build, and the moments that make it feel good are the small ones:
a badge that shows who's here, a page already filled when you arrive, an edit that doesn't get lost.
With a mocked WebSocket server you can stage each of those moments on your own machine, click through them
while you develop, and keep them as Playwright tests once they work.

Clone the demo, run `npm test`, then change something in `collab.js`, and see what happens. That's the quickest
way to get a feel for what Mokapi can do for your own event-driven application.

## Next steps

- [WebSocket mocking overview](https://mokapi.io/docs/websocket/overview.md)
- [WebSocket quick start: a chat room mock](https://mokapi.io/docs/websocket/quick-start.md)
- [Mokapi JavaScript API](https://mokapi.io/docs/javascript-api/overview.md)
- Demo repository: [mokapi-websocket-workflow](https://github.com/marle3003/mokapi-websocket-workflow)