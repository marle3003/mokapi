---
title: KafkaRouter
description: Handle incoming Kafka messages and produce messages to Kafka topics using the app API.
---
# KafkaRouter

`KafkaRouter` registers handlers for incoming Kafka messages and produces messages to topics.
Get one via `app.kafka()` or `app.api('My API').kafka()`.

```javascript
import { app } from 'mokapi'

export default function() {
    const orders = app.api('Order Events').kafka()

    orders.message((message) => {
        console.log(`${message.topic}: ${message.value}`)
    })

    orders.produce('order-placed', { data: { id: 1, item: 'Book' } })
}
```

Use [`topic( name )`](#topic) to work with a single topic without repeating its name.

## Methods

| Method                                                | Returns                                                         | Description                                          |
|-------------------------------------------------------|-----------------------------------------------------------------|------------------------------------------------------|
| [`topic( name )`](#topic)                             | [KafkaTopic](/docs/javascript-api/mokapi/app/kafka-topic.md)    | Returns a router for a single topic                  |
| [`message( handler, [opts] )`](#message)              | `KafkaRouter`                                                   | Registers a handler for messages on all topics       |
| [`produce( topic, message, [opts] )`](#produce)       | `KafkaRouter`                                                   | Produces one or more messages                        |
| [`produceAsync( topic, message, [opts] )`](#produce)  | `Promise<KafkaRouter>`                                          | Produces one or more messages asynchronously         |
| [`api( title )`](#api)                                | `KafkaRouter`                                                   | Returns a router scoped to the given API             |

## topic

Returns a [KafkaTopic](/docs/javascript-api/mokapi/app/kafka-topic.md) for the given topic name,
as defined in the AsyncAPI specification.

```javascript
app.api('Order Events').kafka()
    .topic('order-placed')
        .message((message) => { console.log(message.value) })
        .produce({ data: { id: 1 } })
```

## message

Registers a handler for incoming messages on **all** topics of this router. The handler can
read and modify the message before Mokapi processes it.

| Parameter       | Type     | Description                                                                                           |
|-----------------|----------|-------------------------------------------------------------------------------------------------------|
| handler         | function | A [KafkaEventHandler](/docs/javascript-api/mokapi/eventhandler/kafkaeventhandler.md) that receives a [KafkaEventMessage](/docs/javascript-api/mokapi/eventhandler/kafkaeventmessage.md) |
| opts (optional) | object   | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) such as `priority`, `track`, `tags` |

```javascript
import { app } from 'mokapi'

export default function() {
    // Add a correlation ID to every incoming message
    app.kafka().message((message) => {
        message.headers = { ...message.headers, correlationId: 'abc-123' }
    })
}
```

## produce

Produces one or more messages to the given topic. `produce()` runs synchronously,
`produceAsync()` returns a `Promise` you can `await`.

`produceAsync()` reads the messages and options when you call it, so changing a message afterwards doesn't
affect what is produced. Invalid arguments throw an error at the call. Client errors, such as a broker error,
reject the promise, and so does an error thrown in the `result` callback.

| Parameter       | Type            | Description                                                                                   |
|-----------------|-----------------|-----------------------------------------------------------------------------------------------|
| topic           | string          | Kafka topic name                                                                              |
| message         | object \| array | One [KafkaProduceMessage](/docs/javascript-api/mokapi/app/kafka-produce.md#kafkaproducemessage) or an array of them |
| opts (optional) | object          | [KafkaProduceArgs](/docs/javascript-api/mokapi/app/kafka-produce.md#kafkaproduceargs) with a result callback and retry settings |

If a message sets neither `data` nor `value`, Mokapi generates a random valid value based on
the AsyncAPI schema.

```javascript
import { app } from 'mokapi'

export default function() {
    app.kafka().produce('order-placed', [
        { key: 'order-1', data: { id: 1 } },
        { data: { id: 2 } },
        {} // random value generated from the schema
    ], {
        result: (r) => console.log(`offset=${r.messages[0].offset}`)
    })
}
```

```javascript
import { app } from 'mokapi'

export default async function() {
    await app.kafka().produceAsync('order-placed', { data: { id: 1 } })
}
```

## api

Returns a new `KafkaRouter` scoped to the given API title. This is equivalent to
[`app.api( title ).kafka()`](/docs/javascript-api/mokapi/app/api.md).

```javascript
app.kafka().api('Order Events').topic('order-placed').message((message) => {
    console.log(message.value)
})
```
