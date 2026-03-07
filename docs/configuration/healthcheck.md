---
title: Health Check Configuration
description: Configure Mokapi's health endpoint for uptime monitoring, orchestration systems, and Kubernetes probes.
subtitle: Configure Mokapi's health endpoint for uptime monitoring, orchestration systems, and Kubernetes probes.
---

# Health Check Configuration

## Overview

Mokapi provides a simple health endpoint to monitor service availability. This endpoint can be used for:

- **Uptime monitoring:** Check if Mokapi is responding to requests
- **Load balancer health checks:** Ensure traffic is only routed to healthy instances
- **Kubernetes probes:** Liveness and readiness probes for pod management
- **CI/CD validation:** Verify Mokapi started successfully before running tests

By default, the health check listens on `http://localhost:8080/health` but can be fully customized.

``` box=info title"Simple Health Check"
The health endpoint is minimal and does not perform checks of external dependencies. It simply indicates that Mokapi is up and listening. A 200 OK response means Mokapi is ready to accept requests.
```

## Configuration

Configure the health endpoint using YAML configuration or CLI flags. The health endpoint is controlled by the health section:

```yaml
health:
  enabled: true
  path: /health
  port: 8080
  log: false
```

### Configuration Fields

<div class="table-responsive-sm">
<table>
    <thead>
        <tr>
            <th class="col-2">Field</th>
            <th class="col-2">Type</th>
            <th class="col-2">Default</th>
            <th class="col">Description</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>enabled</td>
            <td>bool</td>
            <td>true</td>
            <td>Enables or disables the health check endpoint. When `false`, Mokapi does not expose any health endpoint.</td>
        </tr>
        <tr>
            <td>path</td>
            <td>string</td>
            <td>/health</td>
            <td>The HTTP path for the health endpoint. Must start with /.</td>
        </tr>
        <tr>
            <td>port</td>
            <td>int</td>
            <td>8080</td>
            <td>The port on which the health endpoint is exposed. If it matches the API/dashboard port, the health endpoint is served by the same HTTP server.</td>
        </tr>
        <tr>
            <td>log</td>
            <td>bool</td>
            <td>false</td>
            <td>If `true`, Mokapi logs all requests to the health endpoint using structured JSON logs. Useful for debugging but can generate high log volume.</td>
        </tr>
    </tbody>
</table>
</div>

### Example Configuration

```yaml
health:
  enabled: true
  path: /healthz        # Custom path
  port: 8081            # Dedicated port
  log: true             # Enable logging for debugging
```

### CLI Flags

The health endpoint can also be configured using command-line flags:

<div class="table-responsive-sm">
<table>
    <thead>
        <tr>
            <th class="col-2">Flag</th>
            <th class="col-4">Description</th>
            <th class="col-4">Example</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>--health-enabled</td>
            <td>Enable or disable the health endpoint</td>
            <td><code>mokapi --health-enabled=true</code></td>
        </tr>
        <tr>
            <td>--health-path</td>
            <td>Path for the health endpoint</td>
            <td><code>mokapi --health-path=/healthz</code></td>
        </tr>
        <tr>
            <td>--health-port</td>
            <td>Port for the health endpoint</td>
            <td><code>mokapi --health-port=8081</code></td>
        </tr>
        <tr>
            <td>--health-log</td>
            <td>Log all health requests</td>
            <td><code>mokapi --health-log</code></td>
        </tr>
    </tbody>
</table>
</div>

### Example Command

```
mokapi /api/spec.yaml \
  --health-path=/healthz \
  --health-port=8081 \
  --health-log
```

## Health Endpoint Response

When Mokapi is healthy and responding, the endpoint returns:

```http 
GET /health HTTP/1.1
Host: localhost:8080

HTTP/1.1 200 OK
Content-Type: application/json

{"status":"healthy"}
```

- **Status Code:** 200 OK indicates the service is healthy and ready to accept requests
- **Content Type:** application/json
- **Body:** JSON object with status: "healthy"

``` box=warning title="Limited Health Check"
The health endpoint only validates that Mokapi's HTTP server is running and can process requests.
```

## Kubernetes Integration

Use Mokapi's health endpoint with Kubernetes liveness and readiness probes to ensure proper orchestration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mokapi
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mokapi
  template:
    metadata:
      labels:
        app: mokapi
    spec:
      containers:
        - name: mokapi
          image: mokapi:latest
          ports:
            - containerPort: 8080
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 2
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 5
            timeoutSeconds: 2
            successThreshold: 1
```

### Understanding the Probes

**Liveness Probe:** Kubernetes uses this to determine if the container is alive. If the liveness probe fails, Kubernetes restarts the container.

**Readiness Probe:** Kubernetes uses this to determine if the container is ready to accept traffic. If the readiness probe fails, Kubernetes removes the pod from service endpoints.

``` box=info title="Single Endpoint for Both Probes"
Mokapi provides only one health endpoint (/health). Both livenessProbe and readinessProbe can point to
the same endpoint. The probes differ in their configuration (timing, thresholds), not the endpoint they check.
```

### Kubernetes Configuration Notes

- **Port conflicts:** Ensure the health port does not conflict with mocked APIs or dashboard ports in the same container
- **Timing:** Adjust `initialDelaySeconds` based on Mokapi's startup time with your configuration
- **Logging:** If `log`: true is enabled, health requests will appear in logs
- **Failure thresholds:** Set appropriate `failureThreshold` values to avoid premature restarts

## Best Practices

- **Use a dedicated port (optional):**   
  If your mocked APIs or dashboard run on port 8080, consider setting `health.port` to a different port (e.g., 8081) to avoid conflicts and simplify network routing.
- **Enable logging only when debugging:**  
  High-frequency probes can generate significant log volume. Use `log: true` only when troubleshooting probe failures or validating probe configuration.