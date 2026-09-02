/**
 * Mokapi JavaScript API
 *
 * This module exposes the core scripting API for Mokapi.
 * It allows you to intercept and manipulate protocol events (HTTP, Kafka, LDAP, SMTP),
 * schedule jobs, generate mock data, and share state between scripts.
 *
 * Documentation:
 * https://mokapi.io/docs/javascript-api/overview
 */

import "./faker";
import "./global";
import "./http";
import "./kafka";
import "./mqtt";
import "./mustache";
import "./yaml";
import "./encoding";
import "./mail";
import "./file";

/**
 * Attaches an event handler for the given event.
 *
 * Event handlers are executed in priority order whenever the event occurs.
 * Multiple handlers can be registered for the same event.
 *
 * https://mokapi.io/docs/javascript-api/mokapi/on
 * @param event Event type such as `http`, `kafka`, `ldap`, or `smtp`
 * @param handler Function executed when the event is triggered
 * @param args Optional event configuration such as priority, tracking, or tags
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
export function on<T extends keyof EventHandler>(event: T, handler: EventHandler[T], args?: TypedEventArgs[T]): void;

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
    mqtt: MqttEventHandler;
    ldap: LdapEventHandler;
    smtp: SmtpEventHandler;
    websocket: WebsocketEventHandler;
}

/**
 * HttpEventHandler is invoked for every incoming HTTP request.
 *
 * Handlers may modify the response object to influence the outgoing response.
 * The return value is ignored.
 *
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httpeventhandler
 * @example
 * export default function() {
 *   on('http', function(request, response) {
 *     if (request.operationId === 'time') {
 *       response.body = date()
 *     }
 *   })
 * }
 */
export type HttpEventHandler = (request: HttpRequest, response: HttpResponse) => void | Promise<void>;

/**
 * HttpRequest is an object used by HttpEventHandler that contains request-specific
 * data such as HTTP headers.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httprequest
 */
export interface HttpRequest {
    /**
     * Request method.
     * @example GET
     */
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

    /** Object contains querystring parameters specified by OpenAPI querystring parameters. */
    readonly querystring: any;

    /** Name of the API, as defined in the OpenAPI `info.title` field */
    readonly api: string;

    /** Path value specified by the OpenAPI path */
    readonly key: string;

    /** OperationId defined in OpenAPI */
    readonly operationId: string;

    /** Returns a string representing this HttpRequest object.  */
    toString(): string;
}

/**
 * HttpResponse is an object used by HttpEventHandler that contains response-specific data such as HTTP headers.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/httpresponse
 */
export interface HttpResponse {
    /** Object contains header parameters specified by OpenAPI header parameters. */
    headers: { [key: string]: any };

    /** Specifies the http status used to select the OpenAPI response definition. */
    statusCode: number;

    /** Response body. It has a higher precedence than data. */
    body: string;

    /** Data will be encoded with the OpenAPI response definition. */
    data: any;

    /**
     * Rebuilds the entire HTTP response using the OpenAPI response definition.
     *
     * This resets the status code, headers, and response body/data
     * based on the OpenAPI specification.
     *
     * - If `statusCode` is omitted, the OpenAPI `default` response is used.
     * - If `contentType` is omitted, the first defined content type for the
     *   selected status code is used.
     *
     * Use this when switching to a different response (e.g. error handling)
     * while keeping the response schema valid.
     *
     * @throws Error if the status code or content type is not defined in the OpenAPI spec
     *
     * @example
     * on('http', (request, response) => {
     *   if (request.path.petId === 10) {
     *     response.rebuild(404)
     *     response.data.message = 'Pet not found'
     *   }
     * })
     */
    rebuild(statusCode?: number, contentType?: string): void;

    /**
     * A request-scoped key-value store for passing data between handlers
     * registered on the same route.
     *
     * Use `context` to share state between `use()` middleware and method handlers
     * without relying on external variables. Values are only available for the
     * lifetime of the current request.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   app.api('Petstore').http()
     *     .route('/pets/{petId}')
     *       .use((req, res) => {
     *         res.context.user = parseToken(req.header['Authorization'])
     *       })
     *       .get((req, res) => {
     *         res.data = { id: req.path.petId, owner: res.context.user.name }
     *       })
     * }
     */
    context: Record<string, any>

    /**
     * Stops the handler pipeline for the current route.
     *
     * By default, all handlers registered on a route run in order.
     * Calling `stopPropagation()` prevents any subsequent handlers
     * on the same route from being executed.
     *
     * This is useful for guard logic such as authentication checks,
     * where further processing should be skipped if a condition is not met.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   app.api('Petstore').http()
     *     .route('/pets/{petId}')
     *       .use((req, res) => {
     *         if (!req.header['Authorization']) {
     *           res.statusCode = 401
     *           res.stopPropagation()  // get/delete will not run
     *           return
     *         }
     *         res.context.user = parseToken(req.header['Authorization'])
     *       })
     *       .get((req, res) => {
     *         res.data = { id: req.path.petId, owner: res.context.user.name }
     *       })
     *       .delete((req, res) => {
     *         res.statusCode = 204
     *       })
     * }
     */
    stopPropagation(): void
}

