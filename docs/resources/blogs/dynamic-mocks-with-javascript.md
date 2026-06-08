---
title: "From Static to Dynamic: Smarter API Mocks with Mokapi Scripts"
description: Static mocks always return the same thing. Mokapi Scripts let your mock react to the actual request, in plain JavaScript, no new DSL to learn.
subtitle: Static mocks return the same response every time. Here's how to write JavaScript handlers that react to real request data, simulate stateful workflows, and behave like a real backend.
---

# From Static to Dynamic: Smarter API Mocks with Mokapi Scripts

## What Static Mocks Can't Do

Static responses are fine for simple cases. But real development surfaces their limits quickly.

Your frontend needs to handle 403s and 429s, but the mock always returns 200. Admin users should see
different data than regular users, but the mock only knows one response. After a POST, the next GET
should reflect the change, but the mock has no memory. Query parameters like `?page=2` or `?name=laptop`
should produce meaningful results, but the mock ignores them entirely.

And then there are the flows that depend on sequence. Login, then fetch a profile, then update settings.
Each step should affect what the next one returns. A static mock can't model that. It just returns the same
thing every time, regardless of what came before.

These aren't edge cases. They're the scenarios you actually need to test.

---

## What Mokapi Scripts Are

A Mokapi Script is a JavaScript module that sits alongside your OpenAPI or AsyncAPI spec. Instead of
returning a fixed response, it lets you define behavior: inspect the incoming request, look at the method,
the path, the query parameters, the headers, the body, and decide what to send back.

They're the difference between a mock that says "here's a user" and one that says "here's the right user, given
who's asking what role they have, and what they just posted."

No new DSL to learn. Just JavaScript, or TypeScript if you prefer.

---

## A Concrete Example: Dynamic Product Search

Let's build a product list endpoint that actually does something. It should return the full catalog when no
filter is applied, filter by name when `?name=` is provided, and return a 400 with a custom message when
`?name=error` is passed to simulate a validation error.

```typescript
import { on } from 'mokapi'

const products = [
  { name: 'Laptop Pro 15' },
  { name: 'Wireless Mouse' },
  { name: 'Mechanical Keyboard' },
  { name: 'Noise Cancelling Headphones' },
  { name: '4K Monitor' },
  { name: 'USB-C Hub' }
]

export default () => {
  on('http', (request, response) => {
    const nameFilter = request.query.name

    if (!nameFilter) {
      response.data = { products: products }
      return
    }

    if (nameFilter === 'error') {
      response.rebuild(400)
      response.data.message = 'A custom error message'
      return
    }

    const matched = products.filter(p =>
      p.name.toLowerCase().includes(nameFilter.toLowerCase())
    )

    response.data = { products: matched }
  })
}
```

A few things worth pointing out. `request.query.name` gives you the query parameter directly, already parsed.
`response.data` tells Mokapi to validate the response against the schema in your OpenAPI spec before returning
it. And `response.rebuild(400)` switches the response to the 400 definition from your spec, so the error
response is also schema-validated.

This is already more useful than any static mock. The same endpoint returns different things based on what
the request actually contains.

---

## What Else You Can Do

The product search example shows the basics: read the request, decide the response. But the same pattern
extends to a lot of other scenarios.

**Role-based responses**. Check a header or a token claim and return different data for admin versus regular users.
The mock behaves the way the real backend would.

**Stateful workflows**. Store state in a JavaScript variable. A POST adds an item, the next GET returns it.
The mock remembers what happened.

**Auth and session behavior**. Missing or invalid tokens return 401. Expired tokens return a specific error message.
Valid tokens return the right data. Your frontend can test all of these without a real auth server.

**Simulating errors and delays**. Force a 500, a 429 rate limit, or an artificial delay to test how your application
handles degraded conditions. No need to break the real backend to test your error handling.

**Mocking an identity provider**. Serve an HTML login form, handle the form submission, issue a redirect with a session
token. You can simulate an entire OAuth flow without a real identity provider.

---

## Getting Started

First, write your OpenAPI or AsyncAPI spec. This defines the shape of your API: the endpoints, the request formats,
the response schemas. Mokapi uses this to validate everything that flows through.

Second, add a JavaScript file next to your spec. Register an `on('http', ...)` handler, inspect the request, set
`response.data` to whatever makes sense for that request. That's the whole model.

Third, run Mokapi. Your dynamic mock is live, validated against the spec, and visible in the Mokapi dashboard
where you can inspect every request and response that came through.

---

## What to Read Next

- [Debugging Mokapi JavaScript](/resources/blogs/debugging-mokapi-scripts)  
  When your handler isn't behaving the way you expect, here's how to use `console.log`, `console.error`, and event handler tracing to see exactly what's happening inside your scripts.
- [End-to-End Testing with Mock APIs](/resources/blogs/end-to-end-testing-with-mocked-apis)  
  How to replace live backend dependencies in your test pipeline with Mokapi mocks, so your tests run fast and pass consistently in CI.
- [Building a Virtual Hardware Simulation with Mokapi](/resources/blogs/building-virtual-hardware-simulation)  
  A deep dive into what's possible when you combine Mokapi's spec validation with its JavaScript API. Uses a parcel terminal as a case study to show how far you can take dynamic mocking.