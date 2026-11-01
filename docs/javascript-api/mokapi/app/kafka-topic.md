---
title: KafkaTopic
description: Handle and produce messages for a single Kafka topic using method chaining.
---
# KafkaTopic

`KafkaTopic` groups message handlers and producers for a single Kafka topic. Get one via
[`router.topic( name )`](/docs/javascript-api/mokapi/app/kafka-router.md#topic).

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Order Events').kafka()
        .topic('order-placed')
            .message((message) => {
                message.headers = { ...message.headers, processed: 'true' }
            })
            .produce({ data: { id: 1 } })
}
```

## Methods

| Method                              | Returns               | Description                                        |
|-------------------------------------|-----------------------|----------------------------------------------------|
| `message( handler, [opts] )`        | `KafkaTopic`          | Registers a handler for messages on this topic     |
| `produce( message, [opts] )`        | `KafkaTopic`          | Produces one or more messages to this topic        |
| `produceAsync( message, [opts] )`   | `Promise<KafkaTopic>` | Produces one or more messages asynchronously       |

The methods behave like their counterparts on
[KafkaRouter](/docs/javascript-api/mokapi/app/kafka-router.md), but `produce()` and
`produceAsync()` take no topic name.

### Parameters

| Parameter       | Type            | Description                                                                                   |
|-----------------|-----------------|-----------------------------------------------------------------------------------------------|
| handler         | function        | A [KafkaEventHandler](/docs/javascript-api/mokapi/eventhandler/kafkaeventhandler.md)          |
| message         | object \| array | One [KafkaProduceMessage](/docs/javascript-api/mokapi/app/kafka-produce.md#kafkaproducemessage) or an array of them |
| opts (optional) | object          | [EventArgs](/docs/javascript-api/mokapi/eventhandler/eventargs.md) for `message()`, [KafkaProduceArgs](/docs/javascript-api/mokapi/app/kafka-produce.md#kafkaproduceargs) for `produce()` |

## Example: Produce Periodically

```javascript
import { app, every } from 'mokapi'

export default function() {
    const topic = app.api('Order Events').kafka().topic('order-placed')

    every('10s', () => {
        topic.produce({}, {
            result: (r) => console.log(`produced offset ${r.messages[0].offset}`)
        })
    })
}
```
