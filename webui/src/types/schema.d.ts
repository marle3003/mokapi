declare interface Schema {
    name: string
    type: string
    format: string
    pattern: string
    description: string
    properties: Schema[]
    additionalProperties: {[string]: Schema}
    items: Schema
    xml: XmlSchmea
    required: string[]
    nullable: boolean
    example: any
    enum: any[]
    minimum: number | undefined
    maximum: number | undefined
    exclusiveMinimum: boolean | undefined
    exclusiveMaximum: boolean | undefined
    anyOf: Schema[]
    allOf: Schema[]
    oneOf: Schema[]
    uniqueItems: boolean
    minItems: number | undefined
    maxItems: number | undefined
    shuffleItems: boolean
    minProperties: number | undefined
    maxProperties: number | undefined
    deprecated: boolean
}

declare interface XmlSchmea {
    wrapped: boolean
    name: string
    attribute: boolean
    prefix: string
    namespace: string
    cdata: boolean
}