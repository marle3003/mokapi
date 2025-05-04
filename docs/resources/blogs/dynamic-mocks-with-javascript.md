---
title: Create Smart API Mocks with Mokapi Scripts
description: Tired of static mocks? Learn how Mokapi Scripts let you create dynamic mock APIs using JavaScript — perfect for development, testing, and rapid prototyping.
---

# Bring Your Mock APIs to Life with Mokapi and JavaScript

<img src="/mokapi-scripts.png" alt="Example JavaScript code for Mokapi Scripts with annotated benefits of using dynamic API mocks">

Mocking APIs is essential for fast development — but static mocks can quickly 
become a bottleneck. Wouldn’t it be better if your mocks could think — 
reacting to queries, headers, or even generating data on the fly?

That's exactly what [Mokapi Scripts](/docs/javascript-api) are designed for.

With just a few lines of JavaScript, you can control how your mocks behave — 
making them dynamic, intelligent, and realistic.

## What Are Mokapi Scripts?

Mokapi Scripts are lightweight JavaScript modules that give you full control 
over how your mock APIs respond. Instead of static JSON, you define behavior 
based on request data — query strings, headers, body content, and more.

## Why Dynamic Mocks Matter

Static mock responses are fine for simple cases, but they quickly fall short when:

- Your frontend depends on different user roles (e.g., admin vs. regular user)
- You need to simulate errors, timeouts, or permission checks
- Backend state changes over time and should affect future responses (e.g., after a POST, the next GET reflects the update)
- You need data that changes depending on query parameters or request bodies
- You want to test workflows or sequences of API calls that depend on each other
- You're working on features like pagination, filtering, or sorting
- You need to simulate authentication and session-specific behavior
- You want to create more realistic test scenarios for CI pipelines or manual testing
- Your team needs fast feedback loops without relying on a fully working backend

Dynamic mocks make your development process more reliable, realistic, and efficient.

## What You Can Do with Mokapi Scripts

- ✅ Return different responses based on query parameters, headers, or body content
- ✅ Simulate authentication, authorization, and role-based access
- ✅ Generate random, structured, or context-aware dynamic data
- ✅ Mock complex workflows with conditional logic and stateful behavior
- ✅ Chain requests together to simulate real-world usage patterns
- ✅ Customize error responses, delays, and status codes

All using familiar JavaScript or TypeScript — no need to learn a new DSL.

## Example: Conditional Response Based on a Query Parameter

Let’s say your API returns a list of products. You want to simulate:

- A search operation when a query parameter (name) is provided 
- An error response when the query parameter is exactly "error"

Here’s how easy it is with Mokapi Scripts:

```typescript
import { on } from 'mokapi';

const products = [
    { name: 'Laptop Pro 15' },
    { name: 'Wireless Mouse' },
    { name: 'Mechanical Keyboard' },
    { name: 'Noise Cancelling Headphones' },
    { name: '4K Monitor' },
    { name: 'USB-C Hub' }
];

export default () => {
    on('http', (request, response): boolean => {
        if (request.query.name) {
            if (request.query.name === 'error') {
                response.body = 'A custom error message';
                response.statusCode = 400;
            } else {
                const matchingProducts = products.filter(p =>
                    p.name.toLowerCase().includes(request.query.name.toLowerCase())
                );
                response.data = {products: matchingProducts};
                return true;
            }
        }
        return false;
    });
}
```

### response.data vs. response.body
In Mokapi, you control the response with either `response.data` or `response.body`:

#### `response.data`

- Any JavaScript value (object, array, number, etc.)
- Mokapi:
    - ✅ Validates it against your OpenAPI specification.
    - ✅ Converts it to the correct format (JSON, XML, etc.)

Use this when you want automatic validation and formatting.

#### `response.body`
- Must be a string.
- Mokapi:
  - ❌ Skips validation
  - ✅ Gives you full control (e.g., raw HTML, plain text)

Use this when you want to simulate freeform or invalid content.

## Use Cases

### 1. Frontend Development

Test UI flows with realistic behavior — pagination, filtering, auth, and more — 
without waiting for backend implementation.

### 2. Testing and QA

Simulate edge cases, failures, and timeouts directly in your mock server — 
ideal for automated or manual testing.

### 3. Rapid Prototyping

Show real-world behavior in your prototypes using dynamic data. Build better 
demos and get faster feedback.

## Getting Started

1. Write your OpenAPI or AsyncAPI spec (or generate it)
2. Add a Script where you control the response with JavaScript
3. Run Mokapi — that’s it!

👉 Try the [OpenAPI Mocking Tutorial](/docs/resources/tutorials/get-started-with-rest-api) for a guided walkthrough.

👉 Check out the [Mokapi Installation Guide](/docs/guides/get-started/installation.md) to get set up in minutes.

## Conclusion

Static mocks are yesterday’s solution. \
Mokapi Scripts bring a new level of control, flexibility, and realism to your API development process — all powered by JavaScript.

No backend? No problem. \
With Mokapi, your mocks behave exactly the way you need them to.

## Further Reading

- [Debugging Mokapi JavaScript](/docs/resources/blogs/debugging-mokapi-scripts)\
  Learn how to debug your JavaScript code inside Mokapi
- [End-to-End Testing with Mock APIs Using Mokapi](/docs/resources/blogs/end-to-end-testing-with-mocked-apis)\
  Improve your end-to-end tests by mocking APIs with Mokapi.

---

*A shorter summary of this article is also available on [LinkedIn](https://www.linkedin.com/pulse/make-your-mocks-smarter-javascript-mokapi-marcel-lehmann-np3ef/), including a link back here for full details.*

---