---
title: Bring Your Mock APIs to Life with JavaScript
description: Tired of static mocks? Mokapi Scripts let you create dynamic, intelligent mock APIs that react to real request data, powered by plain JavaScript.
subtitle: Tired of static mocks? Mokapi Scripts let you create dynamic, intelligent mock APIs that react to real request data, powered by plain JavaScript.
---

# Bring Your Mock APIs to Life with JavaScript

Mocking APIs is essential for fast development, but static mocks can quickly become a bottleneck. What if your mock 
could think? Reacting to query parameters, headers, or body content. Simulating auth, errors, and pagination. 
Reflecting state changes across requests.

That's exactly what Mokapi Scripts are designed for. With just a few lines of JavaScript, you can turn a flat 
JSON file into a dynamic, intelligent mock API.

> No backend? No problem. With Mokapi Scripts, your mocks behave exactly the way you need them to, all in 
> familiar JavaScript or TypeScript, with no new DSL to learn.

## What Are Mokapi Scripts?

Mokapi Scripts are lightweight JavaScript modules that sit alongside your OpenAPI or AsyncAPI specification.
Instead of returning a fixed response, they let you define behavior, inspecting the incoming request and
deciding what to send back.

They're the difference between a mock that says *"here's a user"* and one that says *"here's the right user, 
given who's asking what role they have, and what they just posted."*

## Why Static Mocks Fall Short

Static responses are fine for trivial cases. But real development quickly surfaces their limits:

- **Role-based responses**  
  Admin and regular users see different data. A static mock can only show one.
- **Simulating errors & timeouts**  
  You need your frontend to handle 403s, 429s, and network failures, but your mock always returns 200.
- **Stateful workflows**  
  After a POST, the next GET should reflect the change. Static mocks have no memory.
- **Dynamic filtering & pagination**    
  Query parameters like `?page=2` or `?name=laptop` should produce meaningful results.
- **Sequential request chains**  
  Login → fetch profile → update settings: static mocks can't model these flows.
- **Auth & session behavior**  
  Missing or invalid tokens should behave differently from valid ones in your test environment.

Dynamic mocks solve all of these. They make your development process more reliable, your test suites more
realistic, and your feedback loops faster.

## What You Can Do with Mokapi Scripts

- Return different responses based on query parameters, headers, or body content
- Simulate authentication, authorization, and role-based access control
- Generate random, structured, or context-aware dynamic data on the fly
- Mock complex workflows with conditional logic and stateful behavior
- Chain requests together to simulate real-world usage patterns
- Customize error responses, status codes, and artificial delays
- Mock an HTML login form to simulate an external identity provider

## Example: Dynamic Product Search

Let's build a realistic product list endpoint. It should:
- Return the full product catalog when no filter is applied
- Filter products by name when a ?name= query parameter is provided
- Return a `400` error with a custom message when `?name=error` is passed

```javascript title=products.js
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
            // No filter — return everything
            response.data = { products: products }
            return
        }

        if (nameFilter === 'error') {
            // Simulate a validation error
            response.rebuild(400)  
            response.data.message = 'A custom error message'
            return
        }

        // Filter products by name (case-insensitive)
        const matched = products.filter(p =>
            p.name.toLowerCase().includes(nameFilter.toLowerCase())
        )
        
        response.data = { products: matched }
  })
}
```

## Use Cases

- **Frontend Development**  
  Test UI flows with realistic behavior—pagination, filtering, auth, and error states—without waiting for a working backend.
- **Testing & QA**  
  Simulate edge cases, failures, and timeouts directly in your mock server. Ideal for automated CI pipelines and manual exploratory testing.
- **Rapid Prototyping**  
  Show real-world behavior in demos and prototypes using dynamic data. Build faster, get feedback sooner.

## Getting Started

1. **Write your specification**  
   Start with an OpenAPI spec, or generate one from an existing service. This defines the shape of your API.
2. **Add a Mokapi Script**  
   Drop a JavaScript file next to your spec. Register event handlers that inspect the request and set the response however you need.
3. **Run Mokapi**  
   That's it. Your dynamic mock is live, validated against your spec, and visible in the Mokapi dashboard.

## Further Reading

- [Debugging Mokapi JavaScript](/resources/blogs/debugging-mokapi-scripts)  
  Learn how to use `console.log`, `console.error`, and event handler tracing to see exactly what your scripts are doing.
- [End-to-End Testing with Mock APIs Using Mokapi](/resources/blogs/end-to-end-testing-with-mocked-apis)  
  Improve your end-to-end test suites by replacing live backend dependencies with Mokapi.

---

*A shorter summary of this article is also available on [LinkedIn](https://www.linkedin.com/pulse/make-your-mocks-smarter-javascript-mokapi-marcel-lehmann-np3ef/), including a link back here for full details.*

---