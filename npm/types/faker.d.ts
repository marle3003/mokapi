import { JSONValue } from ".";

/**
 * Creates a fake based on the given schema
 * https://mokapi.io/docs/javascript-api/mokapi-faker/fake
 * @param schema schema - OpenAPI Schema object contains definition of a type
 * @example
 * export default function() {
 *   console.log(fake({type: 'string'}))
 *   console.log(fake({type: 'number'}))
 *   console.log(fake({type: 'string', format: 'date-time'}))
 *   console.log(fake({type: 'string', pattern: '^\d{3}-\d{2}-\d{4}$'})) // 123-45-6789
 * }
 */
export function fake(schema: Schema | FakeRequest): JSONValue;

/**
 * Gets the tree node with the given name
 * @param name name - tree node's name
 * @example
 * export default function() {
 *   const root = findByName(RootName)
 *   root.insert(0, { name: 'foo', test: () => { return true }, fake: () => { return 'foobar' } })
 *   console.log(fake({type: 'string'}))
 * }
 */
export function findByName(name: string): Tree

export const RootName = 'Faker'

export interface Tree {
    name: string
    readonly test: () => boolean
    readonly fake: () => JSONValue

    append: (node: Tree | CustomTree) => void
    insert: (index: number, node: Tree | CustomTree) => void
    removeAt: (index: number) => void
    remove: (name: string) => void
}

export interface CustomTree {
    name: string
    test: (r: Request) => boolean
    fake: (r: Request) => JSONValue
}

export interface FakeRequest {
    names: string[]
    schema: JSONSchema
}

export interface Request {
    Path: PathElement[]

    last: () => PathElement
    lastName: () => string
    lastSchema: () => JSONSchema
}

export interface PathElement {
    name: string
    schema: JSONSchema
}

export interface JSONSchema {
    type: string | string[]
    enum: unknown[]
    const: unknown
    examples: unknown[]

    // Numbers
    multipleOf: number | undefined
    maximum: number | undefined
    exclusiveMaximum: number | undefined
    minimum: number | undefined
    exclusiveMinimum: number | undefined

    // Strings
    maxLength: number | undefined
    minLength: number | undefined
    pattern: string
    format: string

    // Arrays
    items: JSONSchema | undefined
    maxItems: number | undefined
    minItems: number | undefined
    uniqueItems: boolean

    // Objects
    properties: { [name: string]: JSONSchema }
    maxProperties: number | undefined
    minProperties: number | undefined
    required: string[] | undefined
    additionalProperties: boolean | JSONSchema | undefined

    allOf: JSONSchema[] | undefined
    anyOf: JSONSchema[] | undefined
    oneOf: JSONSchema[] | undefined

    isAny: () => boolean
    isString: () => boolean
    isInteger: () => boolean
    isNumber: () => boolean
    isArray: () => boolean
    isObject: () => boolean
    isNullable: () => boolean
    isDictionary: () => boolean
    is: (typeName: string) => boolean
}

export interface Schema {
    /** Type of fake value */
    type?: "object" | "array" | "number" | "integer" | "string" | "boolean";

    /** Serves as a hint at the contents and format of the string.  */
    format?: string;

    /** Specifies regular expression that fake must match.  */
    pattern?: string;

    /** Describes the type and format of array items. */
    items?: Schema;

    /** Specifies a list of required properties. */
    required?: string[];

    /** Specify possible value for this schema. */
    enum?: JSONValue[];

    /** Specifies the minimum range of possible values. */
    minimum?: number;

    /** Specifies the maximum range of possible values. */
    maximum?: number;

    /** Specifies whether minimum value is exluded. Default is false. */
    exclusiveMinimum?: boolean;

    /** ** Specifies whether maximum value is exluded. Default is false. */
    exclusiveMaximum?: boolean;

    /** Valid against one of the specified schemas. */
    anyOf?: Schema[];

    /** Valid against all of the specified schemas. */
    allOf?: Schema[];

    /** Valid against exactly one of the specified schemas. */
    oneOf?: Schema[];

    /** Specifies the minimum number of items an array must have. */
    minItems?: number;

    /** Specifies the maximum number of items an array must have. */
    maxItems?: number;

    /** Specifies whether items in array should have random order. The default is false */
    shuffleItems?: boolean;

    /** Specifies whether all items in an array must be unique. */
    uniqueItems?: boolean;
}
