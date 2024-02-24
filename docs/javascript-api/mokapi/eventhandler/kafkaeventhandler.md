---
title: KafkaEventHandler
description: KafkaEventHandler is a function that is executed when a producer sends a Kafka message.
---
# KafkaEventHandler

KafkaEventHandler is a function that is executed when an HTTP event is triggered.

| Parameter | Type    | Description                                                                                                             |
|-----------|---------|-------------------------------------------------------------------------------------------------------------------------|
| record    | object  | [KafkaRecord](/docs/javascript-api/mokapi/eventhandler/kafkarecord.md) object contains data of a Kafka produce message. |

## Returns

| Type    | Description                                           |
|---------|-------------------------------------------------------|
| boolean | Whether Mokapi should log execution of event handler. |

## Example

```javascript
import { on } from 'mokapi'

export default function() {
    on('kafka', function(record) {
        // add header 'foo' to every Kafka message
        record.headers = { foo: 'bar' }
    })
}
```