/**
 * Represents an URL
 */
export interface Url {
    /** URL scheme. */
    readonly scheme: string;

    /** URL host. */
    readonly host: string;

    /** URL port */
    readonly port: number;

    /** URL path. */
    readonly path: string;

    /** URL query string. */
    readonly query: string;

    /** Returns a string representing this Url object.  */
    toString(): string;
}

/**
 * KafkaEventHandler is a function that is executed when a Kafka message is received.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventHandler
 * @example
 * export default function() {
 *   on('kafka', function(msg) {
 *     // add header 'foo' to every Kafka message
 *     message.headers['correlationId'] = 'abc-123'
 *   })
 * }
 */
export type KafkaEventHandler = (message: KafkaEventMessage) => void | Promise<void>;

/**
 * KafkaEventMessage is an object used by KafkaEventHandler that contains Kafka-specific message data.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventMessage
 */
export interface KafkaEventMessage {
    /** The name of the API, matching the `info.title` field in the AsyncAPI specification */
    readonly api: string;

    /**
     * The name of the Kafka topic where this message originates.
     */
    readonly topic: string;

    /**
     * The partition index within the Kafka topic.
     */
    readonly partition: string;

    /**
     * The unique sequential offset assigned to the message within its partition.
     */
    readonly offset: number;

    /**
     * The key of the Kafka message, used for partitioning and identification.
     */
    key: string;

    /** The Kafka message value */
    value: string;

    /** Key-value pairs representing optional Kafka message headers. */
    headers: { [name: string]: string } | null;

    /**
     * The identifier of the schema used to validate or serialize this message payload.
     *
     * @remarks
     * * This property is **only set** automatically if the schema ID is embedded within the
     *   message payload itself (matching `SchemaIdLocation: "payload"`).
     * * If the ID is transmitted via message metadata instead (matching `SchemaIdLocation: "header"`),
     *   it will **not** populate this field and must instead be extracted manually from the {@link headers} object.
     */
    schemaId: number;
}

export type MqttEventHandler = (message: MqttEventMessage) => void | Promise<void>;

/**
 * KafkaEventMessage is an object used by KafkaEventHandler that contains Kafka-specific message data.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/KafkaEventMessage
 */
export interface MqttEventMessage {
    /** The name of the API, matching the `info.title` field in the AsyncAPI specification */
    readonly api: string;

    /**
     * The name of the MQTT topic where this message originates.
     */
    readonly topic: string;

    /** Specifies whether message should be retained  */
    retain: boolean;

    /** MQTT message value */
    value: string;
}

/**
 * Event handler function passed to `on('websocket', handler)`. Called for every
 * WebSocket lifecycle event — connect, message, and close — on any channel
 * defined in the AsyncAPI specification.
 *
 * Use the `event.type` field to distinguish between event types and access
 * type-specific fields like `message` and `reply`.
 *
 * Async handlers are supported — return a `Promise` if you need to perform
 * asynchronous work before responding.
 *
 * @example
 * // Handle all event types in one handler
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.type === 'connect') {
 *       event.client.send({ text: 'welcome!' })
 *     }
 *     if (event.type === 'message') {
 *       event.reply({ text: 'pong' })
 *     }
 *     if (event.type === 'close') {
 *       console.log(`connection closed by ${event.closedBy}: ${event.reason}`)
 *     }
 *   })
 * }
 *
 * @example
 * // Async handler
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', async function(event) {
 *     if (event.type === 'message') {
 *       await someAsyncOperation()
 *       event.reply({ text: 'done' })
 *     }
 *   })
 * }
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketEventHandler
 */
export type WebsocketEventHandler = (message: WebsocketEvent) => void | Promise<void>;

export type WebsocketEvent = WebsocketEventConnect | WebsocketEventMessage | WebsocketEventClose;

/**
 * Base interface shared by all WebSocket events. Use the `type` field to
 * discriminate between connect, message, and close events.
 *
 * @example
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.type === 'connect') {
 *       event.broadcast({ text: 'a new user joined' })
 *     }
 *     if (event.type === 'message') {
 *       event.reply({ text: 'pong' })
 *     }
 *     if (event.type === 'close') {
 *       console.log(`client disconnected: ${event.reason}`)
 *     }
 *   })
 * }
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketEvent
 */
export interface WebsocketEventBase {
    /**
     * Discriminates between the three WebSocket event types.
     * Use this in a conditional to access type-specific fields like
     * `message` and `reply` (only on `'message'`) or `reason` (only on `'close'`).
     */
    readonly type: 'connect' | 'message' | 'close'

