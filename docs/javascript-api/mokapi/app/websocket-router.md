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

| Method                                           | Returns                                           | Description                                                |
|--------------------------------------------------|---------------------------------------------------|------------------------------------------------------------|
| [`channel( name )`](#channel)                    | [WebsocketChannelRouter](#websocketchannelrouter) | Returns a router for a single channel                      |
| `connect( handler, [opts] )`                     | `WebsocketRouter`                                 | Registers a handler for `connect` on all channels          |
| `message( handler, [opts] )`                     | `WebsocketRouter`                                 | Registers a handler for `message` on all channels          |
| `close( handler, [opts] )`                       | `WebsocketRouter`                                 | Registers a handler for `close` on all channels            |
| [`send( channel, message, [args] )`](#send)      | `WebsocketRouter`                                 | Sends a message to all clients of a channel                |
| [`sendAsync( channel, message, [args] )`](#send) | `Promise<WebsocketRouter>`                        | Sends a message to all clients of a channel asynchronously |
| `api( title )`                                   | `WebsocketRouter`                                 | Returns a router scoped to the given API                   |

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
messages outside an event handler, for example from a scheduled job. `send()` runs
synchronously, `sendAsync()` returns a `Promise` you can `await`.

`sendAsync()` reads the message and arguments when you call it, so changing the message afterwards doesn't
affect what is sent. Invalid arguments, such as a `retry` of the wrong type, throw an error at the call.
Client errors, such as an unknown channel, reject the promise, and so does an error thrown in the `result` callback.

| Parameter       | Type   | Description                                                                     |
|-----------------|--------|---------------------------------------------------------------------------------|
| channel         | string | Channel name as defined in the AsyncAPI specification                           |
| message (optional) | any | Payload, encoded according to the channel's `contentType`. If omitted, Mokapi generates a payload from the channel's message |
| args (optional) | object | [WebsocketSendArgs](#websocketsendargs)                                         |

``` box=warning
Without an API scope, Mokapi searches all APIs for the channel. `send()` throws an error if the
channel exists in more than one API or in none, and `sendAsync()` rejects. Use
`app.api('My API').websocket().send()` to avoid this.
```

```javascript
import { app, every } from 'mokapi'

export default function() {
    every('10s', () => {
        app.api('Chat API').websocket().send('/updates', { text: 'ping' })
    })
}
```

Using `sendAsync()`:

```javascript
import { app } from 'mokapi'

export default async function() {
    await app.api('Chat API').websocket().sendAsync('/updates', { text: 'ping' }, {
        result: (r) => console.log(`Sent to ${r.channel}`)
    })
}
```

## WebsocketChannelRouter

Groups handlers for a single channel. The methods behave like those of `WebsocketRouter`, but
only apply to this channel, and `send()` and `sendAsync()` take no channel name.

| Method                         | Returns                           | Description                                                   |
|--------------------------------|-----------------------------------|---------------------------------------------------------------|
| `connect( handler, [opts] )`   | `WebsocketChannelRouter`          | Registers a handler for `connect` on this channel             |
| `message( handler, [opts] )`   | `WebsocketChannelRouter`          | Registers a handler for `message` on this channel             |
| `close( handler, [opts] )`     | `WebsocketChannelRouter`          | Registers a handler for `close` on this channel               |
| `send( message, [args] )`      | `WebsocketChannelRouter`          | Sends a message to all clients of this channel                |
| `sendAsync( message, [args] )` | `Promise<WebsocketChannelRouter>` | Sends a message to all clients of this channel asynchronously |

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
| messageId (optional) | string | Key of the message under `channels.<channel>.messages` in the AsyncAPI specification, see [Selecting a message](#selecting-a-message) |
| result (optional) | function | Called after the message has been sent, before the `sendAsync()` promise resolves. The result contains `api`, `channel` and `message`.          |
| retry (optional)  | object   | [Retry](/docs/javascript-api/mokapi/app/retry.md) settings, used when the channel is not yet available |

## Selecting a message

A channel can define several messages. Set `messageId` in the send arguments to choose one. The
value is the key of the message under `channels.<channel>.messages` in your AsyncAPI
specification. For AsyncAPI 2, it is the key Mokapi derives: the `operationId` of the operation,
otherwise the last segment of the message `$ref`. It is the same value the dashboard shows as
`messageId` in the message log.

Without a message, Mokapi generates the payload from the selected message. If the selected
message has no payload, Mokapi sends an empty text frame. With a message, Mokapi validates it
against the selected message and throws an error if it is invalid. An unknown `messageId` throws
an error. `sendAsync()` rejects the promise.

```javascript
import { app } from 'mokapi'

export default function() {
    const chat = app.api('Chat API').websocket()

    // payload generated from the message 'serverMessage'
    chat.send('/chat', undefined, { messageId: 'serverMessage' })

    // validated against the message 'serverMessage'
    chat.send('/chat', { text: 'hello' }, { messageId: 'serverMessage' })
}
```
