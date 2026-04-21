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