    /**
     * The name of the API, matching the `info.title` field in the AsyncAPI specification.
     *
     * @example
     * on('websocket', function(event) {
     *   console.log(event.api) // e.g. "Chat API"
     * })
     */
    readonly api: string;

    /**
     * The channel on which the message was received.
     *
     * @example
     * on('websocket', function(event) {
     *   console.log(event.channel.name) // e.g. "/chat"
     * })
     */
    readonly channel: WebsocketChannel;

    /**
     * The client that sent the message. Can be stored and used later
     * to send messages outside the current event handler.
     *
     * @example
     * on('websocket', function(event) {
     *   const userId = event.client.query['userId']
     *   const token  = event.client.headers['authorization']
     * })
     */
    readonly client: WebsocketClient;

    /**
     * Sends a message to all clients currently connected to this channel,
     * including the sender. Shorthand for looping over `event.channel.clients`.
     *
     * @param message - The message payload to send to every connected client.
     *
     * @example
     * // Broadcast a chat message to everyone
     * on('websocket', function(event) {
     *   event.broadcast({ from: event.client.query['userId'], text: event.message.text })
     * })
     *
     * @example
     * // Broadcast excluding the sender
     * on('websocket', function(event) {
     *   for (const client of event.channel.clients) {
     *     if (client.remoteAddr !== event.client.remoteAddr) {
     *       client.send(event.message)
     *     }
     *   }
     * })
     */
    broadcast(message: any): void;
}

/**
 * WebsocketEventMessage is an object passed to a WebSocket event handler whenever
 * a message is received from a client. It provides access to the message payload,
 * connection metadata, and methods to send messages back.
 *
 * @example
 * // Simple request/reply
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     event.reply({ text: 'pong' })
 *   })
 * }
 *
 * @example
 * // Broadcast to all connected clients (e.g. chat)
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     event.broadcast({ from: event.client.query['userId'], text: event.message.text })
 *   })
 * }
 *
 * @example
 * // Store clients and send later
 * import { on } from 'mokapi'
 * const clients: WebsocketClient[] = []
 * export default function() {
 *   on('websocket', function(event) {
 *     clients.push(event.client)
 *   })
 * }
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketEventMessage
 */
export interface WebsocketEventMessage extends WebsocketEventBase {
    readonly type: 'message'

    /**
     * The decoded message payload. The type depends on the `contentType`
     * defined in the AsyncAPI specification — typically an object for
     * `application/json` or a string for `text/plain`.
     *
     * @example
     * on('websocket', function(event) {
     *   console.log(event.message.text)
     * })
     */
    readonly message: any;

    /**
     * Sends a message back to the client that sent the current message.
     * Shorthand for `event.client.send(message)`.
     *
     * The message is validated against the AsyncAPI specification before
     * being sent. If validation fails, the error is recorded in the
     * dashboard event log.
     *
     * @param message - The message payload to send. Will be encoded
     *   according to the channel's `contentType`.
     *
     * @example
     * on('websocket', function(event) {
     *   event.reply({ text: 'got your message' })
     * })
     */
    reply(message: any): void;
}

/**
 * Fired once when a client establishes a WebSocket connection.
 * Use this to send a welcome message to the connecting client or
 * notify all other clients that someone joined.
 *
 * @example
 * // Send a welcome message to the connecting client
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.type === 'connect') {
 *       event.client.send({ text: 'welcome!' })
 *     }
 *   })
 * }
 *
 * @example
 * // Notify all existing clients that someone joined
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.type === 'connect') {
 *       event.broadcast({ text: `${event.client.query['userId']} has joined` })
 *     }
 *   })
 * }
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketEventConnect
 */
export interface WebsocketEventConnect extends WebsocketEventBase {
    readonly type: 'connect'
}

/**
 * Fired when a WebSocket connection is closed, either by the client or the server.
 * Use `closedBy` to determine who initiated the close and `reason` for the explanation.
 *
 * @example
 * // Log when a client disconnects
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.type === 'close') {
 *       console.log(`connection closed by ${event.closedBy}: ${event.reason}`)
 *     }
 *   })
 * }
 *
 * @example
 * // Notify remaining clients when someone leaves
 * import { on } from 'mokapi'
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.type === 'close') {
 *       event.broadcast({ text: `${event.client.query['userId']} has left` })
 *     }
 *   })
 * }
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketEventClose
 */
export interface WebsocketEventClose extends WebsocketEventBase {
    readonly type: 'close'

    /**
     * The reason the connection was closed. Maps to the WebSocket close frame
     * reason string. Empty if no reason was provided.
     *
     * @example
     * on('websocket', function(event) {
     *   if (event.type === 'close') {
     *     console.log(event.reason) // e.g. "going away"
     *   }
     * })
     */
    readonly reason: string

