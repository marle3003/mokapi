---
title: Build Fullstack Apps with Mokapi | Mock APIs for Dev Speed
description: Combine frontend, backend, and Mokapi to mock third-party APIs locally, speed up dev cycles, and eliminate flaky tests.
icon: bi-layers
---

# Build Fullstack Apps Faster with Mokapi

**Combining Frontend + Backend + Mokapi in Local Development**

In a typical fullstack setup, frontend and backend developers often rely on third-party APIs—whether for authentication, payment processing, or external data. But these APIs can introduce friction: they might be unreliable, rate-limited, or unavailable in early development stages. That’s where Mokapi comes in.

In this post, we’ll walk through a real-world setup where frontend and backend devs use Mokapi locally to simulate third-party APIs, speed up iteration, and avoid flaky tests.

## Why Use Mokapi in Local Dev?

- Frontend devs can work independently without waiting for the backend or third-party APIs. 
- Backend devs can simulate APIs and focus on integrating core logic. 
- QA and testers can rely on predictable, reproducible API responses. 
- CI pipelines can run end-to-end tests without depending on unstable external services.

## The Setup: Frontend + Backend + Mokapi
Let’s say your app depends on a payment API, a weather API, and a custom internal service.

### Local Development Stack

- Frontend: Vue.js app served via Vite 
- Backend: Node.js or Go service 
- Mokapi: Mocks external services using OpenAPI specs or custom scripts

```shell
📁 /project
├── frontend/          # Vue.js app
├── backend/           # Express.js or Go service
└── mokapi/            # OpenAPI/AsyncAPI mocks
```

## Step-by-Step: How It All Works

### 1. Define Mocked APIs
   Use OpenAPI or AsyncAPI files in your /mokapi folder. You can even script behavior (e.g., delays, errors) using Mokapi Scripts.

```yaml
# weather-api.yaml
paths:
  /forecast:
    get:
      responses:
        200:
          content:
            application/json:
              example:
                city: "Berlin"
                temp: "18°C"
```

### 2. Configure Frontend and Backend
   In your .env or config files, point your services to the mocked endpoints.

```shell
VITE_WEATHER_API=http://localhost:3000/weather
BACKEND_PAYMENT_API=http://localhost:3000/payment
```

### 3. Run Everything Together
   You can use a docker-compose, npm run dev, or a simple script to spin up all services:

```shell
npm --prefix frontend run dev &
npm --prefix backend start &
mokapi ./mokapi &
```

## Benefits in Practice

- ⚡ Speed Up Iteration\
No need to wait for real APIs to be live or available. Frontend devs can build against predictable mock data from day one.

- 🧪 Reduce Flaky Tests\
By mocking third-party services, your CI tests become stable and repeatable.

- 🔍 Debug Easier\
Mokapi’s dashboard shows incoming requests and responses—making it easy to see what’s going wrong.

- 👫 Better Collaboration\
Frontend and backend teams can develop in parallel, with fewer blockers and better sync.

## Final Thoughts

Mokapi simplifies local development by making API mocking seamless and powerful. Whether you're testing edge cases, working offline, or just want to speed things up.