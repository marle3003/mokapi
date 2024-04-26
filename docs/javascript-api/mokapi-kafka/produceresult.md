---
title: ProduceResult
description: ProduceResult is an object used by functions in the module mokapi/kafka
---
# ProduceResult

ProduceResult is an object used by functions in the module mokapi/kafka 
and contains information of the written Kafka message.

| Name                   | Type                                                                | Description                                               |
|------------------------|---------------------------------------------------------------------|-----------------------------------------------------------|
| cluster                | string                                                              | Name of the  Kafka cluster where the message was written. |
| topic                  | string                                                              | Kafka topic name where the message was written.           |
| messages               | [KafkaMessage[]](/docs/javascript-api/mokapi-kafka/kafkamessage.md) | A list of Kafka written messages                          |
| partition (deprecated) | number                                                              | Kafka partition where the message was written.            |
| offset (deprecated)    | number                                                              | The offset of the written message.                        |
| key (deprecated)       | string                                                              | The key of the written message.                           |
| value (deprecated)     | string                                                              | The value of the written message.                         |

## Example

```javascript
import { produce } from 'mokapi/kafka'

export default function() {
    const res = produce({
        topic: 'foo',
        messages: [
            {
                key: 'foo-1',
                value: `{"foo": "bar"}`
            }
        ]
    })
    console.log(`new kafka message written with offset: ${res.offset}`)
}
```