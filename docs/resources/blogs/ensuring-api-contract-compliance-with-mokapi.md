---
title: "Catch Breaking API Changes Before They Reach Production"
description: Use Mokapi as a validation layer between clients and backends to enforce OpenAPI contracts at runtime. No backend changes needed.
subtitle: API contracts drift silently until something breaks. Use Mokapi as a transparent validation layer to catch violations the moment they happen, in development, CI, or between services.
image:
    url: /mokapi-using-as-proxy.png
    alt: Flow diagram illustrating how Mokapi enforces OpenAPI contracts between clients, Playwright tests, and backend APIs.
tech: http
links:
  items:
    - title: Get Started with Mokapi
      description: Up and running in seconds, no infrastructure changes needed
      href: /docs/get-started/installation
    - title: Record & Replay API Traffic
      description: Capture real API traffic and replay it in CI or offline development
      href: /resources/blogs/record-and-replay-api-interactions
---

# Catch Breaking API Changes Before They Reach Production

## The Problem: Contracts Break Silently

Here's how it usually goes. A backend developer renames a field. It seems harmless. They update the code,
the tests pass, and it ships. Three days later someone files a bug: the mobile app is crashing. The frontend
is showing blank data. An integration partner is getting 422s they can't explain.

The OpenAPI spec was never updated. Or it was updated but the frontend was reading an old version. Or the backend
changed behavior that the spec technically allowed but nobody expected. The contract drifted, silently, and nobody
caught it until users did.

This isn't a discipline problem. It's a tooling problem. When nothing enforces the contract at runtime, drift is
inevitable. You're relying on everyone to manually keep specs, code, and clients in sync across multiple teams,
repositories, and deploy cycles. That doesn't scale.

What you need is something that automatically validates every request and response against the spec as traffic
flows through. That's what I built into Mokapi.

---

## How Mokapi Works as a Validation Layer

The idea is simple. Mokapi sits between your client and your backend. Every request that comes in gets validated
against your OpenAPI spec. Valid requests get forwarded to the backend. The response comes back, gets validated
too, and only then reaches the client.

Your backend doesn't change. Your client doesn't change. You just point the client at Mokapi instead of directly
at the backend, and Mokapi becomes a transparent contract enforcement layer.

When something violates the spec, you get a clear error message telling you exactly what failed and why. Not a
cryptic 500 somewhere downstream. Not a silent data corruption. An immediate, actionable validation error at
the point where the contract was broken.

![Flow diagram showing Mokapi positioned between clients and backend APIs, validating requests and responses](/mokapi-using-as-proxy.png "Mokapi validates all traffic against your OpenAPI spec, catching contract violations before they cause problems")

---

## The Forwarding Script

Setting this up takes one JavaScript file. The script registers an HTTP event handler that forwards every
incoming request to the real backend and lets Mokapi handle the validation automatically.

```typescript
import { on } from 'mokapi'
import { fetch } from 'mokapi/http'

export default async function() {
  on('http', async (request, response) => {
    const url = getForwardUrl(request)

    if (!url) {
      response.statusCode = 500
      response.body = 'Failed to forward request: unknown backend'
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

The `request.api` field contains the `info.title` from your OpenAPI spec and is used here purely for routing:
it tells the script which backend URL to forward the request to. If you have multiple backends,
you just add them to the switch statement.

The choice between `response.data` and `response.body` is what controls validation. Setting `response.data` tells
Mokapi to validate the response against the schema defined in the spec. Setting `response.body` skips validation
and passes the content through as-is. That's why the script uses `response.data` for JSON responses and falls back
to `response.body` for everything else: you get validation where it's meaningful and a clean passthrough where
it isn't.

---

## Where to Put It

The same forwarding script works in several different places depending on where you need contract enforcement.

**Between frontend and backend**. Drop Mokapi into your local development setup and point your frontend at it.
Every API call your frontend makes gets validated both ways. You'll know immediately when the backend changes
something that breaks the contract, rather than finding out when a UI component starts rendering empty data.

**Between services**. In a microservice architecture, service A calling service B is just another HTTP interaction.
Put Mokapi in the path and both sides get validated. When service B's team changes a response format, service
A's CI pipeline fails before the change ships.

**In your Playwright tests**. This is one of the more powerful setups. Run Playwright against your real backend
through Mokapi, and your test suite becomes a contract test suite automatically. If the backend breaks the
contract, the CI pipeline fails with a clear validation error, not a confusing assertion failure buried in
test output.

**In Kubernetes preview environments**. Deploy Mokapi as a sidecar or standalone service in your preview environments.
Every pull request gets contract validation before it reaches staging.

---

## What You Actually Catch

It's worth being concrete about what validation covers, because it's more than just "wrong field name."

Mokapi validates the HTTP method and URL structure against the paths defined in the spec. It checks request
headers and query parameters against their definitions. It validates request bodies against the schema,
including required fields, data types, and format constraints. It does the same for responses: status codes,
response headers, and response body structure.

So a renamed field is caught. But so is a backend that starts returning a string where the spec says number.
Or a client that starts sending a request body without a required field the spec demands. Or a backend that
returns a 200 with a response body that doesn't match the documented schema for that endpoint.

All of it, caught at the boundary, with a clear message about what violated what.

---

## Getting Started

1. Define your OpenAPI spec for the API you want to validate
2. Create the forwarding script above, mapping request.api values to your backend URLs
3. Run Mokapi locally, in Docker, or in Kubernetes
4. Point your client at Mokapi instead of directly at the backend

No changes to your backend. No changes to your client. Just a validation layer in the middle that
holds everyone to the contract.

If your team has been dealing with the slow, painful process of tracking down contract violations after
they've already caused problems, this is the kind of thing that changes how that feels. You find out immediately,
at the source, with a message that tells you exactly what's wrong.

{{ card-grid key="links" }}