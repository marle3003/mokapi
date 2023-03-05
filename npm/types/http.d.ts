declare module 'mustache' {
    function get(url: string, args: Args): Response
    function post(url: string, body: string, args: Args): Response
    function put(url: string, body: string, args: Args): Response
    function head(url: string, args: Args): Response
    function patch(url: string, args: Args): Response
    function del(url: string, body: string, args: Args): Response
    function options(url: string, body: string, args: Args): Response
}

type Header = { [name: string]: string };

declare interface Args {
    header?: Header;
}

declare interface Response {
    body: string
    statusCode: number
    headers: Header 
}