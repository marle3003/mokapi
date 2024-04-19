/// <reference path="ldap.d.ts" />
/// <reference path="kafka.d.ts" />
/// <reference path="smtp.d.ts" />

declare module 'mokapi' {
    /**
     * Attaches an event handler for the given event.
     * @param event Event type such as http
     * @param handler An EventHandler to execute when the event is triggered
     * @param args EventArgs object contains additional event arguments.
     */
    function on<T extends keyof EventHandler>(event: T, handler: EventHandler[T], args?: EventArgs): void

    /** Schedules a new periodic job with interval.
     * Interval string is a possibly signed sequence of
     * decimal numbers, each with optional fraction and a unit suffix,
     * such as "300ms", "-1.5h" or "2h45m".
     * Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
     */
    function every(interval: string, f: ScheduledEventHandler, args?: ScheduledEventArgs): void

    /**
     * Schedules a new periodic job with a cron expression.
     * @param expr cron expression
     * @param f function to execute
     * @param args additional arguments
     */
    function cron(expr: string, f: ScheduledEventHandler, args?: ScheduledEventArgs): void

    /** Retrieves the value of the environment variable named by the key.
     It returns the value, which will be empty if the variable is not present.
     */
    function env(name: string): string

    /** Opens a file and reading all its content. */
    function open(path: string): string

    /** Returns a textual representation of the date. See the documentation to see
     *  how to represent the layout format.
     *  Default layout is RFC3339.
     *  Default timestamp is current UTC */
    function date(args?: DateArgs): string

    /**
     * Suspends the execution for the specified duration.
     * Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`
     * @param time Duration in milliseconds or duration as string with unit.
     */
    function sleep(time: number | string )
}

type EventHandler = {
    http: HttpEventHandler
    ldap: LdapEventHandler
    kafka: KafkaEventHandler
    smtp: SmtpEventHandler
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
    /** OperationId defined in OpenAPI. */
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
    layout?: DateLayout
    timestamp?: number
}

declare type DateLayout = 'DateTime' | 'DateOnly' | 'TimeOnly' | 'UnixDate' | 'RFC882' | 'RFC822Z' | 'RFC850' | 'RFC1123' | 'RFC1123Z' | 'RFC3339' | 'RFC3339Nano'

declare interface EventArgs {
    /**
     * Adds or overrides existing tags used in dashboard
     */
    tags: {[key: string]: string}
}

type ScheduledEventHandler = () => void

declare interface ScheduledEventArgs {
    /**
     * Adds or overrides existing tags used in dashboard
     */
    tags?: {[key: string]: string}

    /**
     * Defines the number of times the scheduled function is executed.
     */
    times?: number

    /**
     * Toggles behavior of first execution. If true job does not start
     * immediately but rather wait until the first scheduled interval.
     */
    skipImmediateFirstRun?: boolean
}

declare const RFC3339 = "RFC3339"