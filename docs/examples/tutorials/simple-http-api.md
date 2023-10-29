---
title: Get started with Mokapi & REST API
description: In this example, you will learn how to mock an OpenAPI specification with Mokapi, how to configure Mokapi, and how to run it in a Docker container.
---
# Get started with Mokapi & REST API

In this example, you will learn how to mock an OpenAPI specification with Mokapi, how to configure Mokapi, 
and how to run it in a Docker container. In further examples we will see how the test quality can be improved with some 
edge cases such as increased response time or response errors like internal server errors.

## Create OpenAPI file

First, we create an OpenAPI specification file `users.yaml` that defines a /users endpoint

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
This specification defines an endpoint /api/users that returns an array of strings containing usernames.

## Create Mokapi Scripts
Next, create a javascript file `users.js` which sets the content of the response

```javascript
import {on} from 'mokapi'
import {fake} from 'mokapi/faker'

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
The script sets a string array containing three random usernames to the response data if the requested endpoint is `/api/users`.

## Create a Dockerfile
Next create a `Dockerfile` to configure Mokapi
```dockerfile
FROM mokapi/mokapi:latest

COPY ./users.yaml /demo/

COPY ./users.js /demo/

CMD ["--Providers.File.Directory=/demo"]
```

## Start Mokapi

```
docker run -p 80:80 -p 8080:8080 --rm -it $(docker build -q .)
```
You can open a browser and go to Mokapi's Dashboard (http://localhost:8080) to see the Mokapi's HTTP REST API.

## Make HTTP request
You can now make HTTP requests to our Sample API and Mokapi creates responses with randomly generated data.

```
curl --header "Accept: application/json" http://localhost/api/users
```
The request and the response are visible in the dashboard.