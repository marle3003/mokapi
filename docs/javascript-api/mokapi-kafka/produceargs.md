---
title: ProduceArgs
description: ProduceArgs is an object used by produce functions to write Kafka messages into a topic.
---
# ProduceArgs

ProduceArgs is an object used by [produce](/docs/javascript-api/mokapi-kafka/produce.md) function

| Name                              | Type   | Description                                                                                                                                                       |
|-----------------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| cluster (optional)                | string | Kafka cluster name. Used when topic name is not unique.                                                                                                           |
| topic (optional)                  | string | Kafka topic name. If not specified, message will be written to a random topic.                                                                                    |
| messages                          | array  | An list of [Message](/docs/javascript-api/mokapi-kafka/message.md) contains Kafka messages to produce into given topic                                            |
| partition (optional, deprecated ) | number | Kafka partition index. If not specified, the message will be written to any partition                                                                             |
| key (optional, deprecated)        | any    | Kafka message key. If not specified, a random key will be generated based on the topic configuration.                                                             |
| value (optional, deprecated)      | any    | Kafka message value. If not specified, a random value will be generated based on the topic configuration. Object will be encoded based on the topic configuration |
| headers (optional, deprecated)    | object | Kafka message headers.                                                                                                                                            |
| retry                             | object | [ProduceRetry](/docs/javascript-api/mokapi-kafka/produceretry.md) object is used if script is executed before Kafka topic is set up.                              |

## Example

```javascript tab=kafka.js
import { produce } from 'mokapi/kafka'

export default function() {
    produce({
        topic: 'foo',
        messages: [
            {
                key: 'foo-1',
                data: { foo: 'bar-1' }
            },
            {
                key: 'foo-2',
                // value is not validated against schema
                value: 'hello mokapi'
            }
        ]
    })
}
```

```yaml tab=kafka.yaml
asyncapi: '2.0.0'
info:
  title: Example
  version: 1.0.0
servers:
  mokapi:
    url: localhost:9092
    protocol: kafka
channels:
  orders:
    publish:
      message:
        $ref: '#/components/messages/Foo'
    subscribe:
      message:
        $ref: '#/components/messages/Foo'
components:
  messages:
    Foo:
      contentType: application/json
      payload:
        type: object
        properties:
          foo:
            type: string
```