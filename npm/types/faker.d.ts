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

/**
 * The name of the root faker tree
 */
export const RootName = 'Faker'

/**
 * The Tree object represents a node in the faker tree
 */
export interface Tree {
    /**
     * Gets the name of the tree node
     */
    name: string

    /**
     * Tests whether the tree node supports the request.
     * @param request request - Request for a new fake value
     * @example
     * export default function() {
     *   const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
     *   const node = findByName('Strings')
     *   node.append({
     *     name: 'Frequency',
     *     test: (r) => { return r.lastName() === 'frequency' },
     *     fake: (r) => {
     *       return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
     *     }
     *   })
     *   return fake({ type: 'string' })
     * }
     */
    readonly test: (request: FakeRequest) => boolean

    /**
     * Gets a new fake value
     * @param request request - Request for a new fake value
     * @example
     * export default function() {
     *   const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
     *   const node = findByName('Strings')
     *   node.append({
     *     name: 'Frequency',
     *     test: (r) => { return r.lastName() === 'frequency' },
     *     fake: (r) => {
     *       return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
     *     }
     *   })
     *   return fake({ type: 'string' })
     * }
     */
    readonly fake: (request: FakeRequest) => JSONValue

    /**
     * Inserts a Tree objects after the last child of this tree.
     * @param node node - A Tree node to insert after the last child.
     */
    append: (node: Tree | CustomTree) => void

    /**
     * Inserts a Tree objects at a specified index position
     * @param index index - The zero-based index position of the insertion.
     * @param node node - The tree node to insert
     */
    insert: (index: number, node: Tree | CustomTree) => void

    /**
     * Removes a Tree node at the specific index position
     * @param index index - The zero-based index position to remove.
     */
    removeAt: (index: number) => void

    /**
     * Removes a Tree node with the given name
     * @param name name - The name of a node to remove
     */
    remove: (name: string) => void
}

/**
 * The CustomTree object represents a custom node in the faker tree
 */
export interface CustomTree {
    /**
     * Gets the name of the custom tree node
     */
    name: string

    /**
     * Tests whether the tree node supports the request.
     * @param request request - Request for a new fake value
     * @example
     * export default function() {
     *   const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
     *   const node = findByName('Strings')
     *   node.append({
     *     name: 'Frequency',
     *     test: (r) => { return r.lastName() === 'frequency' },
     *     fake: (r) => {
     *       return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
     *     }
     *   })
     *   return fake({ type: 'string' })
     * }
     */
    test: (r: Request) => boolean

    /**
     * Gets a new fake value
     * @param request request - Request for a new fake value
     * @example
     * export default function() {
     *   const frequencyItems = ['never', 'daily', 'weekly', 'monthly', 'yearly']
     *   const node = findByName('Strings')
     *   node.append({
     *     name: 'Frequency',
     *     test: (r) => { return r.lastName() === 'frequency' },
     *     fake: (r) => {
     *       return frequencyItems[Math.floor(Math.random()*frequencyItems.length)]
     *     }
     *   })
     *   return fake({ type: 'string' })
     * }
     */
    fake: (r: Request) => JSONValue
}

/**
 * Represents a request for a fake value
 */
export interface FakeRequest {
    /**
     * Contains a list of attribute names representing the object graph for which a new value is to be faked
     */
    names: string[]

    /**
     * Describing the structure of JSON data to be faked
     */
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

/**
 * JSON Schema defines a JSON-based format for describing the structure of JSON data
 * @example
 * {
 *   "type": "string",
 *   "format": "email"
 * }
 */
export interface JSONSchema {
    /**
     * Specifies the data type for a schema.
     */
    type: string | string[]

    /**
     * The enum keyword is used to restrict a value to a fixed set of values.
     */
    enum: unknown[]

    /**
     * The const keyword is used to restrict a value to a single value.
     */
    const: unknown

    /**
     * Contains a list of valid examples.
     */
    examples: unknown[]

    // Numbers
    /**
     * Restricts the number to a multiple of the given number
     */
    multipleOf: number | undefined

    /**
     * Restricts the number to a maximum number
     */
    maximum: number | undefined

    /**
     * Restricts the number to a exclusive maximum number
     */
    exclusiveMaximum: number | undefined

    /**
     * Restricts the number to a minimum number
     */
    minimum: number | undefined

    /**
     * Restricts the number to a exclusive minimum number
     */
    exclusiveMinimum: number | undefined

    // Strings
    /**
     * Restricts the string to a maximum length
     */
    maxLength: number | undefined

    /**
     * Restricts the string to a minimum length
     */
    minLength: number | undefined

    /**
     * The pattern keyword is used to restrict a string to a particular regular expression.
     */
    pattern: string

    /**
     * The format keyword allows for basic semantic identification of certain kinds of string values that are commonly used.
     */
    format: string

    // Arrays
    /**
     * Specifies the schema of the items in the array.
     */
    items: JSONSchema | undefined

    /**
     * Restricts the array to have a maximum length
     */
    maxItems: number | undefined

    /**
     * Restricts the array to have a minimum length
     */
    minItems: number | undefined

    /**
     * Restricts the array to have unique items
     */
    uniqueItems: boolean

    // Objects
    /**
     * Specifies the properties of an object
     */
    properties: { [name: string]: JSONSchema }

    /**
     * Restricts the object to have a maximum of properties
     */
    maxProperties: number | undefined

    /**
     * Restricts the object to have a minimum of properties
     */
    minProperties: number | undefined

    /**
     * Specifies the required properties for an object
     */
    required: string[] | undefined

    /**
     * The additionalProperties keyword is used to control the handling of extra stuff,
     * that is, properties whose names are not listed in the properties keyword or match
     * any of the regular expressions in the patternProperties keyword. By default, any
     * additional properties are allowed.
     */
    additionalProperties: boolean | JSONSchema | undefined

    /**
     * A value must be valid against all the schemas
     */
    allOf: JSONSchema[] | undefined

    /**
     * A value must be valid against any the schemas
     */
    anyOf: JSONSchema[] | undefined

    /**
     * A value must be valid against exactly one the schemas
     */
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
