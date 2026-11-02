---
title: WebsocketRouter
description: Handle WebSocket connect, message and close events and send messages to channels using the app API.
---
# WebsocketRouter

`WebsocketRouter` registers handlers for WebSocket lifecycle events and sends messages to
connected clients. Get one via `app.websocket()` or `app.api('My API').websocket()`.

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Chat API').websocket()
        .channel('/chat/{roomId}')
            .connect((event) => {
                event.client.send({ text: `Welcome to ${event.channel.params['roomId']}!` })
            })
            .message((event) => {
                event.broadcast(event.message)
            })
            .close((event) => {
                event.broadcast({ text: `${event.client.query['userId']} has left` })
            })
}
```

Three events are available:

- `connect`: a client opens a connection
- `message`: a client sends a message
- `close`: the connection is closed by the client or the server

## Methods

| Method                                  | Returns                                         | Description                                          |
|-----------------------------------------|-------------------------------------------------|------------------------------------------------------|
| [`channel( name )`](#channel)           | [WebsocketChannelRouter](#websocketchannelrouter) | Returns a router for a single channel              |
| `connect( handler, [opts] )`            | `WebsocketRouter`                               | Registers a handler for `connect` on all channels    |
| `message( handler, [opts] )`            | `WebsocketRouter`                               | Registers a handler for `message` on all channels    |
| `close( handler, [opts] )`              | `WebsocketRouter`                               | Registers a handler for `close` on all channels      |
| [`send( channel, message, [args] )`](#send) | `void`                                      | Sends a message to all clients of a channel          |
| `api( title )`                          | `WebsocketRouter`                               | Returns a router scoped to the given API             |

### Parameters

| Parameter       | Type     | Description                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------|
| handler         | function | Function that receives the [event object](#event-objects)                                             |
| opts (optional) | object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) such as `priority`, `track`, `tags` |

## channel

Returns a [WebsocketChannelRouter](#websocketchannelrouter) for the given channel. The name must
match the channel address in the AsyncAPI specification, for example `/chat/{roomId}`.

## send

Sends a message to all clients currently connected to a channel. Use it for server-initiated
messages outside an event handler, for example from a scheduled job.

| Parameter       | Type   | Description                                                                     |
|-----------------|--------|---------------------------------------------------------------------------------|
| channel         | string | Channel name as defined in the AsyncAPI specification                           |
| message         | any    | Payload, encoded according to the channel's `contentType`                       |
| args (optional) | object | [WebsocketSendArgs](#websocketsendargs)                                         |

``` box=warning
Without an API scope, Mokapi searches all APIs for the channel. `send()` throws an error if the
channel exists in more than one API or in none. Use `app.api('My API').websocket().send()` to
avoid this.
```

```javascript
import { app, every } from 'mokapi'

export default function() {
    every('10s', () => {
        app.api('Chat API').websocket().send('/updates', { text: 'ping' })
    })
}
```

## WebsocketChannelRouter

Groups handlers for a single channel. The methods behave like those of `WebsocketRouter`, but
only apply to this channel, and `send()` takes no channel name.

| Method                          | Returns                  | Description                                         |
|---------------------------------|--------------------------|-----------------------------------------------------|
| `connect( handler, [opts] )`    | `WebsocketChannelRouter` | Registers a handler for `connect` on this channel   |
| `message( handler, [opts] )`    | `WebsocketChannelRouter` | Registers a handler for `message` on this channel   |
| `close( handler, [opts] )`      | `WebsocketChannelRouter` | Registers a handler for `close` on this channel     |
| `send( message, [args] )`       | `void`                   | Sends a message to all clients of this channel      |

## Event Objects

All events share these properties:

| Name               | Type     | Description                                                       |
|--------------------|----------|-------------------------------------------------------------------|
| type               | string   | `connect`, `message` or `close`                                   |
| api                | string   | Name of the WebSocket API (AsyncAPI `info.title`)                 |
| channel            | object   | The channel: `key`, `name`, `params` and connected `clients`      |
| client             | object   | The client: `remoteAddr`, `query`, `headers` and `send(message)`  |
| broadcast(message) | function | Sends a message to all clients connected to the channel           |

Additional properties per event:

| Event     | Name           | Description                                         |
|-----------|----------------|-----------------------------------------------------|
| `connect` | reply(message) | Sends a message to the connecting client            |
| `message` | message        | The message sent by the client                      |
| `message` | reply(message) | Sends a message back to the sender                  |
| `close`   | reason         | Reason for closing the connection                   |
| `close`   | closedBy       | `client` or `server`                                |

## WebsocketSendArgs

| Name              | Type     | Description                                                                                          |
|-------------------|----------|------------------------------------------------------------------------------------------------------|
| result (optional) | function | Called after the message has been sent. The result contains `api`, `channel` and `message`.          |
| retry (optional)  | object   | [Retry](/docs/javascript-api/mokapi/app/retry.md) settings, used when the channel is not yet available |
