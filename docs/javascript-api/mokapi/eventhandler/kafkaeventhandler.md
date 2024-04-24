---
title: KafkaEventHandler
description: KafkaEventHandler is a function that is executed when a producer sends a Kafka message.
---
# KafkaEventHandler

KafkaEventHandler is a function that is executed when a Kafka message is received.

| Parameter | Type    | Description                                                                                                               |
|-----------|---------|---------------------------------------------------------------------------------------------------------------------------|
| message   | object  | [KafkaMessage](/docs/javascript-api/mokapi/eventhandler/kafkamessage.md) object contains data of a Kafka produce message. |

## Returns

| Type    | Description                                           |
|---------|-------------------------------------------------------|
| boolean | Whether Mokapi should log execution of event handler. |

## Example

```javascript
import { on } from 'mokapi'

export default function() {
    on('kafka', function(message) {
        // add header 'foo' to every Kafka message
        message.headers = { foo: 'bar' }
    })
}
```