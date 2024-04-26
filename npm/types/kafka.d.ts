import { JSONValue } from ".";

/**
 * Sends a single message to a Kafka topic.
 * https://mokapi.io/docs/javascript-api/mokapi-kafka/produce
 * @param args - ProduceArgs object contains Kafka produce arguments.
 * @returns The produce result.
 * @example
 * export default function() {
 *   const res = produce({
 *     topic: 'foo',
 *     messages: [
 *       {
 *         key: 'foo-1',
 *         data: { foo: 'bar' }
 *       }
 *     ]
 *   });
 *   console.log(`new kafka message written with offset: ${res.offset}`)
 * }
 */
export function produce(args: ProduceArgs): ProduceResult;

/**
 * Contains produce-specific arguments.
 * https://mokapi.io/docs/javascript-api/mokapi-kafka/produceargs
 * @example
 * export default function() {
 *   const res = produce({
 *     topic: 'foo',
 *     messages: [
 *       {
 *         key: 'foo-1',
 *         data: { foo: 'bar-1' }
 *       },
 *       {
 *         key: 'foo-2',
 *         data: { foo: 'bar-2' }
 *       }
 *     ]
 *   });
 * }
 */
export interface ProduceArgs {
    /** Kafka cluster name. Used when topic name is not unique. */
    cluster?: string;

    /** Kafka topic name. If not specified, message will be written to a random topic. */
    topic?: string;

    /** Kafka partition index. If not specified, the message will be written to any partition */
    partition?: number;

    /** Kafka message key. If not specified, a random key will be generated based on the topic configuration. */
    key?: JSONValue;

    /** Kafka message value. If not specified, a random value will be generated based on the topic configuration. */
    value?: JSONValue;

    /** Kafka message headers. */
    headers?: { [name: string]: JSONValue };

    messages: KafkaMessage[]
}

/**
 * Represents a Kafka message
 */
export interface KafkaMessage {
    /** Kafka partition index. If not specified, the message will be written to any partition */
    partition?: number;

    /** Kafka message key. If not specified, a random key will be generated based on the topic configuration. */
    key?: JSONValue;

    /** Kafka message value. If data and value are not specified, a random value will be generated based on the topic configuration. */
    data?: JSONValue;

    /** Kafka message value not validating against schema. If data and value are not specified, a random value will be generated based on the topic configuration. */
    value?: string | number | boolean | null

    /** Kafka message headers. */
    headers?: { [name: string]: JSONValue };
}

/**
 * Contains information of the written Kafka message.
 * https://mokapi.io/docs/javascript-api/mokapi-kafka/produceresult
 * @example
 * export default function() {
 *   const res = produce({ topic: 'foo' })
 *   console.log(`new kafka message written with offset: ${res.offset}`)
 * }
 */
export interface ProduceResult {
    /** Name of the Kafka cluster where the message was written. */
    readonly cluster: string;

    /** Kafka topic name where the message was written. */
    readonly topic: string;

    messages: KafkaMessageResult[]

    /** Kafka partition where the message was written. */
    readonly partition: number;

    /** The offset of the written message. */
    readonly offset: number;

    /** The key of the written message, */
    readonly key: string;

    /** The value of the written message. */
    readonly value: string;

    /** The headers of the written message */
    readonly headers: { [name: string]: string };
}

/**
 * Contains information of the written Kafka message.
 * https://mokapi.io/docs/javascript-api/mokapi-kafka/kafkamessageresult
 */
export interface KafkaMessageResult {
    /**
     * Kafka partition index in which the message was written.
     */
    readonly partition: number

    /**
     * Kafka offset of the written message.
     */
    readonly offset: number

    /**
     * Kafka written message key.
     */
    readonly key: string

    /**
     * Kafka written message value.
     */
    readonly value: string

    /**
     * Kafka written message headers.
     */
    readonly headers: { [name: string]: string }

}

/**
 * KafkaEventHandler is a function that is executed when a Kafka message is received.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventHandler
 * @example
 * export default function() {
 *   on('kafka', function(msg) {
 *     // add header 'foo' to every Kafka message
 *     msg.headers = { foo: 'bar' }
 *   })
 * }
 */
export type KafkaEventHandler = (message: KafkaEventMessage) => boolean;

/**
 * KafkaEventMessage is an object used by KafkaEventHandler that contains Kafka-specific message data.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventMessage
 */
export interface KafkaEventMessage {
    /** Kafka partition where the message was written to (read-only). */
    readonly offset: number;

    /** Kafka message key  */
    key: string;

    /** Kafka message value */
    value: string;

    /** Kafka message headers */
    headers: { [name: string]: string } | null;
}
