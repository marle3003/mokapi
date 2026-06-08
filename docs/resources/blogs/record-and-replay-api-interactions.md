---
title: "Record and Replay Real API Traffic for Testing and Offline Development"
description: Capture live API interactions with Mokapi and replay them in CI, demos, or offline dev. No backend required.
subtitle: Stop depending on live backends for tests and demos. Capture real API traffic once with Mokapi and replay it whenever you need it.
tech: http
tags: ['HTTP']
---

# Record and Replay Real API Traffic for Testing and Offline Development

## The Problem With Live Backends

Here's a situation most developers know well. You're writing a test for a feature that depends on an external API.
The API is flaky in CI. Or it requires VPN access. Or it returns slightly different data every time, which makes
your assertions brittle. Or it's a third-party service that you really shouldn't be hammering with test traffic.

So you write a mock. But writing a good mock takes time, and a mock you wrote by hand is only as accurate as
your memory of what the real API actually returns. Edge cases get missed. Field names get typo'd. The mock
drifts from reality and nobody notices until something breaks in production.

What if instead of writing a mock from scratch, you just... captured what the real API actually said and played
it back?

That's the pattern I built into Mokapi. You run Mokapi in front of your real backend, it forwards requests through
and records every request and response, and then you can replay those recordings whenever you need them. No
live backend required.

---

## How It Works

The recording workflow has two phases.

In the **recording phase**, Mokapi sits between your application and the real backend. It forwards every request
through, captures the request and response, and writes them to a JSON file. Your application behaves exactly
as it would in production because it's talking to the real backend. You're just capturing the conversation.

In the **replay phase**, you swap the forwarding script for a replay script. Incoming requests get matched against
the recordings and the captured responses are returned. The backend doesn't need to be reachable at all.

---

## Step 1: The Recording Script

The script uses Mokapi's JavaScript API to register an HTTP event handler. It fires after every request, reads
the existing recordings file, appends the new request and response, and writes it back.

```typescript
import { on, read, writeString } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    let data = []
    try {
      const s = read('./recordings.json')
      data = JSON.parse(s)
    } catch {}

    data.push({ request, response })
    writeString('./recordings.json', JSON.stringify(data, null, 2))
  }, {
    priority: -1
  })
}
```

The priority: `-1 setting is important. It tells Mokapi to run this handler after all other handlers have finished,
meaning the response is already fully populated when we record it. Without this, you might capture an
empty response.

---

## Step 2: Combining Recording With Forwarding

The recording script on its own only captures. To also forward requests to the real backend, you combine both
handlers in the same script. The forwarding handler runs first at the default priority, then the recording
handler runs after it at priority `-1`.

```javascript
import { on } from 'mokapi'
import { fetch } from 'mokapi/http'
import { read, writeString } from 'mokapi'

export default async function() {
  on('http', async (request, response) => {
    const url = getForwardUrl(request)

    if (!url) {
      response.statusCode = 500
      response.body = 'Unknown backend'
      return
    }

    try {
      const res = await fetch(url, {
        method: request.method,
        body: request.body,
        headers: request.header,
        timeout: '30s'
      })

      response.statusCode = res.statusCode
      response.headers = res.headers

      const contentType = res.headers['Content-Type']?.[0] || ''
      if (contentType.includes('application/json')) {
        response.data = res.json()
      } else {
        response.body = res.body
      }
    } catch (e) {
      response.statusCode = 500
      response.body = e.toString()
    }
  })

  on('http', (request, response) => {
    let data = []
    try {
      data = JSON.parse(read('./recordings.json'))
    } catch {}

    data.push({ request, response })
    writeString('./recordings.json', JSON.stringify(data, null, 2))
  }, { priority: -1 })

  function getForwardUrl(request) {
    switch (request.api) {
      case 'backend-1':
        return `https://backend1.example.com${request.url.path}?${request.url.query}`
      case 'backend-2':
        return `https://backend2.example.com${request.url.path}?${request.url.query}`
      default:
        return undefined
    }
  }
}
```

One thing worth knowing: Mokapi still validates requests and responses against your OpenAPI spec while forwarding
and recording. That means you're only recording interactions that conform to the API contract. Replays are reliable
because the source material was valid to begin with.

---

## Step 3: The Replay Script

Once you have recordings, the replay script reads them at startup and matches incoming requests against them.

```javascript
import { on, read } from 'mokapi'

export default function() {
  let recordings = []
  try {
    recordings = JSON.parse(read('./recordings.json'))
  } catch (e) {
    console.error('Failed to load recordings:', e)
    return
  }

  on('http', (request, response) => {
    const match = recordings.find(r =>
      r.request.method === request.method &&
      r.request.url.path === request.url.path &&
      r.request.url.query === request.url.query
    )

    if (match) {
      response.statusCode = match.response.statusCode
      response.headers = match.response.headers
      response.body = match.response.body
      response.data = match.response.data
    } else {
      response.statusCode = 404
      response.body = 'No recording found for this request'
    }
  })
}
```

The matching logic here is intentionally simple: method, path, and query string. Depending on your use case
you might want to extend it to match on request body for POST and PUT requests, ignore headers that change
between runs like timestamps or request IDs, or match path patterns instead of exact paths for routes with
dynamic segments like `/users/:id`.

---

## What You Can Do With This

**Regression testing.** Record a set of real interactions from staging. Replay them in CI after every deploy. If a response
changes unexpectedly, your test catches it before it reaches production.

**Offline development.** Record a comprehensive set of API interactions once, commit the JSON file, and your whole
team can develop without VPN access or network connectivity. Especially useful when the backend is owned by
another team or a third party.

**Demos and presentations.** Replay predictable, polished responses without hoping the backend behaves during a live
demo. Record the happy path once and replay it as many times as you need.

**Load testing.** Replay recorded traffic against your services without hitting real backends. Extend the replay
script to simulate concurrent users.

---

## A Few Things to Keep in Mind

Before committing recordings to version control, strip out authentication tokens, personal information, and
anything else that shouldn't sit in a repo. It's worth building a sanitization step into your workflow early
rather than retrofitting it later.

It also helps to organize recordings by scenario rather than keeping one giant file. A recording for the checkout
flow and a recording for the user profile flow are easier to maintain and reason about separately.

And as your API evolves, keep recordings for older versions if you need to support backward compatibility testing.
The JSON format is simple enough that versioning the files is straightforward.

---

## Where to Go From Here

The record and replay pattern works well as a standalone tool, but it really shines when combined with Mokapi's
other capabilities: OpenAPI validation to ensure you're only recording valid interactions, JavaScript handlers
to modify responses on the fly during replay, and the HTTP API to build a simple interface for browsing and
selecting which recordings to replay.

If you've been hand-writing mocks and watching them drift from reality, recording from the real thing is worth trying.
You capture what actually happens once, and replay it as many times as you need.

> Record real user behavior once. Replay it endlessly. Evolve your APIs with confidence.