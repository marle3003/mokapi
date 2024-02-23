---
title: "Javascript API: ProduceArgs"
description: ProduceArgs is an object used by functions in the module mokapi/kafka
---
# ProduceArgs

ProduceArgs is an object used by functions in the module mokapi/kafka 
and contains produce-specific arguments.

| Name                 | Type   | Description                                                                                                                                                       |
|----------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| cluster (optional)   | string | Kafka cluster name. Used when topic name is not unique.                                                                                                           |
| topic (optional)     | string | Kafka topic name. If not specified, message will be written to a random topic.                                                                                    |
| partition (optional) | number | /** Kafka partition index. If not specified, the message will be written to any partition                                                                         |
| key (optional)       | any    | Kafka message key. If not specified, a random key will be generated based on the topic configuration.                                                             |
| value (optional)     | any    | Kafka message value. If not specified, a random value will be generated based on the topic configuration. Object will be encoded based on the topic configuration |
| headers (optional)   | object | Kafka message headers.                                                                                                                                            |

## Example

```javascript tab=kafka.js
import { produce } from 'mokapi/kafka'

export default function() {
    produce({
        topic: 'foo',
        key: 'foo-1',
        value: {foo: 'bar'}
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