---
title: MqttRouter
description: Handle incoming MQTT messages and publish messages to MQTT topics using the app API.
---
# MqttRouter

`MqttRouter` registers handlers for incoming MQTT messages and publishes messages to topics.
Get one via `app.mqtt()` or `app.api('My API').mqtt()`.

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Smart Home').mqtt()
        .topic('sensors/temperature')
            .message((message) => {
                console.log(`Temperature: ${message.value}`)
            })
            .publish({ value: '21.5', retain: true })
}
```

## Methods

| Method                                              | Returns                  | Description                                       |
|-----------------------------------------------------|--------------------------|---------------------------------------------------|
| [`topic( name )`](#topic)                           | [MqttTopicRouter](#mqtttopicrouter) | Returns a router for a single topic    |
| [`message( handler, [opts] )`](#message)            | `MqttRouter`             | Registers a handler for messages on all topics    |
| [`publish( topic, message, [opts] )`](#publish)     | `MqttRouter`             | Publishes a message                               |
| [`publishAsync( topic, message, [opts] )`](#publish)| `Promise<MqttRouter>`    | Publishes a message asynchronously                |

To scope the router to one API, use [`app.api( title ).mqtt()`](/docs/javascript-api/mokapi/app/api.md).

## topic

Returns an [MqttTopicRouter](#mqtttopicrouter) for the given topic name, as defined in the
AsyncAPI specification. Publishing requires an exact topic name; wildcards (`+`, `#`) are not
allowed.

## message

Registers a handler for incoming messages on **all** topics of this router. The handler can
read and modify the message before Mokapi processes it.

| Parameter       | Type     | Description                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------|
| handler         | function | Function that receives an [MqttEventMessage](#mqtteventmessage)                                      |
| opts (optional) | object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) such as `priority`, `track`, `tags` |

```javascript
import { app } from 'mokapi'

export default function() {
    app.mqtt().message((message) => {
        console.log(`[${message.topic}] ${message.value}`)
    })
}
```

## publish

Publishes a message to the given topic. `publish()` runs synchronously, `publishAsync()`
returns a `Promise` you can `await`.

`publishAsync()` reads the message and options when you call it, so changing the message afterwards doesn't
affect what is published. Invalid arguments throw an error at the call. Client errors reject the promise,
and so does an error thrown in the `result` callback.

| Parameter       | Type   | Description                                              |
|-----------------|--------|----------------------------------------------------------|
| topic           | string | MQTT topic name                                          |
| message         | object | [MqttPublishMessage](#mqttpublishmessage)                |
| opts (optional) | object | [MqttPublishArgs](#mqttpublishargs)                      |

If the message sets neither `data` nor `value`, Mokapi generates a random valid value based on
the AsyncAPI schema.

```javascript
import { app } from 'mokapi'

export default function() {
    app.mqtt().publish('sensors/temperature', { value: '21.5', retain: true }, {
        result: (r) => console.log(`Published to ${r.topic}: ${r.value}`)
    })
}
```

## MqttTopicRouter

Groups handlers and publishers for a single topic. The methods behave like those of
`MqttRouter`, but `publish()` and `publishAsync()` take no topic name.

| Method                              | Returns                    | Description                                    |
|-------------------------------------|----------------------------|------------------------------------------------|
| `message( handler, [opts] )`        | `MqttTopicRouter`          | Registers a handler for messages on this topic |
| `publish( message, [opts] )`        | `MqttTopicRouter`          | Publishes a message to this topic              |
| `publishAsync( message, [opts] )`   | `Promise<MqttTopicRouter>` | Publishes a message asynchronously             |

## MqttEventMessage

| Name   | Type    | Description                                     |
|--------|---------|-------------------------------------------------|
| api    | string  | Name of the MQTT API (AsyncAPI `info.title`)    |
| topic  | string  | Topic the message was published to              |
| value  | string  | Message payload. Can be modified by the handler |
| retain | boolean | Whether the message is retained by the broker   |

## MqttPublishMessage

| Name              | Type    | Description                                                                                     |
|-------------------|---------|-------------------------------------------------------------------------------------------------|
| data (optional)   | any     | Structured data, encoded according to the topic's content type and validated against the schema |
| value (optional)  | any     | Raw value written as-is, without validation. Objects are JSON-serialized.                       |
| retain (optional) | boolean | If `true`, the broker keeps the message and delivers it to new subscribers. Default: `false`    |

## MqttPublishArgs

| Name              | Type     | Description                                                                                         |
|-------------------|----------|-----------------------------------------------------------------------------------------------------|
| result (optional) | function | Called with the result after the message has been published. The result contains `api`, `topic` and `value`. |
| retry (optional)  | object   | [Retry](/docs/javascript-api/mokapi/app/retry.md) settings, used when the topic is not yet available |
