---
title: End-to-End Testing with Mock APIs Using Mokapi
description: Learn how to improve your end-to-end testing with Mokapi by mocking APIs for faster, more reliable test runs. Discover how to integrate Mokapi into your CI/CD pipeline with GitHub Actions for better control and fewer bugs.
---

# End-to-End Testing with Mocked APIs Using Mokapi

Building reliable applications often depends on being able to test real-world scenarios — including how your app 
behaves when communicating with external APIs. However, relying on real APIs during development and CI/CD 
pipelines can introduce problems like slow test runs, flaky results, or even test failures when external 
services are unavailable.

That’s where mocking APIs comes in.

## Why Mock APIs for Testing?

Mocking APIs allows you to simulate external systems under controlled conditions. This means:

- ⚡ Faster and more reliable test runs 
- 🎯 Full control over the data and behavior returned by APIs 
- 🛠️ Easier testing of error scenarios, timeouts, and edge cases 
- 🔗 No dependencies on the availability of third-party services

With [Mokapi](https://mokapi.io), you can easily define API mocks using [OpenAPI](/docs/guides/http/overview.md) or 
[AsyncAPI](/docs/guides/kafka/overview.md) specifications, and serve them locally or in your CI environment.
Mokapi even supports dynamic behavior using simple [JavaScripts](/docs/javascript-api/overview.md), helping you create more realistic test scenarios.

## How Mokapi Fits into Your CI/CD Pipeline

Running Mokapi during your automated tests helps isolate your systems and catch integration bugs earlier.
Here's what a typical flow looks like:

```shell
Code Push → Start Mokapi → Run Tests Against Mocks → Stop Mokapi
```
You can even run Mokapi in your GitHub Actions workflows, making your CI/CD pipelines faster and more predictable.

## Example: GitHub Actions Workflow with Mokapi

Here's a basic example of using Mokapi in GitHub Actions:
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
          docker run -d --rm --name mokapi -p 80:80 -p 8080:8080 -v ${{ github.workspace }}/mocks:/mocks mokapi/mokapi:latest /mocks
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

In this setup, Mokapi runs as in a Docker container in GitHub Actions, making your mocked APIs available for the duration of your tests.

## Beyond the Basics: Advanced Mocking with Mokapi

Want even more control? Mokapi supports advanced mocking features, such as:

- **Simulating Timeouts** — Test how your app handles slow APIs
- **Dynamic Responses** — Return different data based on query parameters or headers
- **Error Scenarios** — Force 500 errors or custom error messages to test error handling

You can even modify mocks dynamically during a test run with simple [JavaScripts](/docs/javascript-api/overview.md)!

## Local Development and Mokapi

Mokapi isn't just for CI. You can also run it locally during development.
There are many ways to run Mokapi depending on your setup — learn more in the [Running Mokapi Guide](/docs/guides/get-started/running.md).

## Conclusion

Mocking APIs is a crucial part of building robust, scalable systems. With Mokapi, you can easily integrate mocking into your local 
development and CI pipelines, leading to faster feedback, fewer bugs, and better products.

**Ready to ship faster, more reliable software?**
[Get started with Mokapi](/docs/guides/get-started/welcome.md)

For a detailed, step-by-step guide on how to use Mokapi in your GitHub Actions workflows, 
see the [GitHub Actions and Mokapi](/docs/resources/tutorials/running-mokapi-github-action.md).

---

*A shorter summary of this article is also available on [LinkedIn](https://www.linkedin.com/pulse/end-to-end-testing-mocked-apis-using-mokapi-github-actions-lehmann-nuylf), including a link back here for full details.*

---