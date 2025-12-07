---
title: Ensure API Contract Compliance with Mokapi Validation
description: Validate HTTP API requests and responses with Mokapi to catch breaking changes early and keep backend implementations aligned with your OpenAPI spec.
image:
    url: /mokapi-using-as-proxy.png
    alt: Flow diagram illustrating how Mokapi enforces OpenAPI contracts between clients, Playwright tests, and backend APIs.
---

# Ensuring Compliance with the HTTP API Contract Using Mokapi for Request Forwarding and Validation

In modern distributed systems, APIs are everywhere — frontend-to-backend,
backend-to-backend, microservices communicating internally, mobile apps, test
automation tools, and more. Each interaction relies on a shared API contract,
often expressed through an OpenAPI specification. Even small
deviations can introduce bugs, break integrations, or slow down development.

By placing Mokapi between a client and a backend, you can ensure that every
**request and response adheres to your OpenAPI specification**. With a few lines
of JavaScript, Mokapi can forward requests to your backend while validating both
sides of the interaction. This provides a powerful way to enforce API correctness — 
whether the client is a browser, Playwright tests, your mobile app, or even
another backend service.

In this article, I explore how Mokapi can act as **a contract-enforcing validation layer**
and why this approach benefits frontend developers, backend teams, QA engineers,
and platform engineers alike.

<img src="/mokapi-using-as-proxy.png" alt="Flow diagram illustrating how Mokapi enforces OpenAPI contracts between clients, Playwright tests, and backend APIs.">

## How to Use Mokapi for API Validation with Request Forwarding?

Mokapi cannot only be used for mocking APIs, but it can also sit between any 
consumer and a backend service to validate real traffic. Using a small
JavaScript script, Mokapi can forward requests to your backend and
validates both requests and responses.


    Consumer (Frontend, Playwright, Microservice) → Mokapi → Backend API

