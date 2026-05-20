# Kafka (AsyncAPI)

Interfaces for inspecting topics, partitions, and message history.

```typescript
interface Kafka extends ApiSummary {
    brokers: { name: string, host: string, description?: string }

    getTopics(): KafkaTopicSummary[]
    getTopic(topicName: string): KafkaTopic
}

interface KafkaTopicSummary {
    /**
     * The unique name of the topic.
     */
    name: string
    title?: string
    summary?: string
}

interface KafkaTopic extends KafkaTopicSummary {
    description: string
    /**
     * List of current partitions and their maximum offsets.
     * Use this to determine the range for the 'consume' method.
     * The 'offset' represents the NEXT available position.
     * (e.g., if offset is 10, messages 0-9 exist; the next message will be 10).
     */
    partitions: { index: number, offset: number }

    operations: KafkaOperation[];

    /**
     * Use 'produce' to send a message to this topic.
     *
     * Use 'operations' with action 'send' to inspect the expected message schema
     * and valid payload examples for this topic.
     * 
     * Mokapi automatically encodes the payload based on the topic configuration:
     * - JSON topics expect a JSON-compatible object
     * - Avro topics expect a value matching the Avro schema
     * 
     * @param partition The target partition index. MUST be one of the indices listed in the 'partitions' array.
     * @param value The message payload. Provide a value that matches the topic message schema.
     * @param key Optional message key.
     * @param headers Optional metadata headers.
     */
    produce(partition: number, value: any, key?: string, headers?: KafkaHeader): void

    /**
     * INSPECT: Retrieves a specific record for analysis or verification.
     * Use this to check if the mock has received or produced a specific message.
     * @param partition The partition index (see 'partitions' list).
     * @param startOffset The offset to start reading from (see 'partitions' for max offset).
     * @param limit The maximum number of records to return in this call.
     * @returns An array of records found starting from the startOffset.
     */
    consume(partition: number, startOffset: number, limit: number): KafkaRecord[]
}

interface KafkaOperation {
    action: 'send' | 'receive'
    title: string
    summary: string;
    description: string
    messages: KafkaMessage[];
}

interface KafkaMessage {
    name: string;
    title: string;
    summary: string
    description: string
    contentType: string
    /**
     * JSON schema describing the Kafka message payload.
     * Use fake(payload) to generate valid example data.
     */
    payload: Schema;
    key: Schema
    headers?: Schema;
}

interface KafkaHeader {
    [name: string]: string
}

interface KafkaRecord {
    offset: number
    key: string
    value: string
    headers: KafkaHeader
}
```

## Examples

Read the very last existing message from a specific partition. Since offset is the NEXT position, we read at offset - 1.
```typescript
const kafka = mokapi.getApi("...");
const topic = kafka.getTopic("...");
const p0 = topic.partitions.find(p => p.index === 0);
let lastRecord = null
// Check if partition exists and contains at least one message
if (p0 && p0.offset > 0) {
    lastRecord = topic.consume(0, p0.offset - 1, 1);
}
lastRecord
```

Write a message
```typescript
const kafka = mokapi.getApi("...");
const topic = kafka.getTopic("...");
topic.produce(0, JSON.stringify({ foo: "123" }), "key-123");
```

Determine last offset. The last offset represents the NEXT available position.
```typescript
const kafka = mokapi.getApi("Kafka Server");
const topic = kafka.getTopic("topic-name");
const partition = topic.partitions.find(p => p.index === 0);
partition.offset;
```