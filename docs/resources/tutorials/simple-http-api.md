---
title: Get started with REST API
description: Learn how to mock a REST API using an OpenAPI specification with Mokapi. This tutorial covers basic setup, configuration, and Docker deployment.
subtitle: Learn how to mock a REST API using an OpenAPI specification with Mokapi. This tutorial covers basic setup, configuration, and Docker deployment.
icon: bi-globe
tech: http
cards:
  items:
    - title: Dynamic Responses
      href: /resources/blogs/bring-your-mock-apis-to-life-with-javascript
      description: Return different data based on query parameters, headers, or request body content
    - title: Debugging
      href: /resources/blogs/debugging-mokapi-scripts
      description: Learn how to debug your JavaScript code with console logging and event tracing
    - title: CI/CD Integration
      href: /resources/blogs/end-to-end-testing-with-mocked-apis
      description: Run Mokapi in GitHub Actions for automated testing workflows
  
---
# Get started with REST API

In this tutorial, you will:

- Create an OpenAPI specification defining a REST API endpoint
- Write a JavaScript script to generate dynamic mock data
- Configure and run Mokapi in a Docker container
- Test the mocked API with HTTP requests

By the end of this tutorial, you'll have a working REST API mock that returns dynamically generated user data.

``` box=tree title="Project Structure"
📄 users.yaml
📄 users.js
📄 Dockerfile
```

## <i class="bi bi-1-circle-fill align-baseline"></i> Create OpenAPI file

Create a file named `users.yaml` that defines a `/users` endpoint:

```yaml
openapi: 3.0.0
info:
  title: Sample API
  version: 0.1.0
servers:
  - url: /api
paths:
  /users:
    get:
      summary: Returns a list of users.
      responses:
        '200': 
          description: A JSON array of user names
          content:
            application/json:
              schema: 
                type: array
                items: 
                  type: string
```

``` box=info title="What This Specification Does"
This OpenAPI specification defines a single endpoint <code>GET /api/users</code> that returns a JSON array 
of strings representing usernames. The specification serves as both documentation and a contract
 for the mock server.
```

## <i class="bi bi-2-circle-fill align-baseline"></i> Create Mokapi Script

Create a JavaScript file named `users.js` to generate dynamic mock data:

```javascript
import { on } from 'mokapi'
import { fake } from 'mokapi/faker'

export default function() {
    on('http', function(request, response) {
        if (request.url.path === '/api/users') {
            response.data = [
                fake({type: 'string', format: '{username}'}),
                fake({type: 'string', format: '{username}'}),
                fake({type: 'string', format: '{username}'})
            ]
        }
    })
}
```

### How This Script Works
- **Event Listener:** The `on('http', ...)` function registers a handler that fires for every HTTP request
- **Path Matching:** The script checks if the request path matches `/api/users`
- **Dynamic Data:** The `fake()` function generates random usernames using Mokapi's built-in Faker library
- **Response Data:** Setting `response.data` tells Mokapi to use this data as response body

``` box=info title="Why Use response.data?"
Using response.data instead of response.body ensures Mokapi validates your mock data against 
the OpenAPI specification. This catches schema mismatches early and guarantees your mocks 
match your API contract.
```

## <i class="bi bi-3-circle-fill align-baseline"></i> Create Dockerfile

Create a `Dockerfile` to configure and package Mokapi with your specification files:

```dockerfile
FROM mokapi/mokapi:latest

COPY ./users.yaml /demo/

COPY ./users.js /demo/

CMD ["--Providers.File.Directory=/demo"]
```

### Dockerfile Breakdown

- **Base Image:** Starts from the official Mokapi Docker image
- **Copy Files:** Copies both the OpenAPI specification and JavaScript script into the `/demo` directory
- **Configuration:** The `CMD` instruction tells Mokapi to load specifications from `/demo`

## <i class="bi bi-4-circle-fill align-baseline"></i> Start Mokapi

Build and run the Docker container in one command:

```
docker run -p 80:80 -p 8080:8080 --rm -it $(docker build -q .)
```

This command:
- **Builds** the Docker image from your Dockerfile
- **Exposes** port 80 for the mock REST API
- **Exposes** port 8080 for the Mokapi dashboard
- **Runs interactively** with automatic cleanup when stopped

``` box=result title="Mokapi is Running"
Open your browser and navigate to the Mokapi Dashboard at http://localhost:8080 to see your mocked REST API. The dashboard shows all configured endpoints, recent requests, and response details.
```

## <i class="bi bi-5-circle-fill align-baseline"></i> Make HTTP request

Test your mocked API by making an HTTP request to the `/api/users` endpoint:

```
curl --header "Accept: application/json" http://localhost/api/users
```

``` box=info title="View in Dashboard"
Every request and response is visible in the Mokapi dashboard at http://localhost:8080. You can 
inspect headers, request bodies, response times, and see which event handlers were triggered.
```

Each time you make a request, Mokapi generates three new random usernames. The response is validated
against your OpenAPI specification to ensure it matches the defined schema.

## Next Steps

Now that you have a working REST API mock, explore these advanced topics:

{{ card-grid key="cards" }}