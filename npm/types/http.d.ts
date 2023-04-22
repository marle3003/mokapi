declare module 'mokapi/http' {
    function get(url: string, args?: Args): Response
    function post(url: string, body?: any, args?: Args): Response
    function put(url: string, body?: any, args?: Args): Response
    function head(url: string, args?: Args): Response
    function patch(url: string, args?: Args): Response
    function del(url: string, body?: any, args?: Args): Response
    function options(url: string, body?: any, args?: Args): Response
}

declare interface Args {
    header?: { [name: string]: any };
}

declare interface Response {
    body: string
    statusCode: number
    headers: { [name: string]: string[] }
    json(): any
}