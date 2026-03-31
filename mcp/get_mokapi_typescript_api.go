package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetMokapiTypeScriptApiInput struct {
	Package string `json:"package"`
}

type GetMokapiTypeScriptApiOutput struct {
	Package string `json:"package"`
	Types   string `json:"types"`
}

func (s *Service) registerGetMokapiTypeScriptApi(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"package": map[string]any{
				"type":        "string",
				"description": "The name of the package to fetch TypeScript definitions for, e.g., 'mokapi/http'",
			},
		},
		"required": []string{"package"},
	}

	registerTool(server, &mcp.Tool{
		Name: "get_mokapi_typescript_api",
		Description: `Returns TypeScript definitions for a specific Mokapi package.

Use this tool after selecting a package via "get_mokapi_typescript_api_list".

The returned types define:
- Event handler signatures
- Request and response structures
- Available properties

Combine this with "get_scenario" to understand correct usage patterns.`,
		InputSchema: inputSchema,
	}, s.GetMokapiTypeScriptApi)
}

func (s *Service) GetMokapiTypeScriptApi(_ context.Context, in GetMokapiTypeScriptApiInput) (GetMokapiTypeScriptApiOutput, error) {
	switch in.Package {
	case "mokapi":
		return GetMokapiTypeScriptApiOutput{
			Package: "mokapi",
			Types:   pkgMokapi,
		}, nil
	case "mokapi/http":
		return GetMokapiTypeScriptApiOutput{
			Package: "mokapi/http",
			Types:   pkgHttp,
		}, nil
	case "mokapi/kafka":
		return GetMokapiTypeScriptApiOutput{
			Package: "mokapi/kafka",
			Types:   pkgKafka,
		}, nil
	case "mokapi/faker":
		return GetMokapiTypeScriptApiOutput{
			Package: "mokapi/faker",
			Types:   pkgFaker,
		}, nil
	case "mokapi/file":
		return GetMokapiTypeScriptApiOutput{
			Package: "mokapi/file",
			Types:   pkgFile,
		}, nil
	}
	return GetMokapiTypeScriptApiOutput{}, fmt.Errorf("unknown Mokapi package: %s", in.Package)
}

