---
title: KafkaEventMessage
description: KafkaEventMessage is an object used by KafkaEventHandler and contains Kafka-specific message data.
---
# KafkaEventMessage

`KafkaEventMessage` is the object passed to a [KafkaEventHandler](/docs/javascript-api/mokapi/eventhandler/kafkaeventhandler.md).
It represents a single Kafka message and gives you access to its topic, partition, offset, key,
value, and headers. You can read these fields to decide how to handle the message, and modify
key, value, and headers to change the message before it's stored.

| Name       | Type     | Description                                                                              |
|------------|----------|------------------------------------------------------------------------------------------|
| api        | string   | The name of the API, matching `info.title` in the AsyncAPI specification (read-only).    |
| topic      | string   | The name of the Kafka topic this message belongs to (read-only).                         |
| partition  | number   | The partition index within the topic (read-only).                                        |
| offset     | number   | Kafka offset of the kafka message (read-only).                                           |
| key        | string   | Kafka message key                                                                        |
| value      | string   | Kafka message value                                                                      |
| headers    | object   | Kafka message headers                                                                    |
| schemaId   | number   | The identifier of the schema used to validate or serialize the payload. See note below.  |

``` box=info
`schemaId` is only populated automatically when the schema ID is embedded in the message payload
itself (`SchemaIdLocation: "payload"`). If the schema ID is transmitted via message metadata
instead (`SchemaIdLocation: "header"`), `schemaId` will not be set, and you'll need to read it
from `headers` yourself.
```

## Example

```javascript
import { on } from 'mokapi'

export default function() {
    on('kafka', function(message) {
        console.log(`Topic: ${message.topic}, Partition: ${message.partition}, Offset: ${message.offset}`)
        console.log(`Key: ${message.key}, Value: ${message.value}`)
        // add header 'foo' to every Kafka message
        message.headers['correlationId'] = 'abc-123'
    })
}
```

## See Also

- [KafkaEventHandler](/docs/javascript-api/mokapi/eventhandler/kafkaeventhandler.md)