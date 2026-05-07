# Mokapi Automation API

Technical reference for exploring the current environment, APIs, and Kafka clusters.

## Category

To see detailed TypeScript interfaces, call this tool with one of the following values for the category parameter:

| Category           | Description                                                             | Parameter Name  |
|--------------------|-------------------------------------------------------------------------|-----------------|
| Core Discovery API | Access to the global mokapi object for infrastructure metadata.         | `core`          |
| HTTP (OpenAPI)     | Interfaces for exploring HTTP endpoints, parameters, and status codes.  | `http`          |
| Kafka (AsyncAPI)   | Interfaces for inspecting topics, partitions, and message history.      | `kafka`         |
| Events             | Definitions for debugging activity via getEvents() and traits.          | `event`         |

## Examples

Get all available APIs
```typescript
// The last expression is returned as the tool result
mokapi.getApis()
```

Get all available Kafka APIs
```typescript
mokapi.getApis().filter(x => x.type === 'kafka')
```

Get specific API to understand the contract
```typescript
mokapi.getApi('API_NAME')
```

Get all recorded events from Mokapi, if user asks "Why did my request fail?"
```typescript
mokapi.getEvents()
```

Get latest HTTP events for a specific API
```typescript
mokapi.getEvents({ type: 'http', name: 'Petstore' })
```

Search for all POST endpoints related to "payments".
```typescript
// 1. Search for a specific capability
const searchResponse = mokapi.search("method:POST AND path:*payments*");
let op = null
if (searchResponse.items.length > 0) {
    const match = searchResponse.items[0];

    // 2. Use metadata to load the full API specification
    const api = mokapi.getApi(match.metadata.service);

    // 3. Find the specific operation to see allowed status codes/schemas
    // Tip: Use the path from metadata to identify the correct operation
    const opSummary = api.getOperations().find(x => x.path === match.metadata.path && x.method === match.metadata.method);

    if (opSummary) {
        op = api.getOperation(opSummary.id);
    }
}
op
```

Search for errors (404 or 500) related to the Petstore
```typescript
const errors = mokapi.search('+api:"Petstore" +response.statusCode:>=400 +type:event')
errors.items.map(x => mokapi.getEvent(x.metadata.id))
```
