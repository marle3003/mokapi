declare interface HttpService extends Service {
    paths: HttpPath[]
    servers: HttpServer[]
}

declare interface HttpPath {
    path: string
    summary: string
    description: string
    operations: HttpOperation[]
}

declare interface HttpOperation {
    method: string
    summary: string
    description: string
    operationId: string
    deprecated: boolean
    parameters: HttpParameter[]
    requestBody: HttpRequestBody
    responses: HttpResponse[]
    security: HttpSecurity[]
}

declare interface HttpParameter {
    name: string
    type: string
    description: string
    required: boolean
    deprecated: boolean
    style: string
    explode?: boolean
    schema: Schema
}

declare interface HttpEventData {
    request: HttpEventRequest
    response: HttpEventResponse
    duration: number
    deprecated: boolean
    actions: Action[]
}

declare interface HttpEventRequest {
    method: string
    url: string
    contentType: string
    parameters?: HttpEventParameter[]
    body: string
}

declare interface HttpRequestBody {
    description: string
    contents: HttpMediaType[]
    required: boolean
}

declare interface HttpResponse {
    statusCode: string
    description: string
    contents: HttpMediaType[]
    headers: HttpParameter[]
}

declare interface HttpMediaType {
    type: string
    schema: Schema
    example: any
}

declare interface HttpEventParameter {
    name: string
    type: string
    value: string
    raw: string
}

declare interface HttpEventResponse {
    statusCode: number
    body: string
    size: number
    headers: HttpHeader
}

declare interface HttpHeader  {[name: string]: string}

declare interface HttpServer {
    url: string
    description: string
}

declare interface Action {
    duration: number
    tags: Tags
    parameters: any[]
    logs: { level: string, message: string}[]
    error?: { message: string }

}

declare interface Tags { [key: string]: string }

declare interface HttpSecurity {
    [ name: string ]: HttpSecurityScheme
}

declare interface HttpSecurityScheme {
    scopes: string[]
    configs: { [ name: string ]: any }
}


