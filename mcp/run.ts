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
     * Generate a random value from a JSON Schema.
     * @example
     * fake({ type: 'string', format: 'email' })
     */
    fake(schema: Schema): any;
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
    method: string
    path: string
    summary: string
    description: string;
    parameters: RequestParameter[]
    requestBody: RequestBody

    /**
     * Invoke this operation against the mocked API.
     */
    invoke(request?: InvokeRequest): InvokeResponse;

    /**
     * Returns the response schema for a given status code.
     */
    getResponseSchema(statusCode: number): Response[];
}

interface RequestParameter {
    name: string
    in: 'path' | 'query' | 'headers'
    required: boolean
    schema: Schema
    description?: string
}

interface RequestBody {
    description: string
    required: boolean
    contents: Content[]
}

interface Content {
    contentType: string
    schema: Schema
}

interface Response {
    statusCode: number
    description: string
    contents: Content[]
}

interface InvokeRequest {
    path?: Record<string, string>;
    query?: Record<string, string>;
    header?: Record<string, string[]>;
    body?: string;
}

interface InvokeResponse {
    statusCode: number;
    headers: Record<string, string[]>;
    body?: string;
}

/**
 * JSON Schema defines a JSON-based format for describing the structure of JSON data
 * @example
 * {
 *   "type": "string",
 *   "format": "email"
 * }
 */
interface Schema {
    /**
     * Specifies the data type for a schema.
     */
    type?: SchemaType | SchemaType[];

    /**
     * The enum keyword is used to restrict a value to a fixed set of values.
     */
    enum?: any[];

    /**
     * The const keyword is used to restrict a value to a single value.
     */
    const?: any;

    /**
     * Contains a list of valid examples.
     */
    examples?: any[];

    /**
     * Specifies a default value.
     */
    default?: any;

    // Numbers
    /**
     * Restricts the number to a multiple of the given number
     */
    multipleOf?: number;

    /**
     * Restricts the number to a maximum number
     */
    maximum?: number;

    /**
     * Restricts the number to an exclusive maximum number
     */
    exclusiveMaximum?: number;

    /**
     * Restricts the number to a minimum number
     */
    minimum?: number;

    /**
     * Restricts the number to an exclusive minimum number
     */
    exclusiveMinimum?: number;

    // Strings
    /**
     * Restricts the string to a maximum length
     */
    maxLength?: number;

    /**
     * Restricts the string to a minimum length
     */
    minLength?: number;

    /**
     * The pattern keyword is used to restrict a string to a particular regular expression.
     */
    pattern?: string;

    /**
     * The format keyword allows for basic semantic identification of certain kinds of string values that are commonly used.
     */
    format?: string;

    // Arrays
    /**
     * Specifies the schema of the items in the array.
     */
    items?: Schema;

    /**
     * Restricts the array to have a maximum length
     */
    maxItems?: number;

    /**
     * Restricts the array to have a minimum length
     */
    minItems?: number;

    /**
     * Restricts the array to have unique items
     */
    uniqueItems?: boolean;

    // Objects
    /**
     * Specifies the properties of an object
     */
    properties?: { [name: string]: Schema };

    /**
     * Restricts the object to have a maximum of properties
     */
    maxProperties?: number;

    /**
     * Restricts the object to have a minimum of properties
     */
    minProperties?: number;

    /**
     * Specifies the required properties for an object
     */
    required?: string[];

    /**
     * The additionalProperties keyword is used to control the handling of extra stuff,
     * that is, properties whose names are not listed in the properties keyword or match
     * any of the regular expressions in the patternProperties keyword. By default, any
     * additional properties are allowed.
     */
    additionalProperties?: boolean | Schema;

    /**
     * A value must be valid against all the schemas
     */
    allOf?: Schema[];

    /**
     * A value must be valid against any the schemas
     */
    anyOf?: Schema[];

    /**
     * A value must be valid against exactly one the schemas
     */
    oneOf?: Schema[];
}

type SchemaType = "object" | "array" | "number" | "integer" | "string" | "boolean" | "null";