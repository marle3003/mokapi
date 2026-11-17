// A small copy of the faker tree of Mokapi and the answers of the faker endpoints of the dashboard
// API. The values are fixed so tests can assert them. Node paths, tokens and sources are what the
// real endpoints return (checked against `generator.Explain`): the property path of a nested value
// starts with the path of the request, a property resolves to a top-level node (`address.city` is
// `/city`), an array item has the segment `[]`, an enum has no node.

interface Node {
    name: string
    path: string
    attributes?: string[]
    weight?: number
    dependsOn?: string[]
    children?: Node[]
    custom?: boolean
    hasFakeFunc?: boolean
}

function leaf(parent: string, name: string, attributes: string[], extra: Partial<Node> = {}): Node {
    return { name, path: `${parent}/${name}`, attributes, hasFakeFunc: true, ...extra }
}

// Same names in different branches: email and name (also under person and contact), code (under
// the custom nodes billing and shipping, which only group). frequency is a custom node.
export const tree: Node = {
    name: 'root',
    path: '/',
    children: [
        leaf('', 'name', ['name']),
        leaf('', 'email', ['email'], { dependsOn: ['firstname', 'lastname'] }),
        leaf('', 'city', ['city', 'locality']),
        leaf('', 'tag', ['tag']),
        {
            ...leaf('', 'person', ['person', 'owner']),
            children: [
                leaf('/person', 'name', ['name', 'fullName'], { dependsOn: ['firstname', 'lastname'] }),
                leaf('/person', 'firstname', ['firstname', 'firstname2'], { dependsOn: ['gender', 'sex'], weight: 1 }),
                leaf('/person', 'lastname', ['lastname', 'surname']),
            ]
        },
        {
            ...leaf('', 'contact', ['contact'], { dependsOn: ['firstname', 'lastname'] }),
            children: [
                leaf('/contact', 'phone', ['phone', 'phoneNumber'], { weight: 0.5 }),
                leaf('/contact', 'email', ['email'], { weight: 0.5 }),
                leaf('/contact', 'type', ['type']),
            ]
        },
        {
            ...leaf('', 'address', ['address']),
            children: [
                leaf('/address', 'street', ['street']),
                leaf('/address', 'zip', ['zip', 'zipCode', 'postcode', 'postal', 'postalCode']),
            ]
        },
        leaf('', 'frequency', ['frequency'], { custom: true }),
        {
            name: 'billing',
            path: '/billing',
            custom: true,
            children: [leaf('/billing', 'code', ['code'], { custom: true })]
        },
        {
            name: 'shipping',
            path: '/shipping',
            custom: true,
            children: [leaf('/shipping', 'code', ['code'], { custom: true })]
        },
    ]
}

const fallbackNode = 'fakeBySchema'
const noNode = '-'

const sampleValues: Record<string, unknown> = {
    '/name': 'Jane',
    '/email': 'jane.doe@example.com',
    '/city': 'Bern',
    '/tag': 'urgent',
    '/frequency': 'weekly',
    '/person': { firstname: 'Jane', lastname: 'Doe' },
    '/person/name': 'Jane Doe',
    '/person/firstname': 'Jane',
    '/person/lastname': 'Doe',
    '/contact': { phone: '+41 31 555 01 01', email: 'contact@example.com' },
    '/contact/phone': '+41 31 555 01 01',
    '/contact/email': 'contact@example.com',
    '/contact/type': 'work',
    '/address': { street: 'Bahnhofstrasse', zip: '3011' },
    '/address/street': 'Bahnhofstrasse',
    '/address/zip': '3011',
}

// Their value is numbered, so Generate gives a new value every time (BILL-1, BILL-2, ...).
const numberedPrefixes: Record<string, string> = {
    '/billing/code': 'BILL',
    '/shipping/code': 'SHIP',
}
const generated: Record<string, number> = {}

// The node the generator picks for a property name, tokens are the name it was matched with. A
// property path that is the path of a node (contact.email) matches this node with all its segments.
const nodeOfName: Record<string, { node: string, tokens: string[] }> = {
    name: { node: '/name', tokens: ['name'] },
    email: { node: '/email', tokens: ['email'] },
    customerEmail: { node: '/email', tokens: ['email'] },
    city: { node: '/city', tokens: ['city'] },
    tags: { node: '/tag', tokens: ['tag'] },
    frequency: { node: '/frequency', tokens: ['frequency'] },
}

interface Match {
    propertyPath: string
    node: string
    tokens: string[]
    source: string
    value: unknown
}

interface JsonSchema {
    type?: string
    properties?: Record<string, JsonSchema>
    items?: JsonSchema
    enum?: unknown[]
}

const schemaTypes = ['string', 'number', 'integer', 'boolean', 'object', 'array', 'null']

export function fakeOfNode(path: string): { found: boolean, value?: unknown, error?: string } {
    const node = findNode(tree, path)
    if (!node) {
        return { found: false }
    }
    if (!node.hasFakeFunc) {
        return { found: true, error: `node ${path} has no fake function` }
    }
    const prefix = numberedPrefixes[path]
    if (prefix) {
        generated[path] = (generated[path] ?? 0) + 1
        return { found: true, value: `${prefix}-${generated[path]}` }
    }
    return { found: true, value: sampleValues[path] }
}

function findNode(node: Node, path: string): Node | undefined {
    if (node.path === path) {
        return node
    }
    return node.children?.map(c => findNode(c, path)).find(n => n)
}

export function explain(path: string[] = [], schema?: JsonSchema): { value?: unknown, matches?: Match[], error?: string } {
    if (schema?.type !== undefined && !schemaTypes.includes(schema.type)) {
        return { error: `unsupported schema: schema type=${schema.type}` }
    }
    const matches: Match[] = []
    const value = generate(path, schema ?? {}, matches)
    // the mock gets the properties of a schema in no fixed order, the real endpoint keeps the order of the schema
    matches.sort((a, b) => a.propertyPath.localeCompare(b.propertyPath))
    return { value, matches }
}

function nodeOf(at: string[]): { node: string, tokens: string[] } | undefined {
    if (at.length === 0) {
        return undefined
    }
    const path = '/' + at.join('/')
    return findNode(tree, path) ? { node: path, tokens: at } : nodeOfName[at[at.length - 1]]
}

function generate(at: string[], schema: JsonSchema, matches: Match[]): unknown {
    if (schema.type === 'object' && schema.properties) {
        const result: Record<string, unknown> = {}
        for (const [name, property] of Object.entries(schema.properties)) {
            result[name] = generate([...at, name], property, matches)
        }
        return result
    }
    const row = (propertyPath: string, node: string, tokens: string[], source: string, value: unknown) => {
        matches.push({ propertyPath, node, tokens, source, value })
        return value
    }
    const propertyPath = at.length > 0 ? at.join('.') : '(value)'

    if (schema.enum) {
        return row(propertyPath, noNode, [], 'enum', schema.enum[0])
    }
    if (schema.type === 'array') {
        const found = nodeOf(at)
        const item = found ? sampleValues[found.node] : 'item'
        return [row(`${propertyPath}[]`, found?.node ?? fallbackNode, found?.tokens ?? [], 'node', item)]
    }
    const found = nodeOf(at)
    if (found) {
        return row(propertyPath, found.node, found.tokens, 'node', sampleValues[found.node])
    }
    return row(propertyPath, fallbackNode, [], 'node', schema.type === 'integer' || schema.type === 'number' ? 42 : 'Qz8TnW')
}
