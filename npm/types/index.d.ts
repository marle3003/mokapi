/**
 * Mokapi JavaScript API
 * https://mokapi.io/docs/guides/welcome
 */

import "./faker";
import "./global";
import "./http";
import "./mustache";
import "./yaml";
import "./encoding";
import "./mail";

/**
 * Attaches an event handler for the given event.
 * https://mokapi.io/docs/javascript-api/mokapi/on
 * @param event Event type such as http
 * @param handler An EventHandler to execute when the event is triggered
 * @param args EventArgs object contains additional event arguments.
 * @example
 * export default function() {
 *   on('http', function(request, response) {
 *     if (request.url.path === '/echo') {
 *       response.data = {
 *         url: request.url.toString(),
 *         body: request.body,
 *       }
 *     }
 *   })
 * }
 */
export function on<T extends keyof EventHandler>(event: T, handler: EventHandler[T], args?: EventArgs): void;

/**
 * Schedules a new periodic job with interval.
 * @param interval interval - Scheduled interval.
 * @param f f - Scheduled event handler
 * @param args args - Additional arguments
 * @example
 * export default function() {
 *   every('1m', function() {
 *     console.log('foo')
 *   })
 * }
 */
export function every(interval: Interval, f: ScheduledEventHandler, args?: ScheduledEventArgs): void;

/**
 * Schedules a new periodic job with a cron expression.
 * @param expr expr - Cron expression
 * @param f f - cheduled event handler
 * @param args args - Additional arguments
 * @example
 * export default function() {
 *   cron('* * * * *', function() {
 *     console.log('foo')
 *   })
 * }
 */
export function cron(expr: string, f: ScheduledEventHandler, args?: ScheduledEventArgs): void;

/**
 * Retrieves the value of the environment variable named by the key.
 * @param name name - The name of the environment variable.
 * @returns The value of the environment variable specified by variable, or an empty string
 * if the environment variable is not found.
 * @example
 * export default function() {
 *   console.log(env('foo'))
 * }
 */
export function env(name: string): string;

/**
 * Returns a textual representation of the date.
 * https://mokapi.io/docs/javascript-api/mokapi/date
 * @description Default layout is RFC3339. Default timestamp is current UTC
 * @param args args - Adjusting textual representation
 * @example
 * export default function() {
 *   console.log(date())
 *   console.log(date({layout: 'UnixDate'}))
 *   console.log(date({timestamp: new Date().getTime()}))
 * }
 */
export function date(args?: DateArgs): string;

/**
 * Suspends the execution for the specified duration.
 * https://mokapi.io/docs/javascript-api/mokapi/sleep
 * @param time Duration in milliseconds or duration as string with unit.
 * Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`
 */
export function sleep(time: number | string): void;

/**
 * Specifies the interval of a periodic job.
 * Interval string is a possibly signed sequence of decimal numbers, each with an optional
 * fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
 * Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
 */
export type Interval = string;

export interface EventHandler {
    http: HttpEventHandler;
    kafka: KafkaEventHandler;
    ldap: LdapEventHandler;
    smtp: SmtpEventHandler;
}

/**
 * HttpEventHandler is a function that is executed when an HTTP event is triggered.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httpeventhandler
 * @example
 * export default function() {
 *   on('http', function(request, response) {
 *     if (request.operationId === 'time') {
 *       response.body = date()
 *       return true
 *     }
 *     return false
 *   })
 * }
 */
export type HttpEventHandler = (request: HttpRequest, response: HttpResponse) => void;

/**
 * HttpRequest is an object used by HttpEventHandler that contains request-specific
 * data such as HTTP headers.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httprequest
 */
export interface HttpRequest {
    /** Request method. */
    readonly method: string;

    /** Represents a parsed URL. */
    readonly url: Url;

    /** Body contains the request body specified by OpenAPI request body. */
    readonly body: any;

    /** Object contains path parameters specified by OpenAPI path parameters. */
    readonly path: { [key: string]: any };

    /** Object contains query parameters specified by OpenAPI query parameters. */
    readonly query: { [key: string]: any };

    /** Object contains header parameters specified by OpenAPI header parameters. */
    readonly header: { [key: string]: any };

    /** Object contains cookie parameters specified by OpenAPI cookie parameters. */
    readonly cookie: { [key: string]: any };

    /** Path value specified by the OpenAPI path */
    readonly key: string;

    /** OperationId defined in OpenAPI */
    readonly operationId: string;
}

/**
 * HttpResponse is an object used by HttpEventHandler that contains response-specific data such as HTTP headers.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httpresponse
 */
export interface HttpResponse {
    /** Object contains header parameters specified by OpenAPI header parameters. */
    headers: { [key: string]: string };

