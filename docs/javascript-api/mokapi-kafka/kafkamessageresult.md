---
title: KafkaMessageResult
description: KafkaMessageResult is an object used by functions in the module mokapi/kafka
---
# KafkaMessageResult

KafkaMessage is an object used by functions in the module mokapi/kafka 
and contains written message result.

| Name      | Type   | Description                                             |
|-----------|--------|---------------------------------------------------------|
| partition | number | Kafka partition index in which the message was written. |
| offset    | number | Kafka offset of the written message.                    |
| key       | string | Kafka written message key.                              |
| value     | string | Kafka written message value.                            |
| headers   | object | Kafka written message headers.                          |

## Example

```javascript tab=kafka.js
import { produce } from 'mokapi/kafka'

export default function() {
    const result = produce({
        topic: 'foo',
        messages: [
            {
                key: 'foo-1',
                data: { foo: 'bar-1' }
            },
        ]
    })
    console.log(result.messages[0].value)
}
```