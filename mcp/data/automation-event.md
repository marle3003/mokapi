# Events

Definitions for debugging activity via getEvents() and traits. 

```typescript
interface Traits {
    type?: ApiType;
    name?: string;
}

interface HttpTraits extends Traits {
    /**
     * Path value specified by the OpenAPI path
     * @example /pet/{petId}
     */
    path?: string

    /**
     * Request method.
     * @example GET
     */
    method?: string
}

interface KafkaTraits extends Traits {
    /**
     * Topic name specified by the AsyncAPI channel
     * @example user.signedup
     */
    topic?: string

    /**
     * Partition in which message was written
     */
    partition?: number

    /**
     * Which client produced the message
     * The value 'mokapi-script' indicates the message was produced by a Mokapi Script
     * The value 'mokapi-mcp' indicates the messages was produced by MCP server
     */
    clientId?: string
}

interface Event {
    /**
     * ID of the event
     */
    id: string;

    /**
     * List of traits
     */
    traits: Traits;

    /**
     * The data of the event
     */
    data: any

    /**
     * Time of the event in the format RFC3339
     * @example 2026-07-21T17:32:28Z
     */
    time: string
}
```