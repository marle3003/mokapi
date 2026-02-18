---
title: End-to-End Testing with Mocked APIs
description: Stop relying on flaky external services. Build faster, more reliable test pipelines with Mokapi-powered API mocks in your CI/CD.
subtitle: Stop relying on flaky external services. Build faster, more reliable test pipelines with Mokapi-powered API mocks in your CI/CD.
---

# End-to-End Testing with Mocked APIs

Building reliable applications depends on being able to test real-world scenarios, including how your app
behaves when communicating with external APIs. But relying on live APIs during development and CI/CD 
pipelines introduces problems: slow test runs, flaky results, and failures when external services are unavailable.

That's where API mocking comes in and specifically, where Mokapi transforms your testing workflow.

> With Mokapi, you can simulate external systems under controlled conditions using OpenAPI
> or AsyncAPI specifications. The result? Faster feedback, fewer bugs, and tests you can actually trust.

## Why Mock APIs for Testing?

Mocking APIs gives you control over the testing environment in ways that real external services simply can't provide:

``` box=benefits title="Faster Test Runs"
No network latency, no waiting for slow third-party responses. Tests complete in seconds, not minutes.
```

``` box=benefits title="Full Control"
Return exactly the data you need. Simulate specific states, edge cases, and user scenarios on demand.
```

``` box=benefits title="Test Error Scenarios"
Force 500 errors, timeouts, malformed responses, and rate limits without breaking real services.
```

``` box=benefits title="Zero Dependencies"
Tests run offline, in CI, or on developer machines—no reliance on third-party availability or credentials.
```

Mokapi makes this straightforward: define your API contracts with OpenAPI or AsyncAPI specifications,
add optional JavaScript for dynamic behavior, and you're done. Your mocks are validated, versioned,
and ready to serve.

## How Mokapi Fits Into Your CI/CD Pipeline

Integrating Mokapi into your automated test pipeline isolates your system under test
and catches integration bugs early. The workflow is simple:

<img src="/ci-pipeline-flow.png" alt="Typical CI/CD Flow with Mokapi" />

Mokapi runs as a service during your test phase—whether in Docker, Kubernetes, or directly on the CI runner.
Your tests hit the mocked endpoints instead of real APIs, and Mokapi validates every response against your
specifications.

## Example: GitHub Actions Workflow with Mokapi

Here's a practical example showing how to integrate Mokapi into a GitHub Actions workflow for a Node.js backend:

```yaml tab=.github/workflows/test.yml
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
          sleep 5  # Ensure Mokapi is running

      - name: Set Up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20

      - name: Install Dependencies
        run: npm install

      - name: Run Tests
        run: npm test

      - name: Stop Mokapi
        run: docker stop mokapi
```

### What's Happening Here?

- **Checkout:** The workflow pulls your code, including your mock definitions stored in `/mocks`
- **Start Mokapi:** A Docker container runs Mokapi, mounting your mock specifications and exposing port 80 (API)
- **Set Up Node.js:** The test environment is configured with Node.js 20
- **Install & Test:** Dependencies are installed, and the test suite runs, hitting Mokapi instead of real APIs
- **Cleanup:** Mokapi stops automatically, keeping the CI environment clean

## Advanced Mocking with Mokapi

Basic mocking gets you far, but Mokapi's advanced features let you simulate complex real-world scenarios:

``` box=feature title="Simulating Timeouts"
Test how your app handles slow APIs by adding artificial delays to responses
```

``` box=feature title="Dynamic Responses"
Return different data based on query parameters, headers, or request body content
```

``` box=feature title="Error Scenarios"
Force 500 errors, 429 rate limits, or custom error messages to validate error handling
```

These capabilities are powered by Mokapi Scripts—lightweight JavaScript modules that let you 
define custom behavior. Here's a quick example for delay simulation:

```javascript tab=simulations.js
import { on, sleep } from 'mokapi'

export default () => {
    let delay = undefined
    
    on('http', (request, response) => {
        if (request.key === '/simulations/delay') {
            switch (request.method) {
                case 'PUT': 
                    delay = request.query.duration;
                    break;
                case 'DELETE':
                    delay = undefined;
                    break;
            }
        }
    })
    on('http', (request, response) => {
        if (!delay) {
            sleep(delay);
        }
    }, { track: true } )
}
```

This script demonstrates how to use Mokapi Scripts to dynamically simulate response delays at runtime using HTTP requests.

It allows you to enable, change, or disable artificial latency without restarting Mokapi.

## Ready to ship faster, more reliable software?

- [Get started with Mokapi](/docs/get-started/installation.md)
- [Full GitHub Actions Tutorial](/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline).


---

*A shorter summary of this article is also available on [LinkedIn](https://www.linkedin.com/pulse/end-to-end-testing-mocked-apis-using-mokapi-github-actions-lehmann-nuylf), including a link back here for full details.*

---