---
title: "WebSocket API Mocking: Build a Chat Room Mock with Mokapi"
description: Learn how to mock WebSocket servers using AsyncAPI specifications. Create a lightweight WebSocket mock server for real-time testing and local development.
---

# Mocking a Chat Room with AsyncAPI WebSocket

## Overview

This guide walks through a complete real-time chat workflow using Mokapi as a mock WebSocket server.
By the end, you'll have:

1. Defined a WebSocket server and a `/chat/{roomId}` channel using AsyncAPI.
2. Started Mokapi to simulate the server and validate messages against the schema.
3. Connected two clients to the same chat room.
4. Used a Mokapi Script to broadcast messages to all clients in the room.
5. Seen what happens when a message doesn't match the schema.

## Define the AsyncAPI Specification

Create an AsyncAPI file (`asyncapi.yaml`) describing the WebSocket server and the chat channel.

```yaml
asyncapi: 3.0.0
info:
  title: Chat Service
  version: 1.0.0
  description: This AsyncAPI document defines the WebSocket server for a simple chat application.

servers:
  mokapi:
    host: 'localhost:8765'
    protocol: ws
    description: Mock WebSocket server provided by Mokapi.

channels:
  chat:
    address: /chat/{roomId}
    description: A chat room. Each roomId is a separate, isolated room.
    parameters:
      roomId:
        description: Unique identifier for the chat room.
        schema:
          type: string
    messages:
      ChatMessage:
        $ref: '#/components/messages/ChatMessage'

operations:
  sendChatMessage:
    action: send
    summary: Client sends a chat message to the room.
    channel:
      $ref: '#/channels/chat'
    messages:
      - $ref: '#/channels/chat/messages/ChatMessage'

  receiveChatMessage:
    action: receive
    summary: Server broadcasts a chat message to all clients in the room.
    channel:
      $ref: '#/channels/chat'
    messages:
      - $ref: '#/channels/chat/messages/ChatMessage'

components:
  messages:
    ChatMessage:
      name: ChatMessage
      title: A chat message
      contentType: application/json
      payload:
        type: object
        properties:
          userId:
            type: string
            description: The ID of the user sending the message.
          text:
            type: string
            description: The message content.
        required:
          - userId
          - text
```

The `required` field ensures every message includes a `userId` and `text`. Mokapi enforces this
automatically once the server is running.

The `{roomId}` parameter in the channel address means `/chat/room-1` and `/chat/room-2` are
separate rooms — clients in different rooms are isolated from each other.

## Start Mokapi

Start Mokapi with the AsyncAPI file to simulate the WebSocket server:

```bash
mokapi asyncapi.yaml
```

Mokapi's log output will confirm a WebSocket server is running at `localhost:8765`, ready to accept
connections.

``` box=warning title="Port Conflict with Mokapi's Built-in Services"
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

## Add a Broadcast Script

By default, Mokapi validates incoming messages but doesn't send anything back. To simulate a chat
server that broadcasts each message to everyone in the room, add a Mokapi Script.

Create a file called `chat.js`:

```javascript
import { on } from 'mokapi'

export default function() {
  on('websocket', function(event) {
    // Broadcast the message to all clients in the same room
    event.broadcast({
      userId: event.message.userId,
      text: event.message.text
    })
  })
}
```

Start Mokapi with both files:

```bash
mokapi asyncapi.yaml chat.js
```

Now every message a client sends is broadcast to all other clients connected to the same room.

## Connect Two Clients

Use any standard WebSocket client to connect two clients to the same room. Here's an example using
the `ws` library for Node.js:

```javascript
import WebSocket from 'ws'

// Connect Alice and Bob to the same room
const alice = new WebSocket('ws://localhost:8765/chat/room-1')
const bob   = new WebSocket('ws://localhost:8765/chat/room-2') // different room — won't receive Alice's messages
const carol = new WebSocket('ws://localhost:8765/chat/room-1') // same room as Alice

carol.on('message', (data) => {
  console.log('Carol received:', data.toString())
  // prints: {"userId":"alice","text":"Hello everyone!"}
})

alice.on('open', () => {
  alice.send(JSON.stringify({
    userId: 'alice',
    text: 'Hello everyone!'
  }))
})
```

Mokapi receives Alice's message, validates it against the `ChatMessage` schema, and broadcasts it to
all clients in `room-1` — Alice and Carol. Bob, who is in `room-2`, receives nothing.

## Monitor Messages in the Dashboard

Instead of writing a subscriber just to check that messages are flowing, you can use Mokapi's web
dashboard to inspect WebSocket traffic directly.

1. Open the dashboard, by default at http://localhost:8080.
2. Go to the WebSocket tab.
3. Select the `Chat Service` to see all active connections and messages, including their direction,
   payload, and timestamp.

The dashboard shows the full conversation per connection:

| Source    | Destination | Value                                       | Time     |
|-----------|-------------|----------------------------------------------|----------|
| 127.0.0.1 | server      | `{"userId":"alice","text":"Hello everyone!"}` | 10:42:01 |
| server    | 127.0.0.1   | `{"userId":"alice","text":"Hello everyone!"}` | 10:42:01 |
| server    | 127.0.0.3   | `{"userId":"alice","text":"Hello everyone!"}` | 10:42:01 |

## Schema Validation

Every message sent by a client is validated against the `ChatMessage` schema defined in
`asyncapi.yaml`. If a message doesn't conform to the schema:

- Mokapi closes the connection with a WebSocket status code and a reason describing the violation.
- Mokapi logs a validation error.
- The invalid message does not appear in the dashboard.

## Test an Invalid Message

Send a message that's missing the required `userId` field:

```javascript
alice.send(JSON.stringify({ text: 'oops, forgot my userId' }))
```

Mokapi rejects this message because `userId` is required by the schema. The connection is closed
with status `1003 Unsupported Data` and a reason like:

```
Validation error: #/required: required properties are missing: userId
```

## Reply to a Specific Client

To send a response only to the client that sent the message rather than broadcasting to everyone,
use `event.reply` in your script:

```javascript
import { on } from 'mokapi'

export default function() {
  on('websocket', function(event) {
    if (event.message.text === 'ping') {
      event.reply({ userId: 'server', text: 'pong' })
      return
    }
    event.broadcast({
      userId: event.message.userId,
      text: event.message.text
    })
  })
}
```

## Next Steps

- [WebSocket Overview](/docs/websocket/overview.md): Learn more about how Mokapi simulates WebSocket servers.
- [JavaScript API](/docs/javascript-api/overview.md): Add dynamic behavior, simulate errors, or build stateful mock logic with Mokapi Scripts.