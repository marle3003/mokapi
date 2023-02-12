declare interface HttpService extends Service {
    paths: HttpPath[]
}

declare interface HttpPath {
    path: string
    summary: string
    description: string
    operations: HttpOperation[]
}

declare interface HttpOperation {
    method: string
}

declare interface HttpEventData {
    request: HttpEventRequest
    response: HttpEventResponse
    duration: number
}

declare interface HttpEventRequest {
    method: string
    url: string
    contentType: string
    parameters: HttpEventParameter[]
    body: string
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