    /** Specifies the http status used to select the OpenAPI response definition. */
    statusCode: number;

    /** Response body. It has a higher precedence than data. */
    body: string;

    /** Data will be encoded with the OpenAPI response definition. */
    data: any;
}

/**
 * Represents an URL
 */
export interface Url {
    /** URL scheme. */
    readonly scheme: string;

    /** URL host. */
    readonly host: string;

    /** URL path. */
    readonly path: string;

    /** URL query string. */
    readonly query: string;
}

/**
 * KafkaEventHandler is a function that is executed when a Kafka message is received.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventHandler
 * @example
 * export default function() {
 *   on('kafka', function(msg) {
 *     // add header 'foo' to every Kafka message
 *     msg.headers = { foo: 'bar' }
 *   })
 * }
 */
export type KafkaEventHandler = (message: KafkaEventMessage) => void;

/**
 * KafkaEventMessage is an object used by KafkaEventHandler that contains Kafka-specific message data.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventMessage
 */
export interface KafkaEventMessage {
    /** Kafka partition where the message was written to (read-only). */
    readonly offset: number;

    /** Kafka message key  */
    key: string;

    /** Kafka message value */
    value: string;

    /** Kafka message headers */
    headers: { [name: string]: string } | null;
}

/**
 * LdapEventHandler is a function that is executed when a LDAP search query is triggered.
 * @example
 * export default function() {
 *   on('ldap', function(request, response) {
 *     if (request.filter === '(objectClass=foo)') {
 *       response.results = [
 *         {
 *           dn: 'CN=bob,CN=users,DC=mokapi,DC=io',
 *           attributes: {
 *             mail: ['bob@mokapi.io'],
 *             objectClass: ['foo']
 *           }
 *         }
 *       ]
 *     }
 *   })
 * }
 */
export type LdapEventHandler = (request: LdapSearchRequest, response: LdapSearchResponse) => void;

/**
 * LdapSearchRequest is an object used by LdapEventHandler that contains request-specific data.
 */
export interface LdapSearchRequest {
    /** Search base DN. */
    baseDN: string;

    /** Search scope. */
    scope: LdapSearchScope;

    /** Alias dereference policy. */
    dereferencePolicy: number;

    /** Maximum number of entries to return from the search. */
    sizeLimit: number;

    /** Maximum length of time in seconds to allow for the search. */
    timeLimit: number;

    /** Only retrieve attribute names but not their values. */
    typesOnly: number;

    /** String representation of an LDAP search filter. */
    filter: string;

    /** Attribute list specifies the attributes to return in the entries found by the search. */
    attributes: string[];
}

/**
 * LdapSearchResponse is an object used by LdapEventHandler that contains response-specific data.
 */
export interface LdapSearchResponse {
    /** List of search result */
    results: LdapSearchResult[];

    /** Status of search operation */
    status: LdapResultStatus;

    /** Search response message */
    message: string;
}

/**
 * LdapSearchResult is an object used by LdapSearchResponse that contains one result of a search request.
 */
export interface LdapSearchResult {
    /** LDAP distinguished name of this result. */
    dn: string;

    /** Attribute list of this result */
    attributes: { [name: string]: string[] };
}

/**
 * Specifies the portion of the target subtree that should be considered.
 */
export enum LdapSearchScope {
    /**
     * Indicates that only the entry specified as sthe search base should be considered.
     * None of its subordinates will be considered.
     */
    BaseObject,

    /**
     * Indicates that only the immediate children of the entry specified should be considered.
     */
    SingleLevel,

    /**
     * Indicates that the entry specified as the search base, and all of its subordinates to any depth.
     */
    WholeSubtree,
}

/**
 * Defines a number of result codes that are intended to be used in LdapSearchResponse.
 */
export enum LdapResultStatus {
    /** The success result code is used to indicate that the associated operation completed successfully. */
    Success = 0,

    /** Indicates that the operation could not be processed because it wasn’t in the expected
     * order relative to other operations on the same connection.
     */
    OperationsError = 1,

    /** Indicates that there was a problem with the client’s use of the LDAP protocol. */
    ProtocolError = 2,

    /**
     *  indicates that the associated search operation failed because the server has determined
     *  that the number of entries that would be returned in response to the search would exceed
     *  the upper bound for that operation.
     */
    SizeLimitExceeded = 4,
}

export type SmtpEventHandler = (record: SmtpEventMessage) => void;

export interface SmtpEventMessage {
    server: string;
    sender?: Address;
    from: Address[];
    to: Address[];
    replyTo?: Address[];
    cc?: Address[];
    bcc?: Address[];
    messageId: string;
    inReplyTo?: string;
    time?: Date;
    subject: string;
    contentType: string;
    encoding: string;
    body: string;
    attachments: Attachment[];
}

