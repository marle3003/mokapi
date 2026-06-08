---
title: Running Mokapi in a CI/CD Pipeline
description: Step-by-step guide to running Mokapi in GitHub Actions to mock external APIs, validate contracts, and run reliable automated tests.
subtitle: This tutorial walks through a complete working example of running Mokapi in GitHub Actions. You'll define a mock API with an OpenAPI spec, write a Node.js backend that calls it, add a Jest test suite, and wire everything together in a CI workflow. By the end, your tests will run against a spec-validated mock with no external dependencies.
icon: bi-gear-wide-connected
---
# Running Mokapi in a CI/CD Pipeline

If you haven't installed Mokapi yet, follow the [Installation Guide](/docs/get-started/installation.md) first.
The full source code for this tutorial is on GitHub: [mokapi-ci-nodejs](https://github.com/marle3003/mokapi-ci-nodejs).

---

## What We're Building

The scenario is a Node.js backend that fetches user data from an external API. In production that's
a real service. In tests, Mokapi mocks it. The CI workflow starts Mokapi, runs the Jest test suite
against it, and stops Mokapi when done.

The project structure looks like this:

```text style=simple
/mokapi-ci-nodejs
 ├── .github/
 │   ├── workflows/
 │       ├── ci.yml       # GitHub Actions workflow
 ├── src/
 │   ├── client.js        # Handles external API requests
 ├── test/
 │   ├── api.test.js      # Jest tests
 ├── mocks/
 │   ├── users.yaml       # OpenAPI spec for the mocked API
 │   ├── users.js         # Mokapi script for dynamic responses
 ├── package.json
```

The `mocks/` directory lives in the repo alongside the code. That's intentional. When the API contract
changes, the spec changes, the mock changes, and the tests run against the updated contract automatically.
Everything stays in sync because it's all in one place.

---

## Step 1: Define the Mock API

Start with the OpenAPI spec. This is the contract Mokapi will enforce. Every request and response gets
validated against it automatically.

```yaml
openapi: "3.0.0"
info:
  title: Mock API
  version: "1.0.0"
paths:
  /users/{id}:
    get:
      summary: Get user info
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
        "404":
          description: User not found
```

---

## Step 2: Add a Script for Dynamic Responses

The spec defines the contract. The script defines the behavior. Mokapi's JavaScript API lets you intercept
requests and return different responses based on what came in.

```javascript
import { on } from 'mokapi'

export default function() {
  on('http', function(request, response) {
    switch (request.path['id']) {
      case '123':
        response.data = { id: '123', name: 'foo' }
      default:
        response.statusCode = 404
        response.data = null
    }
  })
}
```

User 123 gets a valid response. Anything else gets a 404. Setting `response.data` rather than `response.body`
tells Mokapi to validate the response against the schema before returning it. If the data doesn't match the spec,
Mokapi will tell you immediately.

For a full reference of what's available in scripts, see the [JavaScript API docs](/docs/javascript-api/overview.md).

---

## Step 3: The Node.js Backend

The backend reads the API base URL from an environment variable. In production that's the real service.
In tests, it points at Mokapi.

```javascript
const axios = require('axios');

const API_BASE_URL = process.env.API_URL || 'http://localhost';

async function getUser(id) {
  try {
    const response = await axios.get(`${API_BASE_URL}/users/${id}`);
    return response.data;
  } catch (error) {
    return { error: 'User not found' };
  }
}

module.exports = { getUser };
```

No special test mode. No conditionals. The application code doesn't know or care whether it's talking to
Mokapi or the real API.

---

## Step 4: The Jest Tests

The tests verify that the backend handles both the happy path and the not-found case correctly.

```javascript
const { getUser } = require('../src/client');

test('Fetch existing user', async () => {
  const user = await getUser('123');
  expect(user).toEqual({ id: '123', name: expect.any(String) });
});

test('Handle user not found', async () => {
  const user = await getUser('999');
  expect(user).toEqual({ error: 'User not found' });
});
```

The first test hits the mock for user `123` and expects a valid user object back. The second test hits
user `999`, which the script returns a 404 for, and expects the backend to handle it gracefully.
Both scenarios are deterministic because the mock controls what comes back.

---

## Step 5: The GitHub Actions Workflow

This is where it all comes together. The workflow starts Mokapi in Docker, runs the tests, and stops
Mokapi when done.

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
          docker run -d --rm --name mokapi \
            -p 80:80 \
            -p 8080:8080 \
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

A few things worth knowing here.

The `-v` flag mounts the mocks/ directory into the container so Mokapi picks up your spec and script.
If you add more specs later, they get picked up automatically without changing the workflow.

Port 80 serves API traffic. Port 8080 is Mokapi's dashboard and HTTP API, where you can inspect every
request and response that came through during the test run. Useful when a test fails, and you want to
see exactly what Mokapi received and returned.

The `sleep 5` gives the container a moment to be ready before the tests start. For more robust setups
you can poll Mokapi's health endpoint instead, but for most CI environments five seconds is plenty.

The `--rm` flag removes the container automatically when it stops, so you don't need to worry about cleanup
between runs.

---

## What You End Up With

A test suite that runs the same way every time, with no dependency on external services. The mock validates
every request and response against your OpenAPI spec, so if the backend starts sending malformed requests
or the spec changes without a corresponding code update, the CI pipeline catches it before it ships.

And because the specs and scripts live in the repo, they're part of the same review and versioning workflow
as everything else. Contract changes go through pull requests. Nothing drifts silently.

The full working example with all the files from this tutorial is on GitHub: 
[mokapi-ci-nodejs](https://github.com/marle3003/mokapi-ci-nodejs). Clone it, run it,
and use it as a starting point for your own setup.