```typescript
import { on } from 'mokapi';
import { fetch } from 'mokapi/http';

/**
 * This script demonstrates how to forward incoming HTTP requests
 * to a real backend while letting Mokapi validate responses according
 * to your OpenAPI spec.
 *
 * The script listens to all HTTP requests and forwards them based
 * on the `request.api` field. Responses from the backend are
 * validated when possible, and any errors are reported back to
 * the client.
 */
export default async function () {

    /**
     * Register a global HTTP event handler.
     * This function is called for every incoming request.
     */
    on('http', async (request, response) => {

        // Determine the backend URL to forward this request to
        const url = getForwardUrl(request)

        // If no URL could be determined, return an error immediately
        if (!url) {
            response.statusCode = 500;
            response.body = 'Failed to forward request: unknown backend';
            return;
        } 
            
        try {
            // Forward the request to the backend
            const res = await fetch(url, {
                method: request.method,
                body: request.body,
                headers: request.header,
                timeout: '30s'
            });

            // Copy status code and headers from the backend response
            response.statusCode = res.statusCode;
            response.headers = res.headers

            // Check the content type to decide whether to validate the response
            const contentType = res.headers['Content-Type']?.[0] || '';

            if (contentType.includes('application/json')) {
                // Mokapi can validate JSON responses automatically
                response.data = res.json();
            } else {
                // For other content types, skip validation
                response.body = res.body;
            }
            
        } catch (e) {
            // Handle any errors that occur while forwarding
            response.statusCode = 500;
            response.body = e.toString();
        }
    });

    /**
     * Maps the incoming request to a backend URL based on the API name
     * defined in the OpenAPI specification (`info.title`).
     * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httprequest
     *
     * @param request - the incoming Mokapi HTTP request
     * @returns the full URL to forward the request to, or undefined
     */
    function getForwardUrl(request: HttpRequest): string | undefined {
        switch (request.api) {
            case 'backend-1': {
                return `https://backend1.example.com${request.url.path}?${request.url.query}`;
            }
            case 'backend-2': {
                return `https://backend1.example.com${request.url.path}?${request.url.query}`;
            }
            default:
                return undefined;
        }
    }
}
```

For each interaction, Mokapi performs four important steps:

### 1. Validates incoming requests

Mokapi checks every incoming request against your OpenAPI specification:

- HTTP method
- URL & parameters
- headers
- request body

If the client sends anything invalid, Mokapi blocks it and returns a clear 
validation error.

### 2. Forwards valid requests to your backend

If the request is valid, Mokapi forwards it unchanged to the backend using JavaScript.

- No changes are required in your backend.
- No additional infrastructure is necessary.

### 3. Validates backend responses

Once the backend responds, Mokapi validates the response against the OpenAPI specification:

- status codes 
- headers 
- response body

If something doesn't match the contract, Mokapi blocks it and sends a validation error back to the client.

### 4. Return the validated response to the client

Only responses that pass validation reach the client, guaranteeing contract fidelity end-to-end.

## Where You Can Use Mokapi for Request Forwarding and Validation

Mokapi’s forwarding and validation capabilities make it useful far beyond local development or Playwright scripting.

### Between Frontend and Backend

Placing Mokapi between your frontend and backend ensures:
- automatic request and response validation 
- immediate detection of breaking changes 
- backend and API specification evolve together
- fewer “why is the frontend broken?” debugging loops

Frontend developers can experiment with confidence, knowing the backend
cannot silently diverge from the published contract.

### Between Backend Services (Service-to-Service)

In microservice architectures, API drift between services is a frequent cause of instability.
Routing service-to-service traffic through Mokapi gives you:
- strict contract enforcement between services 
- early detection of incompatible changes 
- stable integrations even as teams evolve independently 
- clear validation errors during development and CI

Mokapi becomes a lightweight, spec-driven contract guardian across your backend ecosystem.

### In Automated Testing (e.g., Playwright)

This is one of the most powerful setups.

    Playwright → Mokapi → Backend

Benefits:
- CI fails immediately when the backend breaks the API contract 
- tests interact with the real backend, not mocks
- validation errors are clear and actionable 
- tests remain simpler — no need to validate everything in Playwright

Your tests are guaranteed to hit a backend that actually matches the API contract.

### In Kubernetes Test Environments

Mokapi can also be used in temporary or preview environments to ensure contract validation across the entire cluster.

In Kubernetes, Mokapi can be deployed as:
- a sidecar container
- a standalone validation layer in front of backend services
- a temporary component inside preview environments

This brings:
- consistent contract validation for all cluster traffic
- early detection of breaking API changes before staging
- contract enforcement without modifying backend services
- transparent operation — apps talk to Mokapi, Mokapi talks to the backend

You can integrate Mokapi into Helm charts, GitOps workflows, or test namespaces.

## Why Teams Benefit from Using Mokapi Between Client and Backend

### Automatic Contract Enforcement

Every interaction is validated against your OpenAPI specification. Your backend can no longer quietly drift from the contract.

### Immediate Detection of Breaking Changes

Issues are caught early, not just in staging or production, such as:
  - renamed or missing fields
  - wrong or inconsistent formats
  - unexpected status codes
  - mismatched data types

### More Reliable Frontend Development

Frontend teams get:
- consistent, validated API responses
- fewer sudden breaking changes
- a smoother development workflow

This reduces context-switching and debugging time.

### Better Collaboration Between Teams

With Mokapi validating both sides:
- backend developers instantly see when they violate the contract
- frontend engineers get stable, predictable APIs
- QA gets reliable test environments
- platform engineers reduce risk during deployments

Mokapi becomes a shared API contract watchdog across the organization.

### Smooth Transition from Mocks to Real Systems

Teams often start with mocked endpoints in early development. Later, they can simply begin forwarding requests to the
real backend—while keeping validation in place.

## Conclusion

Using Mokapi between frontend and backend, between backend services, or inside Kubernetes environments provides:
- strong contract enforcement
- automatic validation for every interaction
- early detection of breaking changes
- stable multi-team integration
- more reliable CI pipelines
- a smooth path from mocking to real backend validation

Mokapi ensures your API stays aligned with its specification, no matter how quickly your system evolves.