export interface Address {
    name?: string;
    address: string;
}

export interface Attachment {
    name: string;
    contentType: string;
    data: Uint8Array;
}

/**
 * Contains optional arguments to format a datetime object
 */
export interface DateArgs {
    /**
     * The format of the textual representation, default is RFC3339
     */
    layout?: DateLayout | string;

    /**
     * The timestamp of the date, default is current UTC time
     */
    timestamp?: number;
}

/**
 * These are predefined layouts for use in date()
 */
export type DateLayout =
    | "DateTime"
    | "DateOnly"
    | "TimeOnly"
    | "UnixDate"
    | "RFC882"
    | "RFC822Z"
    | "RFC850"
    | "RFC1123"
    | "RFC1123Z"
    | "RFC3339"
    | "RFC3339Nano";

/**
 * EventArgs object contains additional arguments for an event handler.
 * https://mokapi.io/docs/javascript-api/mokapi/on
 *
 * Use this to customize how the event appears in the dashboard or to control tracking.
 *
 * @property tags Optional key-value pairs used to label the event in the dashboard.
 *
 * @property track Set to `true` to enable tracking of this event handler in the dashboard.
 *                 If omitted, Mokapi automatically checks whether the response object has
 *                 been modified and tracks the handler only if a change is detected.
 *
 * @example
 * export default function() {
 *   on('http', function(req, res) {
 *     res.data = { message: "tracked event" }
 *   }, {
 *     tags: { feature: 'beta', owner: 'team-a' },
 *     track: true
 *   })
 * }
 */
export interface EventArgs {
    /**
     * Adds or overrides existing tags used in dashboard
     */
    tags?: { [key: string]: string };

    /**
     * Set to `true` to enable tracking of this event handler in the dashboard.
     * Set to `false` to disable tracking. If omitted, Mokapi checks the response
     * object to determine if the handler changed it, and tracks it accordingly.
     */
    track?: boolean;
}

/**
 * cheduledEventArgs is an object used by every and cron function.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/scheduledeventargs
 * @example
 * export default function() {
 *   every('1m', function() {
 *     console.log('foo')
 *   }, {times: 1, runFirstTimeImmediately: false})
 * }
 */
export type ScheduledEventHandler = () => void;

export interface ScheduledEventArgs {
    /**
     * Adds or overrides existing tags used in dashboard
     */
    tags?: { [key: string]: string };

    /**
     * Defines the number of times the scheduled function is executed.
     */
    times?: number;

    /**
     * Toggles behavior of first execution. Default is true
     */
    runFirstTimeImmediately?: boolean;
}

/**
 * JavaScript value representable with JSON.
 */
export type JSONValue = null | undefined | boolean | number | string | JSONValue[] | JSONObject;

/**
 * Object representable with JSON.
 */
export interface JSONObject {
    [key: string]: JSONValue;
}

/**
 * Specifies the date-time format defined in [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339).
 * This constant can be used when defining or validating datetime strings.
 *
 * @example
 * const date = new Date().toISOString()
 * if (isValidDate(date, RFC3339)) {
 *   // do something
 * }
 */
export const RFC3339 = "RFC3339";

/**
 * Applies a patch object to a target object. Only properties that are explicitly defined in the patch
 * are applied. This includes nested objects. Properties marked with `Delete` will be removed.
 *
 * This function is especially useful when working with generated mock data in Mokapi that you want to override
 * or refine with specific values.
 *
 * https://mokapi.io/docs/javascript-api/mokapi/patch
 *
 * @param target The original object or value to be patched.
 * @param patch The patch object or value. Only defined values are applied; undefined values are ignored. Use `Delete` to remove fields.
 * @returns A new object or value with the patch applied.
 *
 * @example
 * const result = patch({ name: "foo", age: 42 }, { name: "bar" })
 * // result: { name: "bar", age: 42 }
 *
 * @example
 * const result = patch({ name: "foo", meta: { version: 1 } }, { meta: { version: 2 } })
 * // result: { name: "foo", meta: { version: 2 } }
 *
 * @example
 * const result = patch({ name: "foo", age: 42 }, { age: Delete })
 * // result: { name: "foo" }
 */
export function patch(target: any, patch: any): any;

/**
 * Special marker used with the `patch` function to indicate a property should be removed.
 *
 * When used as a value inside a patch object, the corresponding property will be deleted
 * from the result.
 *
 * This is useful when refining or overriding mock data in a script while keeping validation logic intact.
 *
 * https://mokapi.io/docs/javascript-api/mokapi/patch#delete
 *
 * @example
 * const result = patch({ name: "foo", age: 42 }, { age: Delete })
 * // result: { name: "foo" }
 */
export const Delete: unique symbol;