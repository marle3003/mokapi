---
title: Mock WebSocket APIs for Testing and Development
description: Mock WebSocket servers using AsyncAPI specification for seamless testing and development of real-time and event-driven applications
---

# Mocking WebSocket APIs with AsyncAPI

Mokapi turns an AsyncAPI specification into a working **WebSocket** mock server. Rather than standing up
a real backend, Mokapi focuses on the message contract: channels, payloads, and the bidirectional
communication between your client and the server it talks to. This removes the need for a running
backend during development and testing, while still ensuring your application strictly adheres to
its message contracts.

```yaml
asyncapi: 3.0.0
info:
  title: Chat Service
  version: 1.0.0

servers:
  mokapi:
    host: 'localhost:8765'
    protocol: ws

channels:
  chat:
    address: /chat/{roomId}
    messages:
      chatMessage:
        payload:
          type: object
          properties:
            userId: { type: string }
            text:   { type: string }
          required: [userId, text]
```

``` box=tip title=Recommendation
Ready to dive in? Head over to the WebSocket [Quick Start Guide](/docs/websocket/quick-start.md) and
run your first WebSocket mock in seconds.
```

``` box=warning title="Avoid port 8080"
Mokapi uses port `8080` by default for its dashboard, health checks, and MCP server. If your
WebSocket server runs on the same port, these services will conflict. Choose a different port for
your WebSocket server — `8765` is a common choice — or move Mokapi's built-in services to a
different port using the following flags:

| Service       | Flag                  | Default  |
|---------------|-----------------------|----------|
| Dashboard/API | `--api-port`          | `8080`   |
| Health check  | `--health-port`       | `8080`   |
| MCP server    | `--mcp-server-port`   | `8080`   |

See the [CLI flags reference](/docs/configuration/static/cli-flags) for details.
```

## Why Use Mokapi for WebSocket?

Testing real-time applications against a live backend is slow, hard to reproduce, and requires every
dependent service to be running. Mokapi provides a lightweight, stable **WebSocket mock server** built
specifically for local development and CI/CD pipelines.

**Zero infrastructure overhead**: No backend to spin up. Mokapi runs as a single binary or container
and is ready to accept WebSocket connections immediately.

**Contract-first validation**: Every message sent by a client is validated against your AsyncAPI schema
in real time, catching malformed payloads before they reach your application logic.

**Reproducible test suites**: Mokapi runs entirely in memory with no persistent state between runs, so
every test starts from a clean slate with no leftover connections or messages from previous runs.

## Supported Standards for WebSocket Mocking

Mokapi integrates with the existing WebSocket ecosystem and supports modern industry standards:

**AsyncAPI specifications**: Full support for both version 2.x and version 3.0.

**Schema formats**: Built-in validation for JSON Schema.

**WebSocket protocol**: Compatible with any standard WebSocket client — browsers, Node.js (`ws`),
Python (`websockets`), and others — over the standard WebSocket protocol (RFC 6455).

## Key Features of the Mokapi WebSocket Mock Server

### Automated Channel Provisioning

Mokapi reads your AsyncAPI definition and automatically provisions the channels it describes.
No manual setup required — channels are ready to connect to as soon as Mokapi starts.

### Channel Parameters

Real-world WebSocket APIs often use dynamic paths to separate concerns — for example, a chat
application where each room is a separate channel. Mokapi supports AsyncAPI channel parameters
out of the box:

```yaml
channels:
  chat:
    address: /chat/{roomId}
    parameters:
      roomId:
        description: The chat room identifier
        schema:
          type: string
```

Clients connecting to `/chat/room-1` and `/chat/room-2` are isolated from each other. A broadcast
from a client in `room-1` only reaches other clients in `room-1`.

### Bidirectional Messaging

WebSocket connections are full-duplex. Mokapi models both directions of the conversation as
separate operations in your AsyncAPI spec:

```yaml
operations:
  onChatMessage:
    action: receive          # client → server
    channel:
      $ref: '#/channels/chat'

  sendChatMessage:
    action: send             # server → client
    channel:
      $ref: '#/channels/chat'
```

### Scripted Mock Behavior

Use [Mokapi Scripts](/docs/javascript-api/overview.md) to implement dynamic server behavior — reply
to a specific client, broadcast to a room, or build stateful mock logic entirely in JavaScript:

```javascript
import { on } from 'mokapi'

export default function() {
  on('websocket', function(event) {
    if (event.type === 'connect') {
      // send a welcome message when a client connects
      event.client.send({ userId: 'server', text: 'welcome!' })
    }
    if (event.type === 'message') {
      // broadcast every message to all clients in the same channel
      event.broadcast({
        userId: event.message.userId,
        text: event.message.text
      })
    }
  })
}
```

### Error and Edge Case Simulation

Use scripts to simulate conditions that are difficult to reproduce with a real backend:

- Inject latency before responding
- Simulate a server that closes the connection unexpectedly
- Build stateful behavior, for example a chat room that rejects messages from banned users

## Message Flow

Unlike HTTP, a WebSocket connection is long-lived. Mokapi tracks each connection and the messages
that flow over it. The dashboard shows the full conversation per connection:

| Source    | Destination     | Value                            | Time     |
|-----------|-----------------|----------------------------------|----------|
| 127.0.0.1 | server          | `{"userId":"alice","text":"hi"}` | 10:42:01 |
| server    | 127.0.0.1       | `{"userId":"alice","text":"hi"}` | 10:42:01 |
| server    | 127.0.0.2       | `{"userId":"alice","text":"hi"}` | 10:42:01 |

## Next Steps

- [Quick Start Guide](/docs/websocket/quick-start.md): Run Mokapi and load your first AsyncAPI file.
- [Mokapi CLI](/docs/configuration/static/cli-usage.md): Command-line options and runtime configuration.
- [JavaScript API](/docs/javascript-api/overview.md): Write scripts to control mock behavior dynamically.