    /**
     * Indicates who initiated the close — either the client or the server.
     *
     * @example
     * on('websocket', function(event) {
     *   if (event.type === 'close') {
     *     if (event.closedBy === 'client') {
     *       console.log('client disconnected')
     *     } else {
     *       console.log('server closed the connection')
     *     }
     *   }
     * })
     */
    readonly closedBy: 'client' | 'server'
}

/**
 * Represents a WebSocket client connected to a channel.
 * The client object is populated from the HTTP upgrade handshake,
 * so `query` and `headers` reflect the values sent during the
 * initial connection request.
 *
 * Client references to remain valid after the event handler returns,
 * so they can be stored and used to push messages at any time
 * while the connection is open.
 *
 * @example
 * // Store clients on connect, send later from another event
 * import { on } from 'mokapi'
 * const clients: WebsocketClient[] = []
 * export default function() {
 *   on('websocket', function(event) {
 *     if (event.message.type === 'join') {
 *       clients.push(event.client)
 *     }
 *   })
 * }
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketClient
 */
export interface WebsocketClient {
    /**
     * The remote address of the client in `host:port` format.
     * Can be used to distinguish clients when no application-level
     * identity is available.
     *
     * @example
     * on('websocket', function(event) {
     *   console.log(event.client.remoteAddr) // e.g. "127.0.0.1:54321"
     * })
     */
    readonly remoteAddr: string;

    /**
     * Query parameters from the HTTP upgrade request, as defined
     * in the channel's WebSocket binding (`bindings.ws.query`).
     *
     * @example
     * // AsyncAPI binding:
     * // bindings:
     * //   ws:
     * //     query:
     * //       properties:
     * //         userId:
     * //           type: string
     *
     * on('websocket', function(event) {
     *   const userId = event.client.query['userId']
     * })
     */
    readonly query: Record<string, string>;

    /**
     * HTTP headers from the upgrade request, as defined in the
     * channel's WebSocket binding (`bindings.ws.headers`).
     *
     * @example
     * // AsyncAPI binding:
     * // bindings:
     * //   ws:
     * //     headers:
     * //       properties:
     * //         Authorization:
     * //           type: string
     *
     * on('websocket', function(event) {
     *   const token = event.client.headers['authorization']
     * })
     */
    readonly headers: Record<string, string>;

    /**
     * Sends a message to this specific client. The client reference
     * can be stored and called outside of the event handler to push
     * messages at any time while the connection is open.
     *
     * @param message - The message payload to send. Will be encoded
     *   according to the channel's `contentType`.
     *
     * @example
     * // Immediate send (equivalent to event.reply)
     * on('websocket', function(event) {
     *   event.client.send({ text: 'hello' })
     * })
     *
     * @example
     * // Deferred send after storing the client
     * const clients: WebsocketClient[] = []
     * on('websocket', function(event) {
     *   clients.push(event.client)
     * })
     * // later, from a timer or another event:
     * clients.forEach(c => c.send({ text: 'server push' }))
     */
    send(message: any): void;
}

/**
 * Represents the WebSocket channel on which a message was received.
 * Provides access to all currently connected clients, which is useful
 * for targeted sends or building broadcast logic manually.
 *
 * @see https://mokapi.io/docs/javascript-api/mokapi/eventhandler/WebsocketChannel
 */
export interface WebsocketChannel {
    /**
     * The channel path, matching the channel address in the AsyncAPI
     * specification.
     *
     * @example
     * on('websocket', function(event) {
     *   console.log(event.channel.name) // e.g. "/chat"
     * })
     */
    readonly name: string;

    /**
     * The parameter values extracted from the channel address for this
     * specific channel instance, keyed by parameter name.
     *
     * @example
     * // channel address: /doc/{roomId}
     * on('websocket', function(event) {
     *   console.log(event.channel.params['roomId']) // e.g. "room-1"
     * })
     */
    readonly params: Record<string, string>;

    /**
     * All clients currently connected to this channel. Use this for
     * custom fan-out logic beyond what `broadcast` provides — for
     * example, filtering by query parameter or sending to a subset.
     *
     * @example
     * // Send to all clients with a specific userId
     * on('websocket', function(event) {
     *   for (const client of event.channel.clients) {
     *     if (client.query['room'] === event.message.room) {
     *       client.send(event.message)
     *     }
     *   }
     * })
     */
    readonly clients: readonly WebsocketClient[];
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
export type LdapEventHandler = (request: LdapSearchRequest, response: LdapSearchResponse) => void | Promise<void>;

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

export type SmtpEventHandler = (record: SmtpEventMessage) => void | Promise<void>;

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
 * EventArgs provides optional configuration for an event handler.
 * https://mokapi.io/docs/javascript-api/mokapi/on
 *
 * Use this object to control how the event is tracked, labeled,
 * and ordered in the execution pipeline.
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
     * Adds or overrides tags used to label this event in the dashboard.
     * Tags can be used for filtering, grouping, or ownership attribution.
     */
    tags?: { [key: string]: string };

