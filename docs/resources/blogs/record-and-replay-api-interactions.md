---
title: "Record & Replay: API Interactions with Mokapi"
description: Capture real-world API traffic for testing, debugging, and offline development—all with a simple JavaScript script
subtitle: Capture real-world API traffic for testing, debugging, and offline development—all with a simple JavaScript script
tech: http
tags: ['HTTP']
---

# Record & Replay: API Interactions with Mokapi

In a [previous article](ensuring-api-contract-compliance-with-mokapi.md), we explored how Mokapi can act as an API 
specification guard between services—validating requests and responses against your OpenAPI specs in real-time. 
But Mokapi's capabilities go far beyond validation. In this article, we'll dive into another powerful use case: 
recording and replaying API interactions.

Imagine capturing real API traffic from your production or staging environment and replaying it later for testing,
demos, or offline development. With Mokapi Script, this becomes remarkably straightforward.

## Why Record API Interactions?

Before we jump into the code, let's understand the value of recording API traffic:

- **Complex User Scenarios**  
  Capture multistep workflows with real data to reproduce edge cases and validate business logic against actual usage patterns.
- **Regression Testing**  
  Build a suite of real-world request/response pairs to ensure new changes don't break existing functionality.
- **Demos & Presentations**  
  Replay realistic API interactions without depending on live services—perfect for sales demos or conference presentations.
  
## The Recording Workflow

At a high level, the recording workflow looks like this:

<img src="/recording.png" alt="Mokapi Record & Replay Workflow">

## Building the Recording Script

Let's build a script that records all HTTP traffic passing through Mokapi. The script will capture both requests
and responses, storing them in a JSON file for later replay.

### The Complete Recording Script

```javascript
import { on } from 'mokapi'
import { read, writeString } from 'mokapi'

export default function() {
  on('http', (request, response) => {
    // Initialize an empty array to hold recorded interactions
    let data = []
    
    try {
      // Attempt to read existing recordings from file
      const s = read('./recordings.json')
      data = JSON.parse(s)
    } catch {}
    
    // Append the current request/response pair
    data.push({ request, response })
    
    // Write the updated recordings back to file
    writeString('./recordings.json', JSON.stringify(data, null, 2))
  }, { 
    // Use priority -1 to ensure this runs AFTER response is populated
    priority: -1 
  })
}
```

``` box=warning title="Understanding the Priority Setting"
The priority: -1 parameter is crucial here. It ensures this handler runs after the response has been
fully populated by other handlers (like your forwarding script or mock generators).
```

### How It Works

Let's break down what's happening in this compact script:

- **Event Listener:**  The on('http', ...) function registers a handler that fires for every HTTP request passing through Mokapi.
- **Read Existing Data:** We attempt to read recordings.json to retrieve previously recorded interactions. If the file doesn't
  exist (first run), the try-catch silently handles it.
- **Append New Data:** The current request/response pair is added to the array.
- **Persist to Disk:** The entire array is serialized to JSON and written back to the file.

## Combining Recording with Forwarding

The real power emerges when you combine recording with the API forwarding pattern from our previous article. 
Here's how they work together:

