package models

import (
	"fmt"
	"strconv"
	"strings"
)

type WebService struct {
	// The title of the service
	Name string

	// A short description of the API. CommonMark syntax MAY be
	// used for rich text representation.
	Description string

	// The version of the service
	Version  string
	Servers  []Server
	Endpoint map[string]*Endpoint
	Models   map[string]*Schema

	// The mokapi file used by this service.
	MokapiFile string
}

func (s *WebService) AddServer(server Server) {
	for _, v := range s.Servers {
		if v.Host == server.Host && v.Port == server.Port && v.Path == server.Path {
			return
		}
	}
	s.Servers = append(s.Servers, server)
}

func (w *WebService) Key() string {
	return w.Name
}

type Server struct {
	// The server host name
	Host string

	// The server port number
	Port int

	// A relative path to the location where the OpenAPI definition
	// is being served
	Path string

	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string
}

type Endpoint struct {
	// A relative path to an individual endpoint. The path MUST begin
	// with a forward slash ('/'). The path is appended to the url from
	// server objects url field in order to construct the full URL
	Path string

	// An optional, string summary, intended to apply to all operations
	// in this path.
	Summary string

	// An optional, string description, intended to apply to all operations
	// in this path. CommonMark syntax MAY be used for rich text representation.
	Description string

	// A definition of a GET operation on this path.
	Get *Operation

	// A definition of a POST operation on this path.
	Post *Operation

	// A definition of a PUT operation on this path.
	Put *Operation

	// A definition of a PATCH operation on this path.
	Patch *Operation

	// A definition of a DELETE operation on this path.
	Delete *Operation

	// A definition of a HEAD operation on this path.
	Head *Operation

	// A definition of a OPTIONS operation on this path.
	Options *Operation

	// A definition of a TRACE operation on this path.
	Trace *Operation

	// TODO: implementing
	// A list of parameters that are applicable for all
	// the operations described under this path. These
	// parameters can be overridden at the operation level,
	// but cannot be removed there
	Parameters []*Parameter

	// The pipeline name used for all the operation described
	// under this path. This pipeline name can be overridden
	// at the operation level, but cannot reset to the default
	// empty pipeline name.
	Pipeline string

	// The service which this endpoint belongs to
	Service *WebService
}

func NewEndpoint(path string, service *WebService) *Endpoint {
	return &Endpoint{Path: path, Service: service}
}

// Gets the operation for the given method name
func (e *Endpoint) GetOperation(method string) *Operation {
	switch strings.ToUpper(method) {
	case "GET":
		return e.Get
	case "POST":
		return e.Post
	case "Put":
		return e.Put
	case "Patch":
		return e.Patch
	case "Delete":
		return e.Delete
	case "Head":
		return e.Head
	case "Options":
		return e.Options
	case "Trace":
		return e.Trace
	}

	return nil
}

type Operation struct {
	// A short summary of what the operation does.
	Summary string

	// A verbose explanation of the operation behavior.
	// CommonMark syntax MAY be used for rich text representation.
	Description string

	// Unique string used to identify the operation. The id MUST be unique
	// among all operations described in the API. The operationId value is
	// case-sensitive. Tools and libraries MAY use the operationId to uniquely
	// identify an operation, therefore, it is RECOMMENDED to follow common
	// programming naming conventions.
	OperationId string

	// todo: implement feature overridable
	// A list of parameters that are applicable for this operation.
	// If a parameter is already defined at the Path Item, the new definition
	// will override it but can never remove it. The list MUST NOT include
	// duplicated parameters. A unique parameter is defined by a combination
	// of a name and location
	Parameters []*Parameter

	RequestBody *RequestBody

	// The list of possible responses as they are returned from executing this
	// operation.
	Responses map[HttpStatus]*Response

	// The pipeline name used to identify the pipeline in the mokapi file.
	// If pipeline name is already defined at the Path Item, the new definition
	// will override it but can not set to empty pipeline name.
	Pipeline string

	// The endpoint which this operation belongs to
	Endpoint *Endpoint
}

func NewOperation(summary string, description string, operationId string, pipeline string, endpoint *Endpoint) *Operation {
	return &Operation{
		Summary:     summary,
		Description: description,
		OperationId: operationId,
		Endpoint:    endpoint,
		Pipeline:    pipeline,
		Responses:   make(map[HttpStatus]*Response),
	}
}

type HttpStatus int

