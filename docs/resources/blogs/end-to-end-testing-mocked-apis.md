---
title: End-to-End Testing with Mocked APIs in CI/CD
description: Replace flaky external API dependencies in your test pipeline with Mokapi mocks. Faster runs, reliable results, zero external dependencies.
subtitle: External APIs make tests slow, flaky, and fragile. Here's how to replace them with Mokapi mocks that are spec-validated, fully controllable, and ready to run anywhere.
---

# End-to-End Testing with Mocked APIs in CI/CD

## The Problem With External APIs in Tests

You've probably felt this. A test suite that passes locally starts failing in CI for no obvious reason.
You dig in and find that a third-party API returned a 503, or was just slow enough to trip a timeout. You
rerun the pipeline, it passes, and you move on. Until it happens again tomorrow.

Or maybe the problem is subtler. Your tests pass fine, but they only cover the happy path because you can't
control what the external API returns. You can't make it return a 429 rate limit on demand. You can't make
it time out. You can't make it return a malformed response to test your error handling. So those scenarios
go untested and you find out about them in production.

Mocking the external API solves both problems. You get full control over what the API returns, and you remove
the external dependency entirely. But mocks that are just hardcoded response fixtures drift from reality
the same way handwritten mocks do. What you want is a mock that's defined by the actual API contract.

That's the core idea behind Mokapi. You give it your OpenAPI or AsyncAPI spec, and it serves as a fully
spec-validated mock server. Every response conforms to the contract. Every request gets validated against it.
And because the spec is the source of truth, your mocks stay accurate as your API evolves.

---

## How It Fits Into CI/CD

The setup is straightforward. Mokapi runs as a service during your test phase, alongside your application.
Your tests hit Mokapi's endpoints instead of the real external APIs.

Your application doesn't need to know it's talking to a mock. It just points at a different host.
In CI that's typically an environment variable swap. Locally it's the same.

<img src="/ci-pipeline-flow.png" alt="Typical CI/CD Flow with Mokapi" />

---

## A Real GitHub Actions Example

Here's what this looks like in practice with GitHub Actions and a Node.js backend. The mock specs live in a
`/mocks` directory in the repo, and Mokapi runs in Docker during the test job.

```yaml
name: Node.js Backend Tests

on:
  push:
    branches:
      - main
  pull_request:
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Start Mokapi
        run: |
          docker run -d --name mokapi \
            -p 80:80 \
            -v ${{ github.workspace }}/mocks:/mocks \
            mokapi/mokapi:latest /mocks
          sleep 5

      - name: Set Up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 24

      - name: Install Dependencies
        run: npm install

      - name: Run Tests
        run: npm test

      - name: Stop Mokapi
        run: docker stop mokapi
```

A few things worth pointing out here.

The mock specs are committed to the repo alongside the code. That's intentional. When the API contract
changes, the spec changes, the mock changes, and the tests run against the updated contract. Everything
stays in sync because it's all in one place.

The `sleep 5` after starting Mokapi is a simple way to make sure the container is ready before the tests
start. For more robust setups you can use a health check endpoint instead.

Mokapi exposes port 80 for API traffic and port 8080 for its dashboard and HTTP API, which is useful for
inspecting what requests came in during a test run.

---

## Testing Scenarios You Can't Test With a Live API

This is where mocking gets really useful beyond just "removing the external dependency." With Mokapi's
JavaScript API you can simulate conditions that a real API either can't reproduce on demand or that you'd
never want to trigger against a live service.

**Timeouts and slow responses**. The script below lets you control response delay at runtime via HTTP requests,
without restarting Mokapi:

```typescript
import { on, sleep } from 'mokapi'

export default () => {
  let delay = undefined

  on('http', (request, response) => {
    if (request.key === '/simulations/delay') {
      switch (request.method) {
        case 'PUT':
          delay = request.query.duration
          break
        case 'DELETE':
          delay = undefined
          break
      }
    }
  }, { track: true })

  on('http', (request, response) => {
    if (delay) {
      sleep(delay)
    }
  }, { track: () => delay !== undefined })
}
```

Your test sends a PUT to /simulations/delay?duration=5s, runs the scenario it wants to test, then
sends a DELETE to reset. No restart needed.

The `track: true` option on the first handler is worth explaining. Mokapi's dashboard shows every HTTP
request and response along with all the JavaScript handlers that executed for it. But the first handler
here only updates internal state, it doesn't touch the response. Without `track: true`, Mokapi has no way
to know the handler was relevant to that request and won't show it in the dashboard. Setting it explicitly
keeps the full picture visible when you're debugging a test run.

**Error scenarios**. Force a 500, a 429 rate limit, or a malformed response body. Test that your application handles
each one gracefully before those conditions appear in production.

**Dynamic responses**. Return different data based on query parameters, headers, or request body content.
Simulate different user states, different account tiers, different permission levels, all from the same
mock server.

---

## What You Get Out of This

Tests that pass consistently because they don't depend on anything outside your control. A test suite that
covers error scenarios and edge cases that a live API can't reproduce on demand. A faster pipeline because
there's no network latency and no waiting on third-party response times.

And because Mokapi validates every request and response against your spec, you're also catching contract
violations in CI. If your application starts sending a request that doesn't match the spec, Mokapi tells
you immediately rather than letting it through silently.

The mock specs live in the repo, versioned alongside the code. When the API contract changes, the spec
changes, the mock changes, and the tests run against the updated contract automatically. That's the kind
of setup that keeps tests honest over time

---

## Getting Started

If you want to try this on your own project:

1. Add your OpenAPI or AsyncAPI specs to a /mocks directory in your repo
2. Add the Mokapi Docker step to your CI workflow
3. Point your application at localhost:80 instead of the real API during tests
4. Run your test suite and watch it stop depending on things outside your control

The [full CI/CD tutorial](/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline) walks through a complete
working example if you want to see all the pieces together.