---
title: produce( [args] )
description: Send a single message to a Kafka topic.
---
# produce( [args] )

Sends a Kafka message to a Kafka topic.

| Parameter       | Type   | Description                                                                                               |
|-----------------|--------|-----------------------------------------------------------------------------------------------------------|
| args (optional) | object | [ProduceArgs](/docs/javascript-api/mokapi-kafka/produceargs.md) object contains Kafka produce arguments   |

## Returns

| Type   | Description                                                         |
|--------|---------------------------------------------------------------------|
| object | [ProduceResult](/docs/javascript-api/mokapi-kafka/produceresult.md) |

## Examples

Use method produce to publish messages to Kafka topics.

```javascript
import { produce } from 'mokapi/kafka'

export default function() {
    const result = produce({
        topic: 'topic', 
        messages: [
            { key: 'key1', value: 'hello Mokapi' },
            { key: 'key2', value: 'hello world' }
        ],
    })
    logMessage(result.messages[0])
    logMessage(result.messages[1])
}

function logMessage(msg) {
    console.log(`offset=${msg.offset}, key=${msg.key}, value=${msg.value}`)
}
```

Example with defined partitions

```javascript
import { produce } from 'mokapi/kafka'

export default function() {
    const result = produce({
        topic: 'topic', 
        messages: [
            { key: 'key1', value: 'hello Mokapi', partition: 0 },
            { key: 'key2', value: 'hello world', partition: 1 }
        ],
    })
}
```

Example with data that Mokapi serializes and validates against the schema

```javascript
import { produce } from 'mokapi/kafka'

export default function() {
    const result = produce({
        topic: 'topic', 
        messages: [
            { 
                key: 'key1',
                data: {
                    text: 'hello Mokapi'
                },
                partition: 0
            },
            { 
                key: 'key2',
                data: { 
                    text: 'hello world'
                }, 
                partition: 1
            }
        ],
    })
}
```

Example with headers

```javascript
import { produce } from 'mokapi/kafka'

export default function() {
    const result = produce({
        topic: 'topic',
        messages: [
            { 
                key: 'key1',
                data: {
                    text: 'hello Mokapi'
                },
                headers: {
                    'system-id': 'my-producer',
                    'message-type': 'greetings'
                }
            },
        ],
    })
}
```