const (
	Invalid            HttpStatus = -1
	Continue           HttpStatus = 100 // RFC 7231, 6.2.1
	SwitchingProtocols HttpStatus = 101 // RFC 7231, 6.2.2
	Processing         HttpStatus = 102 // RFC 2518, 10.1
	EarlyHints         HttpStatus = 103 // RFC 8297

	OK                   HttpStatus = 200 // RFC 7231, 6.3.1
	Created              HttpStatus = 201 // RFC 7231, 6.3.2
	Accepted             HttpStatus = 202 // RFC 7231, 6.3.3
	NonAuthoritativeInfo HttpStatus = 203 // RFC 7231, 6.3.4
	NoContent            HttpStatus = 204 // RFC 7231, 6.3.5
	ResetContent         HttpStatus = 205 // RFC 7231, 6.3.6
	PartialContent       HttpStatus = 206 // RFC 7233, 4.1
	MultiStatus          HttpStatus = 207 // RFC 4918, 11.1
	AlreadyReported      HttpStatus = 208 // RFC 5842, 7.1
	IMUsed               HttpStatus = 226 // RFC 3229, 10.4.1

	MultipleChoices   HttpStatus = 300 // RFC 7231, 6.4.1
	MovedPermanently  HttpStatus = 301 // RFC 7231, 6.4.2
	Found             HttpStatus = 302 // RFC 7231, 6.4.3
	SeeOther          HttpStatus = 303 // RFC 7231, 6.4.4
	NotModified       HttpStatus = 304 // RFC 7232, 4.1
	UseProxy          HttpStatus = 305 // RFC 7231, 6.4.5
	_                 HttpStatus = 306 // RFC 7231, 6.4.6 (Unused)
	TemporaryRedirect HttpStatus = 307 // RFC 7231, 6.4.7
	PermanentRedirect HttpStatus = 308 // RFC 7538, 3

	BadRequest                   HttpStatus = 400 // RFC 7231, 6.5.1
	Unauthorized                 HttpStatus = 401 // RFC 7235, 3.1
	PaymentRequired              HttpStatus = 402 // RFC 7231, 6.5.2
	Forbidden                    HttpStatus = 403 // RFC 7231, 6.5.3
	NotFound                     HttpStatus = 404 // RFC 7231, 6.5.4
	MethodNotAllowed             HttpStatus = 405 // RFC 7231, 6.5.5
	NotAcceptable                HttpStatus = 406 // RFC 7231, 6.5.6
	ProxyAuthRequired            HttpStatus = 407 // RFC 7235, 3.2
	RequestTimeout               HttpStatus = 408 // RFC 7231, 6.5.7
	Conflict                     HttpStatus = 409 // RFC 7231, 6.5.8
	Gone                         HttpStatus = 410 // RFC 7231, 6.5.9
	LengthRequired               HttpStatus = 411 // RFC 7231, 6.5.10
	PreconditionFailed           HttpStatus = 412 // RFC 7232, 4.2
	RequestEntityTooLarge        HttpStatus = 413 // RFC 7231, 6.5.11
	RequestURITooLong            HttpStatus = 414 // RFC 7231, 6.5.12
	UnsupportedMediaType         HttpStatus = 415 // RFC 7231, 6.5.13
	RequestedRangeNotSatisfiable HttpStatus = 416 // RFC 7233, 4.4
	ExpectationFailed            HttpStatus = 417 // RFC 7231, 6.5.14
	Teapot                       HttpStatus = 418 // RFC 7168, 2.3.3
	MisdirectedRequest           HttpStatus = 421 // RFC 7540, 9.1.2
	UnprocessableEntity          HttpStatus = 422 // RFC 4918, 11.2
	Locked                       HttpStatus = 423 // RFC 4918, 11.3
	FailedDependency             HttpStatus = 424 // RFC 4918, 11.4
	TooEarly                     HttpStatus = 425 // RFC 8470, 5.2.
	UpgradeRequired              HttpStatus = 426 // RFC 7231, 6.5.15
	PreconditionRequired         HttpStatus = 428 // RFC 6585, 3
	TooManyRequests              HttpStatus = 429 // RFC 6585, 4
	RequestHeaderFieldsTooLarge  HttpStatus = 431 // RFC 6585, 5
	UnavailableForLegalReasons   HttpStatus = 451 // RFC 7725, 3

	InternalServerError           HttpStatus = 500 // RFC 7231, 6.6.1
	NotImplemented                HttpStatus = 501 // RFC 7231, 6.6.2
	BadGateway                    HttpStatus = 502 // RFC 7231, 6.6.3
	ServiceUnavailable            HttpStatus = 503 // RFC 7231, 6.6.4
	GatewayTimeout                HttpStatus = 504 // RFC 7231, 6.6.5
	HTTPVersionNotSupported       HttpStatus = 505 // RFC 7231, 6.6.6
	VariantAlsoNegotiates         HttpStatus = 506 // RFC 2295, 8.1
	InsufficientStorage           HttpStatus = 507 // RFC 4918, 11.5
	LoopDetected                  HttpStatus = 508 // RFC 5842, 7.2
	NotExtended                   HttpStatus = 510 // RFC 2774, 7
	NetworkAuthenticationRequired HttpStatus = 511 // RFC 6585, 6
)

func parseHttpStatus(s string) (HttpStatus, error) {
	i, error := strconv.Atoi(s)
	if error != nil {
		return Invalid, fmt.Errorf("Can not parse status code %v", s)
	}
	status := HttpStatus(i)
	if !isValidHttpStatus(status) {
		return Invalid, fmt.Errorf("Unsupport status code %v", s)
	}

	return status, nil
}