    /**
     * Defines the execution order of the event handler.
     *
     * Event handlers are executed in descending priority order.
     * Handlers with the same priority are executed in registration order.
     *
     * Handlers with higher priority values run first.
     * Handlers with lower priority values run later.
     *
     * Use negative priorities (e.g. -1) to run a handler after
     * the response has been fully populated by other handlers,
     * such as for logging or recording purposes.
     */
    priority?: number;
}

/**
 * TypedEventArgs provides strongly typed argument objects
 * for each supported event type.
 *
 * It is mainly used internally to map event names
 * (e.g. `http`, `kafka`) to their corresponding argument types.
 */
export interface TypedEventArgs {
    /**
     * Arguments for HTTP event handlers.
     */
    http: HttpEventArgs;
    /**
     * Arguments for Kafka event handlers.
     */
    kafka: KafkaEventArgs;
    /**
     * Arguments for MQTT event handlers.
     */
    mqtt: MqttEventArgs;

    /**
     * Arguments for Websocket event handlers.
     */
    websocket: WebsocketEventArgs
    /**
     * Arguments for LDAP event handlers.
     */
    ldap: LdapEventArgs;
    /**
     * Arguments for SMTP event handlers.
     */
    smtp: SmtpEventArgs;
}

export interface HttpEventArgs extends EventArgs {
    /**
     * Controls whether this event handler is tracked in the dashboard.
     *
     * - true: always track this handler
     * - false: never track this handler
     * - undefined: Mokapi determines tracking automatically based on
     *   whether the response object was modified by the handler
     */
    track?: boolean | ((request: HttpRequest, response: HttpResponse) => boolean);
}

/**
 * Configuration options for Kafka event handlers.
 *
 * These arguments control execution behavior such as
 * priority, tagging, and dashboard tracking.
 */
export interface KafkaEventArgs extends EventArgs {
    /**
     * Controls whether this event handler is tracked in the dashboard.
     *
     * - true: always track this handler
     * - false: never track this handler
     * - undefined: Mokapi determines tracking automatically based on
     *   whether the message was modified or acknowledged by the handler
     */
    track?: boolean | ((message: KafkaEventMessage) => boolean);
}

/**
 * Configuration options for MQTT event handlers.
 *
 * These arguments control execution behavior such as
 * priority, tagging, and dashboard tracking.
 */
export interface MqttEventArgs extends EventArgs {
    /**
     * Controls whether this event handler is tracked in the dashboard.
     *
     * - true: always track this handler
     * - false: never track this handler
     * - undefined: Mokapi determines tracking automatically based on
     *   whether the message was modified or acknowledged by the handler
     */
    track?: boolean | ((message: MqttEventMessage) => boolean);
}

/**
 * Configuration options for Websocket event handlers.
 *
 * These arguments control execution behavior such as
 * priority, tagging, and dashboard tracking.
 */
export interface WebsocketEventArgs extends EventArgs {
    /**
     * Controls whether this event handler is tracked in the dashboard.
     *
     * - true: always track this handler
     * - false: never track this handler
     * - undefined: Mokapi determines tracking automatically based on
     *   whether the message was modified or acknowledged by the handler
     */
    track?: boolean | ((message: WebsocketEvent) => boolean);
}

/**
 * Configuration options for LDAP event handlers.
 *
 * These arguments control execution behavior such as
 * priority, tagging, and dashboard tracking.
 */
export interface LdapEventArgs extends EventArgs {
    /**
     * Controls whether this event handler is tracked in the dashboard.
     *
     * - true: always track this handler
     * - false: never track this handler
     * - undefined: Mokapi determines tracking automatically based on
     *   whether the response object was modified by the handler
     */
    track?: boolean | ((request: LdapSearchRequest, response: LdapSearchResponse) => boolean);
}

/**
 * Configuration options for SMTP event handlers.
 *
 * These arguments control execution behavior such as
 * priority, tagging, and dashboard tracking.
 */
export interface SmtpEventArgs extends EventArgs {
    /**
     * Controls whether this event handler is tracked in the dashboard.
     *
     * - true: always track this handler
     * - false: never track this handler
     * - undefined: Mokapi determines tracking automatically based on
     *   whether the message was processed or modified by the handler
     */
    track?: boolean | ((record: SmtpEventMessage) => boolean);
}

/**
 * ScheduledEventHandler is an object used by every and cron function.
 * https://mokapi.io/docs/javascript-api/mokapi/eventhandler/scheduledeventargs
 * @example
 * export default function() {
 *   every('1m', function() {
 *     console.log('foo')
 *   }, {times: 1, runFirstTimeImmediately: false})
 * }
 */
