# produce( [args] )

Sends a single message to a Kafka topic.

| Parameter       | Type   | Description                                                                                           |
|-----------------|--------|-------------------------------------------------------------------------------------------------------|
| args (optional) | object | [ProduceArgs](/docs/javascript-api/mokapi-kafka/produceargs.md)  object contains definition of a type |

## Returns

| Type   | Description                                                         |
|--------|---------------------------------------------------------------------|
| object | [ProduceResult](/docs/javascript-api/mokapi-kafka/produceresult.md) |

## Examples

```javascript
import { produce } from 'mokapi/kafka'

export default function() {
    const result = produce({
        topic: 'topic', 
        value: 'value', 
        key: 'key',
        partition: 2
    })
    console.log(`offset=${result.offset}, key=${result.key}, value=${result.value}`)
}
```