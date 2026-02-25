---
title: "Guard Your API Contracts: Catch Breaking Changes Before Production"
description: Stop API drift in its tracks. Use Mokapi as a validation layer to enforce OpenAPI contracts between clients and backends
subtitle: Stop API drift in its tracks. Use Mokapi as a validation layer to enforce OpenAPI contracts between clients and backends, no matter who's calling or what they're building.
image:
    url: /mokapi-using-as-proxy.png
    alt: Flow diagram illustrating how Mokapi enforces OpenAPI contracts between clients, Playwright tests, and backend APIs.
tech: http
links:
  items:
    - title: Get Started with Mokapi
      href: /docs/get-started/installation
    - title: Record & Replay API Traffic
      href: /resources/blogs/record-and-replay-api-interactions
---

# Guard Your API Contracts: Catch Breaking Changes Before Production

In modern distributed systems, APIs are everywhere, frontend to backend, service
to service, mobile apps, test automation, third-party integrations. Each interaction
relies on a shared API contract, typically expressed through an OpenAPI specification.
But what happens when that contract drifts?

A renamed field breaks the mobile app. A missing validation lets bad data into production.
An unexpected status code crashes the frontend. Small deviations compound into debugging
nightmares, broken integrations, and slowed development velocity.

> What if every API interaction was automatically validated against your OpenAPI spec?
> That's exactly what Mokapi enables, a lightweight validation layer that sits between
> client and backend, enforcing contract compliance on both sides of every request.

## The Problem: API Drift is Inevitable

Even with the best intentions, API contracts drift from reality:

- **Backend changes without updating the spec:** A developer renames a field, changes a response structure, or adds a new required parameter, but forgets to update the OpenAPI documentation.
- **Frontend assumes behavior that was never specified:** A client sends malformed requests or expects fields that don't exist, and the backend silently accepts it (for now).
- **Microservices evolve independently:** Service A updates its contract, but Service B keeps calling the old version. Everything works until it doesn't.
- **Tests pass, but production breaks:** E2E tests mock the API instead of hitting the real backend, masking contract violations until deployment.

The result? Teams spend hours debugging "why is this suddenly broken?" instead of shipping features.

## The Solution: Mokapi as a Contract Guardian

Mokapi can sit between any client and backend to validate every request and response against your OpenAPI specification.
With a simple JavaScript forwarding script, Mokapi becomes a transparent validation layer that:

- Blocks invalid requests before they reach your backend
- Validates backend responses before they reach the client
- Provides clear, actionable error messages when violations occur
- Works with browsers, test frameworks, mobile apps, and service-to-service traffic

![Flow diagram showing Mokapi positioned between clients and backend APIs, validating requests and responses](/mokapi-using-as-proxy.png "Mokapi validates all traffic against your OpenAPI spec, catching contract violations before they cause problems")

## How It Works: Request Forwarding with Validation

The core concept is simple: Mokapi intercepts HTTP traffic, validates it against your OpenAPI spec,
forwards valid requests to the backend, validates the response, and only then returns it to the client.

Here's the complete forwarding and validation script:

```typescript
import { on } from 'mokapi';
import { fetch } from 'mokapi/http';

/**
 * Forward incoming requests to backend services while validating
 * both requests and responses against OpenAPI specifications.
 */
export default async function () {
    
    on('http', async (request, response) => {

        // Map request to backend URL based on OpenAPI spec name
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

            // Copy status code and headers
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
                return `https://backend2.example.com${request.url.path}?${request.url.query}`;
            }
            default:
                return undefined;
        }
    }
}
```

### The Four-Step Validation Flow

#### <i class="bi bi-1-circle-fill align-baseline"></i> Validate Incoming Request

Mokapi checks the HTTP method, URL parameters, headers, and request body against your OpenAPI spec.
Invalid requests are blocked with clear error messages.

#### <i class="bi bi-2-circle-fill align-baseline"></i> Forward Valid Requests

Only requests that pass validation are forwarded to the backend. No changes to your backend code
or infrastructure are required.

#### <i class="bi bi-3-circle-fill align-baseline"></i> Validate Backend Response

Mokapi validates the backend's response, status codes, headers, and response body against the OpenAPI specification.

#### <i class="bi bi-4-circle-fill align-baseline"></i> Return Validated Response

Only responses that match the contract reach the client, guaranteeing end-to-end contract fidelity across every interaction.

## Where to Use Mokapi for Contract Validation

Mokapi's forwarding and validation capabilities work in multiple scenarios across your architecture:

### Frontend ↔ Backend

Place Mokapi between your frontend and backend to catch contract violations during development:
- Automatic request and response validation
- Immediate detection of breaking changes
- Backend and OpenAPI spec evolve together
- Fewer "why is the frontend broken?" debugging loops

### Service ↔ Service

In microservice architectures, API drift between services causes instability. Mokapi provides:
- Strict contract enforcement between services
- Early detection of incompatible changes
- Stable integrations as teams evolve independently
- Clear validation errors during development and CI

### Playwright Tests ↔ Backend

One of the most powerful setups: Playwright → Mokapi → Backend
- CI fails immediately when backend breaks the contract
- Tests interact with real backend, not mocks
- Validation errors are clear and actionable
- Tests stay simpler, no manual validation needed

### Kubernetes Test Environments

Deploy Mokapi as a sidecar or standalone validation layer in preview environments:
- Consistent contract validation for cluster traffic
- Early detection before staging deployment
- No modifications to backend services
- Integrates with Helm charts and GitOps workflows

## Why Teams Choose This Approach

``` box=feature title="Automatic Contract Enforcement"
Every interaction is validated against your OpenAPI spec. Your backend can no longer silently drift from the contract.
```

``` box=feature title="Immediate Detection of Breaking Change"
Issues are caught early, renamed fields, wrong formats, unexpected status codes, mismatched data types before they reach production.
```

``` box=feature title="More Reliable Frontend Development"
Frontend teams get consistent, validated API responses with fewer sudden breaking changes and a smoother development workflow.
```

``` box=feature title="Better Cross-Team Collaboration"
Backend developers instantly see contract violations. Frontend engineers get stable APIs. QA gets reliable test environments. Platform teams reduce deployment risk.
```

``` box=feature title="Smooth Mock-to-Real Transition"
Start with mocked endpoints in early development. Later, simply forward requests to the real backend while keeping validation in place.
```

``` box=feature title="Actionable Error Messages"
When validation fails, Mokapi provides clear, detailed error messages showing exactly what doesn't match the spec and why.
```

> "Mokapi becomes your always-on API contract guardian, lightweight, transparent, and spec-driven."

## Getting Started

Implementing Mokapi as a validation layer is straightforward:

1. **Define your OpenAPI specification** for the API you want to validate
2. **Create a forwarding script** using the example above, mapping `request.api` to your backend URLs
3. **Run Mokapi** between your client and backend (locally, in Docker, or in Kubernetes)
4. **Point your client** at Mokapi instead of directly at the backend
5. **Watch validation errors surface** in real-time as you develop, test, and deploy

No changes to your backend code. No infrastructure overhaul. Just a transparent validation layer
that enforces your API contract at runtime.

## Stop API Drift. Start Enforcing Contracts.

With Mokapi as a validation layer, you get automatic contract enforcement, early detection of breaking
changes, and stable multi-team integration, all without touching your backend code.

{{ cta-grid key="links" }}

<p class="text-center text-muted mt-3 pt-3" style="font-size: 0.9rem;">Ready to guard your API contracts? Start validating today.</p>