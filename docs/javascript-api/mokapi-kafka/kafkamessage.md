---
title: KafkaMessage
description: KafkaMessage is an object used by functions in the module mokapi/kafka
---
# KafkaMessage

KafkaMessage is an object used by functions in the module mokapi/kafka 
and contains produce-specific arguments.

| Name                 | Type                          | Description                                                                                                                                                                                                             |
|----------------------|-------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| partition (optional) | number                        | Kafka partition index. If not specified, the message will be written to any partition                                                                                                                                   |
| key  (optional)      | any                           | Kafka message key. If not specified, a random key will be generated based on the topic configuration.                                                                                                                   |
| data (optional)      | any                           | Kafka message data that is validated against schema. If data and value not specified, a random value will be generated based on the topic configuration. Object will be encoded based on the topic configuration        |
| value (optional)     | string, number, boolean, null | Kafka message value that is *not* validated against schema. If value and data not specified, a random value will be generated based on the topic configuration. Object will be encoded based on the topic configuration |
| headers (optional)   | object                        | Kafka message headers.                                                                                                                                                                                                  |

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