export type ScheduledEventHandler = () => void | Promise<void>;

/**
 * Configuration options for scheduled event handlers
 * created via `every` or `cron`.
 */
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

export interface SharedMemory {
    /**
     * Returns the value associated with the given key.
     * @param key The key to retrieve.
     * @returns The stored value, or `undefined` if not found.
     */
    get(key: string): any;

    /**
     * Sets a value for the given key.
     * If the key already exists, its value will be replaced.
     * @param key The key to store the value under.
     * @param value The value to store.
     */
    set(key: string, value: any): void;

    /**
     * Updates a value atomically using an updater function.
     * The current value is passed into the updater function.
     * The returned value is stored and also returned by this method.
     *
     * Example:
     * ```js
     * mokapi.shared.update("requests", count => (count ?? 0) + 1)
     * ```
     *
     * @param key The key to update.
     * @param updater Function that receives the current value and returns the new value.
     * @returns The new value after update.
     */
    update<T = any>(key: string, updater: (value: T | undefined) => T): T;

    /**
     * Checks if the given key exists in shared memory.
     * @param key The key to check.
     * @returns `true` if the key exists, otherwise `false`.
     */
    has(key: string): boolean;

    /**
     * Removes the specified key and its value from shared memory.
     * @param key The key to remove.
     */
    delete(key: string): void;

    /**
     * Removes all stored entries from shared memory.
     * Use with caution — this clears all shared state.
     */
    clear(): void;

    /**
     * Returns a list of all stored keys.
     * @returns An array of key names.
     */
    keys(): string[];

    /**
     * Creates or returns a namespaced shared memory store.
     * Namespaces help avoid key collisions between unrelated scripts.
     *
     * Example:
     * ```js
     * const petstore = mokapi.shared.namespace("petstore")
     * petstore.set("sessions", [])
     * ```
     *
     * @param name The namespace identifier.
     * @returns A `SharedMemory` object scoped to the given namespace.
     */
    namespace(name: string): SharedMemory;
}

/**
 * Shared memory API for Mokapi scripts.
 *
 * The `mokapi.shared` object provides a way to persist and share
 * data between multiple scripts running in the same Mokapi instance.
 *
 * Values are stored in memory and shared across all scripts
 * within the same Mokapi process.
 * This allows you to coordinate state, cache data, or simulate
 * application-level variables without using global variables.
 * All values are persisted for the lifetime of the Mokapi process.
 *
 * Example:
 * ```js
 * // Increment a shared counter
 * mokapi.shared.update("counter", c => (c ?? 0) + 1)
 *
 * // Retrieve the current counter value
 * const count = mokapi.shared.get("counter")
 * console.log(`Current counter: ${count}`)
 * ```
 */
export const shared: SharedMemory;

/**
 * The central application object for registering HTTP (and future protocol) handlers
 * in a structured, route-oriented style.
 *
 * `app` is a singleton — the same instance is returned every time it is accessed.
 * Use `app.api()` to scope handlers to a specific API by its OpenAPI `info.title`,
 * which is required when multiple APIs share the same path.
 *
 * @example
 * // Without API scope — matches all APIs with this path
 * import { app } from 'mokapi'
 * export default function() {
 *   app.http().get('/pets', (req, res) => {
 *     res.data = [{ name: 'Bello' }]
 *   })
 * }
 *
 * @example
 * // With API scope — only matches the 'Petstore' API
 * import { app } from 'mokapi'
 * export default function() {
 *   app.api('Petstore').http().get('/pets', (req, res) => {
 *     res.data = [{ name: 'Bello' }]
 *   })
 * }
 */
export const app: App

/**
 * The top-level application interface providing access to protocol routers
 * and API scoping.
 *
 * Use `api()` to narrow handlers to a specific API by its OpenAPI `info.title`.
 * Use `http()` directly for handlers that apply across all APIs.
 */
export interface App {
    /**
     * Creates or returns a scoped context for the given API title.
     *
     * The title must match the `info.title` field in the OpenAPI specification.
     * Scoping is required when multiple APIs define the same path, to avoid
     * a handler running for unintended APIs.
     *
     * @param title The API title as defined in `info.title` of the OpenAPI spec.
     * @returns An `ApiScope` bound to the given API title.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   const petstore = app.api('Petstore')
     *   petstore.http().get('/pets', (req, res) => {
     *     res.data = [{ name: 'Bello' }]
     *   })
     *
     *   const zoo = app.api('Zoo API')
     *   zoo.http().get('/pets', (req, res) => {
     *     res.data = [{ name: 'Simba' }]
     *   })
     * }
     */
    api(title: string): ApiScope

