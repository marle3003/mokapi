---
title: "Javascript API: ProduceArgs"
description: ProduceArgs is an object used by functions in the module mokapi/kafka
---
# ProduceArgs

ProduceArgs is an object used by functions in the module mokapi/kafka 
and contains produce-specific arguments.

| Name                 | Type   | Description                                                                                                                                            |
|----------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| cluster (optional)   | string | Kafka cluster name                                                                                                                                     |
| topic (optional)     | string | Kafka topic name                                                                                                                                       |
| partition (optional) | number | Kafka partition index                                                                                                                                  |
| key (optional)       | any    | Kafka message key, if null Mokapi generates a random key based on the topic configuration                                                              |
| value (optional)     | any    | Kafka message value, if null Mokapi generates a random value based on the topic configuration. Object will be encoded based on the topic configuration |

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