```javascript
import { on } from 'mokapi'
import { fetch } from 'mokapi/http'
import { read, writeString } from 'mokapi'

export default async function() {
  // FIRST: Forward requests to real backend (default priority: 0)
  on('http', async (request, response) => {
    const url = getForwardUrl(request)
    
    if (!url) {
      response.statusCode = 500
      response.body = 'Unknown backend'
      return
    }
    
    try {
      const res = await fetch(url, {
        method: request.method,
        body: request.body,
        headers: request.header,
        timeout: '30s'
      })
      
      response.statusCode = res.statusCode
      response.headers = res.headers
      
      const contentType = res.headers['Content-Type']?.[0] || ''
      if (contentType.includes('application/json')) {
        response.data = res.json()
      } else {
        response.body = res.body
      }
    } catch (e) {
      response.statusCode = 500
      response.body = e.toString()
    }
  })
  
  // SECOND: Record the complete request/response (priority: -1)
  on('http', (request, response) => {
    let data = []
    try {
      data = JSON.parse(read('./recordings.json'))
    } catch {}
    
    data.push({ request, response })
    writeString('./recordings.json', JSON.stringify(data, null, 2))
  }, { priority: -1 })
  
  // Helper function to determine backend URL
  function getForwardUrl(request: HttpRequest): string | undefined {
    switch (request.api) {
      case 'backend-1':
        return `https://backend1.example.com${request.url.path}?${request.url.query}`
      case 'backend-2':
        return `https://backend2.example.com${request.url.path}?${request.url.query}`
      default:
        return undefined
    }
  }
}
```

This setup gives you the best of both worlds: real backend responses validated against your OpenAPI specs,
plus a growing library of recorded interactions for later use.

``` box=info
When Mokapi forwards and records traffic, it still validates requests and
responses against the OpenAPI specification whenever possible. This means you
only record interactions that conform to your API contract—making replays
reliable and spec-compliant.
```

## Building the Replay Script

Now comes the fun part: replaying your recorded interactions! Here's a script that reads the recorded data
and matches incoming requests to previously captured responses:

```javascript
import { on } from 'mokapi'
import { read } from 'mokapi'

export default function() {
  // Load all recorded interactions at startup
  let recordings = []
  try {
    const data = read('./recordings.json')
    recordings = JSON.parse(data)
  } catch (e) {
    console.error('Failed to load recordings:', e)
    return
  }
  
  on('http', (request, response) => {
    // Find matching recorded request
    const match = recordings.find(r => 
      r.request.method === request.method &&
      r.request.url.path === request.url.path &&
      r.request.url.query === request.url.query
    )
    
    if (match) {
      // Replay the recorded response
      response.statusCode = match.response.statusCode
      response.headers = match.response.headers
      response.body = match.response.body
      response.data = match.response.data
    } else {
      // No recording found
      response.statusCode = 404
      response.body = 'No recording found for this request'
    }
  })
}
```

### Enhancing the Replay Logic

The basic matching logic above works for simple cases, but you might want to enhance it for production use:

- **Ignore specific headers**: Filter out timestamp headers or request IDs that change between requests
- **Match on request body**: For POST/PUT requests, compare the request body to find the right response
- **Fuzzy matching:** Match path patterns instead of exact paths (e.g., /users/:id)
- **Sequential playback:** Replay requests in the order they were recorded for complex workflows

## Practical Use Cases

### 1. Building Regression Test Suites

Record interactions from your staging environment, then replay them in CI/CD pipelines to ensure new
code doesn't break existing functionality:

> Record once in staging → Replay thousands of times in CI/CD → Catch regressions early

### 2. Offline Development Environments

Developers can work without VPN access or network connectivity. Simply record a comprehensive set of
API interactions once, and your entire team can develop offline using the replay script.

### 3. Demo Environments

Create polished demos with predictable responses. No more crossed fingers hoping the backend behaves
during that critical sales pitch!

### 4. Performance Testing

Replay recorded traffic to load-test your services without hitting real backends. Modify the
replay script to simulate concurrent users.

## Best Practices & Tips

- **Sanitize sensitive data:** Before committing recordings to version control, strip out authentication tokens, personal information, and other sensitive data.
- **Organize by scenario:** Instead of one giant recordings.json, create separate recording files for different user journeys or test scenarios.
- **Version your recordings:** As your API evolves, maintain recordings for different API versions to support backward compatibility testing.
- **Combine with validation:** Use Mokapi's OpenAPI validation alongside recording to ensure you're capturing valid interactions.

## Taking It Further

The recording and replay pattern opens up even more possibilities:

- **Record/Replay UI:** Build a simple web interface to browse, filter, and selectively replay specific recordings
- **Dynamic modification:** Modify responses on-the-fly during replay to test error handling or edge cases
- **Traffic analysis:** Analyze recorded traffic to identify patterns, optimize API design, or detect anomalies
- **Mock generation:** Use recorded interactions to automatically generate OpenAPI examples or mock data

## Start Recording Today

By combining OpenAPI validation, HTTP forwarding, and file-based recording,
Mokapi becomes more than a mock server—it becomes a control plane for API
interactions across the entire API lifecycle.

> Record real user behavior once. Replay it endlessly. Evolve your APIs with confidence.