// The body of `POST /api/faker/fake`, both parts are optional.
export interface FakeRequest {
    path?: string[]
    schema?: unknown
}

// One generated value: where it is in the result, which node produced it (a node path,
// "fakeBySchema" for the fallback, "-" for a value that is not from a node) and with which tokens.
export interface FakeMatch {
    propertyPath: string
    node: string
    tokens: string[]
    // "node" or the reason of a value that is not from a node: enum, const, examples, context, null
    source: string
    value: unknown
}

export interface FakeResult {
    value: unknown
    matches: FakeMatch[]
}

// "user.email" is the path ["user", "email"], each part is one path segment of the generator.
export function splitPropertyPath(text: string): string[] {
    return text.split('.').map(segment => segment.trim()).filter(segment => segment !== '')
}

// An empty field is no schema. Invalid JSON is reported to the user without a request.
export function parseSchema(text: string): { schema?: unknown } | { error: string } {
    if (text.trim() === '') {
        return {}
    }
    try {
        return { schema: JSON.parse(text) }
    } catch (e) {
        return { error: `Invalid JSON: ${(e as Error).message}` }
    }
}

export function fakeRequest(pathText: string, schemaText: string): { request: FakeRequest } | { error: string } {
    const path = splitPropertyPath(pathText)
    const parsed = parseSchema(schemaText)
    if ('error' in parsed) {
        return parsed
    }
    return { request: { path: path.length > 0 ? path : undefined, schema: parsed.schema } }
}

export function fakeRequestInit(request: FakeRequest): RequestInit {
    return {
        method: 'POST',
        body: JSON.stringify(request),
        headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
    }
}

// Only a value from a node of the tree has a node to select.
export function isTreeNode(match: FakeMatch): boolean {
    return match.node.startsWith('/')
}

// Why a value does not come from a node ("from enum"), nothing for a value from a node.
export function sourceNote(match: FakeMatch): string {
    return match.source === 'node' ? '' : `from ${match.source}`
}

// useFetch reports an HTTP error as "Error: <status text>: <body>", the generator's message is the body.
export function generatorError(fetchError: string): string {
    return fetchError.replace(/^Error: [^:]*: /, '')
}
