declare const mokapi: Mokapi;

interface Mokapi {
    /**
     * Returns all mocked APIs (lightweight, no schemas).
     */
    getApis(): ApiSummary[];

    /**
     * Returns a specific API by name.
     */
    getApi(name: string): Api;

    /**
     * Directly invoke an API operation.
     */
    invoke(apiName: string, operationId: string, request?: InvokeRequest): InvokeResponse;
}

interface ApiSummary {
    name: string;
    type: 'openapi' | 'asyncapi' | 'ldap' | 'mail';
}

interface Api extends ApiSummary {
    spec: OpenApi
}

interface OpenApi {
    /**
     * Returns operations of this API.
     * By default, only minimal fields are returned.
     */
    getOperations(): OperationSummary[];

    getOperationDetails(path: string, method: string): OperationDetails
}

interface OperationSummary {
    method: string;
    path: string;
    summary: string;
}

interface OperationDetails {
    operationId: string;
    description: string;

    /**
     * Invoke this operation against the mocked API.
     */
    invoke(request?: InvokeRequest): InvokeResponse;

    /**
     * Returns the response schema for a given status code.
     */
    getResponseSchema(status: number): Schema;
}

interface InvokeRequest {
    path?: Record<string, string | number>;
    query?: Record<string, string | number>;
    headers?: Record<string, string>;
    body?: any;
}

interface InvokeResponse {
    status: number;
    headers: Record<string, string[]>;
    body?: any;
}

interface Schema {
    type?: string;
    properties?: Record<string, Schema>;
}