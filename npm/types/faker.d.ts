declare module 'mokapi/faker' {
    /**
     * Creates a fake based on the given schema
     * @param schema
     */
    function fake(schema: Schema): any
}

declare interface Schema {
    type: "object" | "array" | "number" | "integer" | "string" | "boolean"
    format?: string
    pattern?: string
    items?: Schema
    required?: string[]
    enum?: any[]
    minimum?: number
    maximum?: number
    exclusiveMinimum?: boolean
    exclusiveMaximum?: boolean
    anyOf?: Schema[]
    allOf?: Schema[]
    oneOf?: Schema[]
    minItems?: number
    maxItems?: number
    shuffleItems?: boolean
    uniqueItems?: boolean
}