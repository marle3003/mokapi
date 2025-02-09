---
title: Mokapi behind Reverse Proxy
description: How to use Mokapi behind a reverse proxy with CORS header
icon: bi-diagram-3-fill
---

# Mokapi behind Reverse Proxy

Example how to use Mokapi behind a reverse proxy by setting CORS header for each HTTP response

```javascript tab=cors.js  
import { on } from 'mokapi'

export default function () {
    on('http', function (request, response) {
        response.headers["access-control-allow-origin"] = "*"
        return true
    })
}
```