const (
	pkgMokapi = `
export function on<T extends keyof EventHandler>(event: T, handler: EventHandler[T], args?: TypedEventArgs[T]): void;

export function every(interval: Interval, f: ScheduledEventHandler, args?: ScheduledEventArgs): void;

export function cron(expr: string, f: ScheduledEventHandler, args?: ScheduledEventArgs): void;

export function env(name: string): string;

export function date(args?: DateArgs): string;

export function sleep(time: number | string): void;

export type Interval = string;

export interface EventHandler {
    http: HttpEventHandler;
    kafka: KafkaEventHandler;
    ldap: LdapEventHandler;
    smtp: SmtpEventHandler;
}

export type HttpEventHandler = (request: HttpRequest, response: HttpResponse) => void | Promise<void>;

export interface HttpRequest {
    readonly method: string;
    readonly url: Url;
    readonly body: any;
    readonly path: { [key: string]: any };
    readonly query: { [key: string]: any };
    readonly header: { [key: string]: any };
    readonly cookie: { [key: string]: any };
    readonly querystring: any;
    readonly api: string;
    readonly key: string;
    readonly operationId: string;
    toString(): string;
}

export interface HttpResponse {
    headers: { [key: string]: any };
    statusCode: number;
    body: string;
    data: any;
    rebuild: (statusCode?: number, contentType?: string) => void;
}

export interface Url {
    readonly scheme: string;
    readonly host: string;
    readonly port: number;
    readonly path: string;
    readonly query: string;
    toString(): string;
}

export type KafkaEventHandler = (message: KafkaEventMessage) => void | Promise<void>;

export interface KafkaEventMessage {
    readonly offset: number;
    key: string;
    value: string;
    headers: { [name: string]: string } | null;
}

export type LdapEventHandler = (request: LdapSearchRequest, response: LdapSearchResponse) => void | Promise<void>;

export interface LdapSearchRequest {
    baseDN: string;
    scope: LdapSearchScope;
    dereferencePolicy: number;
    sizeLimit: number;
    timeLimit: number;
    typesOnly: number;
    filter: string;
    attributes: string[];
}
export interface LdapSearchResponse {
    results: LdapSearchResult[];
    status: LdapResultStatus;
    message: string;
}

export interface LdapSearchResult {
    dn: string;
    attributes: { [name: string]: string[] };
}

export enum LdapSearchScope {
    BaseObject,
    SingleLevel,
    WholeSubtree,
}

export enum LdapResultStatus {
    Success = 0,
    OperationsError = 1,
    ProtocolError = 2,
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

export interface DateArgs {
    layout?: DateLayout | string;
    timestamp?: number;
}

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

export interface EventArgs {
    tags?: { [key: string]: string };
    priority?: number;
}

export interface TypedEventArgs {
    http: HttpEventArgs;
    kafka: KafkaEventArgs;
    ldap: LdapEventArgs;
    smtp: SmtpEventArgs;
}

export interface HttpEventArgs extends EventArgs {
    track?: boolean | ((request: HttpRequest, response: HttpResponse) => boolean);
}

export interface KafkaEventArgs extends EventArgs {
    track?: boolean | ((message: KafkaEventMessage) => boolean);
}

export interface LdapEventArgs extends EventArgs {
    track?: boolean | ((request: LdapSearchRequest, response: LdapSearchResponse) => boolean);
}

export interface SmtpEventArgs extends EventArgs {
    track?: boolean | ((record: SmtpEventMessage) => boolean);
}

export type ScheduledEventHandler = () => void | Promise<void>;

export interface ScheduledEventArgs {
    tags?: { [key: string]: string };
    times?: number;
    runFirstTimeImmediately?: boolean;
}

export const RFC3339 = "RFC3339";

export function patch(target: any, patch: any): any;

export const Delete: unique symbol;

export interface SharedMemory {
    get(key: string): any;
    set(key: string, value: any): void;
    update<T = any>(key: string, updater: (value: T | undefined) => T): T;
    has(key: string): boolean;
    delete(key: string): void;
    clear(): void;
    keys(): string[];
    namespace(name: string): SharedMemory;
}

export const shared: SharedMemory;
`
	pkgHttp = `export function get(url: string, args?: Args): Response;
export function post(url: string, body?: any, args?: Args): Response;
export function put(url: string, body?: any, args?: Args): Response;
export function head(url: string, args?: Args): Response;
export function patch(url: string, body?: any, args?: Args): Response;
export function del(url: string, body?: any, args?: Args): Response;
export function options(url: string, body?: any, args?: Args): Response;
export function fetch(url: string, opts?: FetchOptions): Promise<Response>

export interface FetchOptions {
    method?: string;
    body?: any;
    headers?: { [name: string]: string };
    maxRedirects?: number;
    timeout?: number | string;
}

export interface Args {
    headers?: { [name: string]: string };
    maxRedirects?: number;
    timeout?: number | string;
}

export interface Response {
    body: string;
    statusCode: number;
    headers: { [name: string]: string[] };
    json(): JSONValue;
}
`
	pkgKafka = `export function produce(args?: ProduceArgs): ProduceResult;
export function produceAsync(args?: ProduceArgs): Promise<ProduceResult>;
export interface ProduceArgs {
    cluster?: string;
    topic?: string;
    messages?: Message[];
    retry?: ProduceRetry;
}

export interface Message {
    partition?: number;
    key?: any;
    data?: any;
    value?: string | number | boolean | null;
    headers?: { [name: string]: any };
}

export interface ProduceResult {
    readonly cluster: string;
    readonly topic: string;
    messages: MessageResult[];
    readonly partition: number;
    readonly offset: number;
    readonly key: string;
    readonly value: string;
    readonly headers: { [name: string]: string };
}

export interface MessageResult {
    readonly partition: number;
    readonly offset: number;
    readonly key: string;
    readonly value: string;
    readonly headers: { [name: string]: string };
}

export interface ProduceRetry {
    maxRetryTime: string | number;
    initialRetryTime: string | number;
    factor: number;
    retries: number;
}`
	pkgFaker = `export function fake(schema: Schema | JSONSchema): any;

export function findByName(name: string): Node;
export const ROOT_NAME = "root";
export interface Node {
    name: string;
    attributes: string[];
    weight: number;
    children: Array<Node | AddNode>;
    fake: (r: Request) => any;
}

export interface AddNode {
    name: string;
    attributes?: string[];
    weight?: number;
    children?: Array<Node | AddNode>;
    fake: (r: Request) => any;
}

export interface Request {
    path: string[];
    schema: JSONSchema;
    context: Context;
}

export interface Context {
    values: { [name: string]: any };
}

export interface JSONSchema {
    type?: SchemaType | SchemaType[];
    enum?: any[];
    const?: any;
    examples?: any[];
    default?: any;
    multipleOf?: number;
    maximum?: number;
    exclusiveMaximum?: number;
    minimum?: number;
    exclusiveMinimum?: number;
    maxLength?: number;
    minLength?: number;
    pattern?: string;
    format?: string;
    items?: JSONSchema;
    maxItems?: number;
    minItems?: number;
    uniqueItems?: boolean;
    properties?: { [name: string]: JSONSchema };
    maxProperties?: number;
    minProperties?: number;
    required?: string[];
    additionalProperties?: boolean | JSONSchema;
    allOf?: JSONSchema[];
    anyOf?: JSONSchema[];
    oneOf?: JSONSchema[];
}

export type SchemaType = "object" | "array" | "number" | "integer" | "string" | "boolean" | "null";

export interface Schema {
    type?: SchemaType | SchemaType[];
    format?: string;
    pattern?: string;
    minLength?: number;
    maxLength?: number;
    items?: Schema;
    required?: string[];
    enum?: any[];
    minimum?: number;
    maximum?: number;
    exclusiveMinimum?: number | boolean;
    exclusiveMaximum?: number | boolean;
    properties?: { [name: string]: Schema };
    additionalProperties?: boolean | Schema | undefined;
    anyOf?: Schema[];
    allOf?: Schema[];
    oneOf?: Schema[];
    minItems?: number;
    maxItems?: number;
    shuffleItems?: boolean;
    uniqueItems?: boolean;
}`
	pkgFile = `export function read(path: string): string;
export function writeString(path: string, s: string): void;
export function appendString(path: string, s: string): void;`
)