    /**
     * Returns an `HttpRouter` not bound to any specific API.
     * Handlers registered here will run for all APIs that match
     * the given path and method.
     *
     * Use this for cross-cutting concerns such as logging, shared headers,
     * or common error responses.
     *
     * @returns An `HttpRouter` with no API scope.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   app.http().get('/pets', (req, res) => {
     *     res.headers['X-Mock'] = 'true'
     *   })
     * }
     */
    http(): HttpRouter
}

/**
 * A scoped application context bound to a specific API by its `info.title`.
 *
 * Obtain an `ApiScope` via `app.api('My API')`. All handlers registered
 * through this scope will only run for requests matched to the named API.
 *
 * @example
 * import { app } from 'mokapi'
 * export default function() {
 *   const petstore = app.api('Petstore')
 *   petstore.http().route('/pets')
 *     .get((req, res) => { res.data = [] })
 *     .post((req, res) => { res.statusCode = 201 })
 * }
 */
export interface ApiScope {
    /**
     * Returns an `HttpRouter` scoped to this API.
     *
     * Handlers registered on this router will only run for HTTP requests
     * that are matched to the API identified by this scope's title.
     *
     * @returns An `HttpRouter` bound to this API scope.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   app.api('Petstore').http().get('/pets', (req, res) => {
     *     res.data = [{ name: 'Bello' }]
     *   })
     * }
     */
    http(): HttpRouter
}

/**
 * Registers and organizes HTTP event handlers by path, operation, and method.
 *
 * Obtain an `HttpRouter` via `app.http()` or `app.api('My API').http()`.
 *
 * Two styles are supported:
 * - **Shorthand**: `router.get(path, handler)` — registers a handler directly for a path and method.
 * - **Chained**: `router.route(path).get(handler).post(handler)` — groups multiple methods under one path.
 *
 * Paths must match the path template defined in the OpenAPI specification,
 * not the absolute server URL. For example, use `/pets/{petId}`, not `/v1/pets/123`.
 *
 * When multiple APIs share the same path, use `app.api()` to scope the router
 * to a specific API title to avoid unintended matches.
 *
 * All methods return `this`, enabling fluent chaining:
 * ```ts
 * app.http()
 *   .use(authMiddleware)
 *   .get(listHandler)
 *   .post(createHandler)
 * ```
 *
 * All handlers registered on the same path and method run in registration order
 * (or by `priority` if specified). The last handler to write wins.
 *
 * @example
 * // Shorthand style
 * import { app } from 'mokapi'
 * export default function() {
 *   const router = app.http().api('Petstore')
 *   router.get('/pets', (req, res) => { res.data = [] })
 *   router.post('/pets', (req, res) => { res.statusCode = 201 })
 * }
 *
 * @example
 * // Chained style
 * import { app } from 'mokapi'
 * export default function() {
 *   app.http().api('Petstore')
 *     .route('/pets')
 *       .get((req, res) => { res.data = [] })
 *       .post((req, res) => { res.statusCode = 201 })
 * }
 */
export interface HttpRouter {

