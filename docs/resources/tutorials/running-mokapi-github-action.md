---
title: "Running Mokapi in CI/CD: Mock APIs for Automated Testing"
description: Learn how to use Mokapi in CI/CD pipelines to mock APIs, automate tests, and ensure reliable backend validation without live dependencies.
icon: bi-gear-wide-connected
---
# Running Mokapi in a CI/CD Pipeline

Integrating Mokapi into a CI/CD pipeline ensures API contracts and Kafka topics interactions are validated before deployment. 
This helps catch issues early, making the development process more reliable and reducing the risk of breaking changes in production.
If you haven't installed Mokapi yet, follow the [Installation Guide](/docs/guides/get-started/installation.md) to get started.

This guide explains how to:
 
- Run Mokapi in a CI/CD pipeline (GitHub Actions).
- Mock an HTTP API during automated tests.
- Make sure the API is used according to the OpenAPI contract.

The source code for this tutorial is available in the GitHub repository [mokapi-ci-nodejs](https://github.com/marle3003/mokapi-ci-nodejs).

## Scenario

- The Node.js backend interacts with an external API.
- Mokapi is used to mock the API locally and in GitHub Actions.
- Jest (or any test framework) validates the backend's response.

## Project Structure

```
/my-node-app
 ├── .github/
 │   ├── workflows/
 │       ├── ci.yml  # GitHub Actions workflow
 ├── src/
 │   ├── client.js # Handles external API requests
 ├── test/
 │   ├── api.test.js # Jest test for API interactions
 ├── mocks/
 │   ├── users.yaml # OpenAPI specification for external API
 │   ├── users.js # Mokapi Script to set response data
 ├── package.json
```

## 1. Define the Mock API (OpenAPI Spec)

Create users.yaml to define the mock API that the backend calls.

```yaml tab=users.yaml
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

### Use Mokapi Script to Set Response

Mokapi allows customizing API responses dynamically using scripts. The following example shows how to intercept HTTP requests and return different responses based on the request path.
For a complete list of all available API endpoints, refer to our [API Reference](/docs/javascript-api/overview.md).

```javascript tab=users.js
import { on } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        console.log(request)
        switch (request.path['id']) {
            case '123':
                response.data = { id: '123', name: 'foo' }
                return true
            default:
                response.statusCode = 404
                response.data = null
                return true
        }
    })
}
```

## 2. Backend Code: Calling the API

The backend application fetches user data from the (mocked) API.

```javascript tab=client.js
const axios = require('axios');

const API_BASE_URL = process.env.API_URL || "http://localhost";

async function getUser(id) {
  try {
    const response = await axios.get(`${API_BASE_URL}/users/${id}`);
    return response.data
  } catch (error) {
    return { error: "User not found" };
  }
}

module.exports = { getUser };
```

## 3. Jest Test: Validating API Response

The test verifies that the backend properly interacts with the mock API.

```javascript tab=api.test.js
const { getUser } = require("../src/client");

test("Fetch existing user", async () => {
  const user = await getUser("123");
  expect(user).toEqual({ id: "123", name: expect.any(String) });
});

test("Handle user not found", async () => {
  const user = await getUser("999");
  expect(user).toEqual({ error: "User not found" });
});
```

## 4. GitHub Actions Workflow: Running Tests with Mokapi

- Starts Mokapi in GitHub Actions as a mock API server.
- Runs Jest tests, ensuring the backend interacts correctly with the mocked API.
- Stops Mokapi after testing to clean up the environment.

```yaml tab=ci.yaml
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