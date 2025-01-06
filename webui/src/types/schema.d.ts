declare interface Schema {
  type: string[] | string
  items: Schema | undefined
  deprecated: boolean
}

declare interface SchemaFormat {
  format?: string
  schema: Schema
}