func isValidHttpStatus(status HttpStatus) bool {
	switch status {
	case
		// 100
		Invalid, Continue, SwitchingProtocols, Processing, EarlyHints,
		// 200
		OK, Created, Accepted, NonAuthoritativeInfo, NoContent, ResetContent, PartialContent, MultiStatus, AlreadyReported, IMUsed,
		// 300
		MultipleChoices, MovedPermanently, Found, SeeOther, NotModified, UseProxy, TemporaryRedirect, PermanentRedirect,
		// 400
		BadRequest, Unauthorized, PaymentRequired, Forbidden, NotFound, MethodNotAllowed, NotAcceptable, ProxyAuthRequired,
		RequestTimeout, Conflict, Gone, LengthRequired, PreconditionFailed, RequestEntityTooLarge, RequestURITooLong,
		UnsupportedMediaType, RequestedRangeNotSatisfiable, ExpectationFailed, Teapot, MisdirectedRequest,
		UnprocessableEntity, Locked, FailedDependency, TooEarly, UpgradeRequired, PreconditionRequired,
		TooManyRequests, RequestHeaderFieldsTooLarge, UnavailableForLegalReasons,
		// 500
		InternalServerError, NotImplemented, BadGateway, ServiceUnavailable, GatewayTimeout, HTTPVersionNotSupported,
		VariantAlsoNegotiates, InsufficientStorage, LoopDetected, NotExtended, NetworkAuthenticationRequired:
		return true

	default:
		return false
	}
}

type Parameter struct {
	// The name of the parameter. Parameter names are case sensitive.
	Name string

	// The location of the parameter
	Location ParameterLocation

	// The schema defining the type used for the parameter
	Schema *Schema

	// Determines whether the parameter is mandatory.
	// If the location of the parameter is "path", this property
	// is required and its value MUST be true
	Required bool

	// A brief description of the parameter. This could contain examples
	// of use. CommonMark syntax MAY be used for rich text representation.
	Description string

	// Defines how multiple values are delimited. Possible styles depend on
	// the parameter location
	Style string

	// specifies whether arrays and objects should generate separate
	// parameters for each array item or object property
	Explode bool
}

type ParameterLocation int

const (
	PathParameter   ParameterLocation = 1
	QueryParameter  ParameterLocation = 2
	HeaderParameter ParameterLocation = 3
	CookieParameter ParameterLocation = 4
)

func (p ParameterLocation) String() string {
	switch p {
	case PathParameter:
		return "path"
	case QueryParameter:
		return "query"
	case HeaderParameter:
		return "header"
	case CookieParameter:
		return "cookie"
	default:
		return "unknown"
	}
}

type RequestBody struct {
	// A brief description of the request body. This could contain
	// examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string

	// The content of the request body. The key is a media type or media type range
	// and the value describes it. For requests that match multiple keys, only the
	// most specific key is applicable. e.g. text/plain overrides text/*
	ContentTypes map[string]*MediaType

	// Determines if the request body is required in the request. Defaults to false.
	Required bool
}

type MediaType struct {
	// The schema defining the content of the request, response.
	Schema *Schema
}

type Response struct {
	// A short description of the response. CommonMark syntax
	// MAY be used for rich text representation.
	Description string

	// A map containing descriptions of potential response payloads.
	// The key is a media type or media type range and the value describes
	// it. For responses that match multiple keys, only the most specific
	// key is applicable. e.g. text/plain overrides text/*
	ContentTypes map[string]*MediaType
}

type ContentType struct {
	Type       string
	Subtype    string
	Parameters map[string]string
	raw        string
}

func ParseContentType(s string) *ContentType {
	c := &ContentType{raw: s, Parameters: make(map[string]string)}
	a := strings.Split(s, ";")
	m := strings.Split(a[0], "/")
	c.Type = strings.ToLower(strings.TrimSpace(m[0]))
	if len(m) > 1 {
		c.Subtype = strings.ToLower(strings.TrimSpace(m[1]))
	}
	for _, p := range a[1:] {
		kv := strings.Split(p, "=")
		if len(kv) > 1 {
			c.Parameters[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			c.Parameters[kv[0]] = ""
		}
	}

	return c
}

func (c *ContentType) Key() string {
	if len(c.Subtype) > 0 {
		return fmt.Sprintf("%v/%v", c.Type, c.Subtype)
	}
	return c.Type
}

func (c *ContentType) String() string {
	return c.raw
}

func (c *ContentType) Equals(other *ContentType) bool {
	return c.Type == other.Type && c.Subtype == other.Subtype
}

type Schema struct {
	Name                 string
	Type                 string
	Format               string
	Description          string
	Properties           map[string]*Schema
	Faker                string
	Items                *Schema
	Xml                  *XmlEncoding
	AdditionalProperties *Schema
	Reference            string
	Required             []string
	isResolved           bool
	Nullable             bool
}

func (s *Schema) IsPropertyRequired(name string) bool {
	if s.Required == nil {
		return false
	}
	for _, p := range s.Required {
		if p == name {
			return true
		}
	}
	return false
}

type XmlEncoding struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
