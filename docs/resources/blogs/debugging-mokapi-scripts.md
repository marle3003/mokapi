---
title: Debugging Mokapi JavaScript
description: Learn how to debug your JavaScript code inside Mokapi using console.log, console.error, and event handler tracing.
subtitle: Master console logging, error tracking, and event handler tracing to build bulletproof Mokapi scripts
tags: [Development]
---

# Debugging Mokapi JavaScript

Mokapi makes it remarkably easy to mock HTTP APIs, LDAP servers, Kafka topics, and mail servers—all customizable through
JavaScript. But as your scripts grow more sophisticated, you need powerful debugging tools to understand what's
happening under the hood.

This article will show you how to leverage Mokapi's built-in debugging capabilities: `console.log`, `console.error`,
and advanced event handler tracing.

> Whether you're troubleshooting a failing integration or optimizing complex request flows, these debugging
> techniques will give you complete visibility into your Mokapi scripts.

## Console Output: Your First Line of Defense

| Method          | Purpose             | When to Use                                                    |
|-----------------|---------------------|----------------------------------------------------------------|
| console.log()   | General information | Tracing values, flow control, debug checkpoints                |
| console.error() | Error reporting     | Unexpected conditions, failures, validation errors             |
| console.warn()  | Warning messages    | Deprecations, potential issues, non-critical problems          |
| console.debug() | Verbose debugging   | Detailed internal state, variable dumps, step-by-step tracing  |

### Basic Console Logging

Use `console.log` to output general debug information:

```javascript
import { on } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    // Log simple messages
    console.log('Request received')
    
    // Log with context
    console.log('Handling request for:', request.url.path)
    
    // Log complex objects
    console.log('Request headers:', request.header)
    
    // Log formatted data
    console.log(`${request.method} ${request.url.path}`)
  })
}
```

### Error Logging Best Practices

Use `console.error` to report errors and unexpected behavior:

```javascript
import { on } from 'mokapi'
import { fetch } from 'mokapi/http'

export default function() {
  on('http', async (request, response) => {
    try {
      const res = await fetch('https://api.backend.com/data')
      
      if (res.statusCode !== 200) {
        // Log unexpected status codes
        console.error('Backend returned error:', res.statusCode)
        console.error('Response body:', res.body)
      }
      
      response.data = res.json()
      
    } catch (error) {
      // Log exceptions with full context
      console.error('Failed to fetch backend data:', error.toString())
      console.error('Request details:', {
        method: request.method,
        path: request.url.path,
        timestamp: new Date().toISOString()
      })
      
      response.statusCode = 500
      response.data = { error: 'Internal server error' }
    }
  })
}
```

## Event Handler Tracing: Deep Request Visibility

Console logging shows you what's happening inside your scripts, but sometimes you need to
understand the bigger picture: which handlers fired, in what order, and with what configuration.

That's where Mokapi's event handler tracing comes in.

### How Tracing Works

Mokapi can automatically track which event handlers process each request and display this
information in the dashboard. You control tracking behavior through the EventArgs configuration:

- **Auto-Detect**  
  Default behavior: tracks handlers that modify the response
- **Always Track**  
  Set track: true to monitor all executions
- **Never Track**  
  Set track: false to hide noisy handlers

### Enabling Handler Tracing

```javascript
import { on } from 'mokapi'

export default function() {
  // Always track required due no changes to response
  on('http', (request, response) => {
      console.log('Handling request for:', request.url.path)
  }, { 
    track: true, // Always visible in dashboard
  })
  
  // Auto-track admin routes (only when they modify the response)
  on('http', (request, response) => {
    if (request.url.path.startsWith('/admin')) {
      console.log('Admin path detected:', request.url.path)
      
      // Add admin-specific headers
      response.headers = response.headers || {}
      response.headers['X-Admin-Route'] = ['true']
    }
  })
  
  // Never track health checks
  on('http', (request, response) => {
    if (request.url.path === '/health') {
      response.statusCode = 200
      response.data = { status: 'ok' }
    }
  }, {
    track: false, // Never show in dashboard
  })
}
```

### Dashboard Visualization

When handlers are tracked, the Mokapi dashboard displays detailed information for each request:

<img src="/javascript-debug-event-handler.png" alt="Mokapi dashboard displaying event handler execution details for a specific HTTP request">

## Debug with Confidence

With console logging, event handler tracing, and strategic use of the track parameter, you have complete visibility
into your Mokapi scripts. Whether you're building complex mock logic or troubleshooting production issues, these
debugging tools will help you understand exactly what's happening at every step.