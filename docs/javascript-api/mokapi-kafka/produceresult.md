---
title: ProduceResult
description: ProduceResult is an object used by functions in the module mokapi/kafka
---
# ProduceResult

ProduceResult is an object used by [produce](/docs/javascript-api/mokapi-kafka/produce.md) function
and contains information of the written Kafka messages.

| Name     | Type   | Description                                                                                              |
|----------|--------|----------------------------------------------------------------------------------------------------------|
| api      | string | Name of the Kafka cluster (AsyncAPI `info.title`) where the messages were written.                       |
| topic    | string | Kafka topic name where the messages were written.                                                        |
| messages | array  | A list of [MessageResult](/docs/javascript-api/mokapi-kafka/messageresult.md) with one entry per written message, in input order. |

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
    console.log(`new kafka message written to ${res.api}/${res.topic} with offset: ${res.messages[0].offset}`)
}
```
