declare module 'mokapi' {

    /** Listener for http events. */
    function on(event: 'http', f: HttpEventHandler, args: EventArgs): void

    /** Schedules a new periodic job with interval.
     * Interval string is a possibly signed sequence of
     * decimal numbers, each with optional fraction and a unit suffix,
     * such as "300ms", "-1.5h" or "2h45m".
     * Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
     */
    function every(interval: string, f: () => void, args: ScheduleArgs): void

    /**
     * Schedules a new periodic job with a cron expression.
     * @param expr cron expression
     * @param f function to execute
     * @param args additional arguments
     */
    function cron(expr: string, f: () => void, args: ScheduleArgs): void

    /** Returns the environment variable named by the key. */
    function env(key: string): string

    /** Opens a file and reading all its content. */
    function open(path: string): string

    /** Returns a textual representation of the date. See the documentation to see
     *  how to represent the layout format.
     *  Default layout is RFC3339.
     *  Default timestamp is current UTC */
    function date(args?: DateArgs): string
}

type HttpEventHandler = (request: HttpRequest, response: HttpResponse) => boolean

declare interface HttpRequest {
    /** Request method. */
    method: string
    /** Request url. */
    url: Url
    /** Request body. */
    body: any
    /** Path parameters. */
    path: { [key: string]: any; }
    /** Query parameters. */
    query: { [key: string]: any; }
    /** Header parameters. */
    header: { [key: string]: any; }
    /** Cookie parameters. */
    cookie: { [key: string]: any; }
    /** Path defined in OpenAPI. */
    key: string
    /** OperationId definied in OpenAPI. */
    operationId: string
}

declare interface HttpResponse {
    /** Response headers. */
    headers: { [key: string]: string; }
    /** HTTP status code. */
    statusCode: number
    /** Response body. It has a higher precedence than data. */
    body: string
    /** Response data. It has a lower precedence than body. */
    data: any
}

declare interface Url {
    /** URL scheme. */
    scheme: string
    /** URL host. */
    host: string
    /** URL path. */
    path: string
    /** URL query string. */
    query: string
}

declare interface DateArgs {
    layout?: string
    timestamp?: number
}

declare interface EventArgs {
    /**
     * Adds or overrides existing tags used in dashboard
     */
    tags: {[key: string]: string}
}

declare interface ScheduleArgs {
    /**
     * Adds or overrides existing tags used in dashboard
     */
    tags: {[key: string]: string}

    /**
     * Defines the number of times the scheduled function is executed.
     */
    times: number
}

declare const RFC3339 = "RFC3339"