    /**
     * Registers a middleware handler that runs for all methods.
     *
     * Middleware is useful for shared logic that should execute regardless
     * of the HTTP method — for example, authentication checks, setting up
     * `res.ctx`, or adding common response headers.
     *
     * Middleware handlers run in registration order relative to other handlers
     * on the same route. Use `opts.priority` to control ordering explicitly.
     *
     * @param handler The middleware handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     *
     * @example
     * router
     *   .use((req, res) => {
     *     const token = req.header['Authorization']
     *     if (!token) {
     *       res.statusCode = 401
     *       return
     *     }
     *     res.ctx.user = parseToken(token)
     *   })
     *   .get((req, res) => {
     *     res.data = { id: req.path.petId, owner: res.ctx.user.name }
     *   })
     */
    use(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Returns an `HttpRoute` for the given OpenAPI path template.
     *
     * Use this to register multiple method handlers under the same path
     * using method chaining, or to attach middleware via `.use()`.
     *
     * The path must match the path template in the OpenAPI specification,
     * e.g. `/pets/{petId}`, not the absolute URL.
     *
     * @param path The OpenAPI path template, e.g. `/pets/{petId}`.
     * @returns An `HttpRoute` for the given path.
     *
     * @example
     * router.route('/pets/{petId}')
     *   .use((req, res) => {
     *     if (!req.header['Authorization']) {
     *       res.statusCode = 401
     *       res.stopPropagation()  // get/delete will not run
     *       return
     *     }
     *     res.context.user = parseToken(req.header['Authorization'])
     *   })
     *   .get((req, res) => {
     *     res.data = { id: req.path.petId, owner: res.context.user.name }
     *   })
     *   .delete((req, res) => {
     *     res.statusCode = 204
     *   })
     */
    route(path: string): HttpRoute

    /**
     * Returns an `HttpRoute` matched by the given `operationId` as defined
     * in the OpenAPI specification.
     *
     * Use this as an alternative to `route()` when you prefer to reference
     * operations by name rather than by path template.
     *
     * @param operationId The `operationId` defined in the OpenAPI operation.
     * @returns An `HttpRoute` for the matching operation.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   app.api('Petstore').http()
     *     .operation('listPets')
     *     .get((req, res) => { res.data = [] })
     * }
     */
    operation(operationId: string): HttpRoute

    /**
     * Narrows this router to a specific API by title.
     *
     * This is an alternative to `app.api(title).http()` for cases where
     * the router is already obtained and you want to further scope it.
     *
     * @param title The API title as defined in `info.title` of the OpenAPI spec.
     * @returns A new `HttpRouter` scoped to the given API title.
     *
     * @example
     * import { app } from 'mokapi'
     * export default function() {
     *   app.http().api('Petstore').get('/pets', (req, res) => {
     *     res.data = []
     *   })
     * }
     */
    api(title: string): HttpRouter

    /**
     * Registers a handler for `GET` requests to the given path.
     * @param path The OpenAPI path template, e.g. `/pets/{petId}`.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     *
     * @example
     * app.http().get('/pets', (req, res) => { res.data = [] })
     */
    get(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `POST` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     *
     * @example
     * app.http().post('/pets', (req, res) => { res.statusCode = 201 })
     */
    post(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `PUT` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    put(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `PATCH` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    patch(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `DELETE` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    delete(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `HEAD` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    head(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `OPTIONS` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    options(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `TRACE` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    trace(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `QUERY` requests to the given path.
     * @param path The OpenAPI path template.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     */
    query(path: string, handler: HttpEventHandler, opts?: HttpEventArgs): this

    /** Allows registering a handler for any HTTP method, including custom ones.
     * @example
     * app.http().purge('/cache', (req, res) => { res.statusCode = 200 })
     */
    [method: string]: any
}

/**
 * Represents a single route (path or operation) within an `HttpRouter`,
 * allowing multiple method handlers and middleware to be registered
 * via method chaining.
 *
 * Obtain an `HttpRoute` via `router.route(path)` or `router.operation(operationId)`.
 *
 * All methods return `this`, enabling fluent chaining:
 * ```ts
 * router.route('/pets')
 *   .use(authMiddleware)
 *   .get(listHandler)
 *   .post(createHandler)
 * ```
 *
 * Handlers registered on the same route run in registration order
 * (or by `priority`). The last handler to write a value wins.
 *
 * Middleware registered via `use()` runs for all methods on this route
 * and is well suited for shared logic such as authentication,
 * request context setup, or response enrichment.
 *
 * @example
 * import { app } from 'mokapi'
 * export default function() {
 *   app.api('Petstore').http()
 *     .route('/pets/{petId}')
 *       .use((req, res) => {
 *         // Runs before get and delete — set up shared context
 *         res.ctx.requestedAt = Date.now()
 *       })
 *       .get((req, res) => {
 *         res.data = { id: req.path.petId, requestedAt: res.ctx.requestedAt }
 *       })
 *       .delete((req, res) => {
 *         res.statusCode = 204
 *       })
 * }
 */
export interface HttpRoute {

    /**
     * Registers a middleware handler that runs for all methods on this route.
     *
     * Middleware is useful for shared logic that should execute regardless
     * of the HTTP method — for example, authentication checks, setting up
     * `res.ctx`, or adding common response headers.
     *
     * Middleware handlers run in registration order relative to other handlers
     * on the same route. Use `opts.priority` to control ordering explicitly.
     *
     * @param handler The middleware handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     *
     * @example
     * router.route('/pets/{petId}')
     *   .use((req, res) => {
     *     const token = req.header['Authorization']
     *     if (!token) {
     *       res.statusCode = 401
     *       return
     *     }
     *     res.ctx.user = parseToken(token)
     *   })
     *   .get((req, res) => {
     *     res.data = { id: req.path.petId, owner: res.ctx.user.name }
     *   })
     */
    use(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `GET` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     *
     * @example
     * router.route('/pets').get((req, res) => { res.data = [] })
     */
    get(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `POST` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     *
     * @example
     * router.route('/pets').post((req, res) => { res.statusCode = 201 })
     */
    post(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `PUT` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    put(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `PATCH` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    patch(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `DELETE` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    delete(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `HEAD` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    head(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `OPTIONS` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    options(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `TRACE` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    trace(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /**
     * Registers a handler for `QUERY` requests on this route.
     * @param handler The event handler to invoke.
     * @param opts Optional handler configuration such as priority and tracking.
     * @returns This `HttpRoute` instance for chaining.
     */
    query(handler: HttpEventHandler, opts?: HttpEventArgs): this

    /** Allows registering a handler for any HTTP method, including custom ones such as `PURGE` or `SEARCH`.
     * @example
     * router.route('/cache').purge((req, res) => { res.statusCode = 200 })
     */
    [method: string]: any
}