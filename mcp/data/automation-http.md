# HTTP (OpenAPI)

Interfaces for exploring HTTP endpoints, parameters, and status codes.

```typescript
interface Http extends ApiSummary {
    servers: { url: string, description: string }

    /**
     * Returns all operations of this API.
     */
    getOperations(): HttpOperationSummary[];

    /**
     * Returns details about specific operation
     * Use id from the operation summary list
     * @param id The id of the operation
     */
    getOperation(id: string): HttpOperation
}

interface HttpOperationSummary {
    /** generated from method and path if missing in spec */
    id: string
    method: string;
    path: string;
    summary: string;
    /** Names of required parameters to help decide if this is the right endpoint */
    parameters?: string[]
}

interface HttpOperation {
    id: string;
    path: string
    method: string
    summary: string
    description: string;
    parameters: HttpRequestParameter[]
    requestBody: HttpRequestBody
    /**
     * List of allowed responses.
     * IMPORTANT: You must only use these status codes for this operation!
     */
    responses: HttpResponse[]

    /**
     * Invoke this operation against the mocked API.
     * @example operation.invoke({ path: { id: 1 }, body: JSON.stringify({ name: "test" }) })
     */
    invoke(request?: HttpInvokeRequest): HttpInvokeResponse;
}

interface HttpRequestParameter {
    name: string
    in: 'path' | 'query' | 'headers'
    required: boolean
    schema: Schema
    description?: string
}

interface HttpRequestBody {
    description: string
    required: boolean
    contents: Content[]
}

interface HttpContent {
    contentType: string
    schema: Schema

    /**
     * Generates mock data based on the HttpContent schema.
     * Use this function to generate valid data. You can adjust the value.
     * @returns {any} The generated example data.
     */
    generateExample(): any
}

interface HttpResponse {
    statusCode: number
    description: string
    contents: HttpContent[]
}

interface HttpInvokeRequest {
    path?: Record<string, string>;
    query?: Record<string, string>;
    header?: Record<string, string[]>;
    body?: string;
}

interface HttpInvokeResponse {
    statusCode: number;
    headers: Record<string, string[]>;
    body: string
}
```