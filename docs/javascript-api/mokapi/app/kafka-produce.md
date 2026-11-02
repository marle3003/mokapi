---
title: Kafka Produce Types
description: Message, options and result objects used by produce() and produceAsync() of KafkaRouter and KafkaTopic.
---
# Kafka Produce Types

These objects are used by `produce()` and `produceAsync()` of
[KafkaRouter](/docs/javascript-api/mokapi/app/kafka-router.md) and
[KafkaTopic](/docs/javascript-api/mokapi/app/kafka-topic.md).

## KafkaProduceMessage

A single message to produce.

| Name                 | Type   | Description                                                                                              |
|----------------------|--------|----------------------------------------------------------------------------------------------------------|
| key (optional)       | any    | Message key. If omitted, Mokapi generates a random key based on the AsyncAPI schema.                     |
| data (optional)      | any    | Structured data. Mokapi encodes it according to the topic's content type and validates it against the schema. |
| value (optional)     | any    | Raw value written as-is, without validation. Objects are JSON-serialized; `ArrayBuffer` is written unchanged. |
| headers (optional)   | object | Kafka headers. They are not validated against the specification.                                        |
| partition (optional) | number | Zero-based partition index. If omitted, Mokapi selects a partition.                                      |

If neither `data` nor `value` is set, Mokapi generates a random valid value based on the
AsyncAPI schema.

``` box=tip
Prefer `data` when the topic has a schema. It guarantees that the message matches your
AsyncAPI contract. Use `value` when you need full control over the bytes, for example to test
how a consumer handles invalid messages.
```

## KafkaProduceArgs

Optional settings for a produce call.

| Name               | Type     | Description                                                                                         |
|--------------------|----------|-----------------------------------------------------------------------------------------------------|
| result (optional)  | function | Called with a [KafkaProduceResult](#kafkaproduceresult) after the messages have been written        |
| retry (optional)   | object   | [Retry](/docs/javascript-api/mokapi/app/retry.md) settings, used when the topic is not yet available |

## KafkaProduceResult

| Name     | Type   | Description                                                                         |
|----------|--------|-------------------------------------------------------------------------------------|
| api      | string | Name of the Kafka API (AsyncAPI `info.title`)                                       |
| topic    | string | Name of the topic                                                                   |
| messages | array  | One [KafkaMessageResult](#kafkamessageresult) per produced message, in input order  |

## KafkaMessageResult

| Name      | Type   | Description                                         |
|-----------|--------|-----------------------------------------------------|
| partition | number | Partition the message was written to                |
| offset    | number | Offset of the message within its partition          |
| key       | string | Key of the written message                          |
| value     | string | Serialized payload of the written message           |
| headers   | object | Headers of the written message                      |

## Example

```javascript
import { app } from 'mokapi'

export default function() {
    app.api('Order Events').kafka().produce('order-placed', {
        key: 'order-1',
        data: { id: 1, item: 'Book' },
        headers: { correlationId: 'abc-123' }
    }, {
        retry: { retries: 10, initialRetryTime: '500ms' },
        result: (r) => {
            for (const m of r.messages) {
                console.log(`partition=${m.partition} offset=${m.offset}`)
            }
        }
    })
}
```
