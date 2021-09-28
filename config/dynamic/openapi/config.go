package openapi

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mokapi"
	"mokapi/models/media"
	"net/http"
	"net/url"
	"strconv"
)

func init() {
	dynamic.Register("openapi", &Config{}, func(path string, o dynamic.Config, cr dynamic.ConfigReader) (bool, dynamic.Config) {
		eh := dynamic.NewEmptyEventHandler(o)
		switch c := o.(type) {
		case *Config:
			c.ConfigPath = path

			if len(c.Info.Name) == 0 {
				log.Errorf("missing required property title: %v", path)
				return false, nil
			}

			r := ReferenceResolver{reader: cr, path: path, config: c, eh: eh}

			if err := r.ResolveConfig(); err != nil {
				log.Errorf("error in resolving references in config %q: %v", path, err)
			}

			return true, c
		}
		return false, nil
	})
}

type Config struct {
	ConfigPath string `yaml:"-" json:"-"`
	Info       Info
	Servers    []*Server

	// A relative path to an individual endpoint. The path MUST begin
	// with a forward slash ('/'). The path is appended to the url from
	// server objects url field in order to construct the full URL
	EndPoints  map[string]*EndpointRef `yaml:"paths" json:"paths"`
	Components Components
}

func (c *Config) Key() string {
	return c.ConfigPath
}

type Info struct {
	// The title of the service
	Name string `yaml:"title" json:"title"`

	// A short description of the API. CommonMark syntax MAY be
	// used for rich text representation.
	Description string

	// The version of the service
	Version string
	//Mokapi  *MokapiRef `yaml:"x-mokapi" json:"x-mokapi"`
}

type MokapiRef struct {
	Ref   string
	Value *mokapi.Config
}

type Server struct {
	Url string

	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string
}

type EndpointRef struct {
	Ref   string
	Value *Endpoint
}

type Endpoint struct {
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

	// A list of parameters that are applicable for all
	// the operations described under this path. These
	// parameters can be overridden at the operation level,
	// but cannot be removed there
	Parameters Parameters

	// The pipeline name used for all the operation described
	// under this path. This pipeline name can be overridden
	// at the operation level, but cannot reset to the default
	// empty pipeline name.
	Pipeline string `yaml:"x-mokapi-pipeline" json:"x-mokapi-pipeline"`
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

	// A list of parameters that are applicable for this operation.
	// If a parameter is already defined at the Path Item, the new definition
	// will override it but can never remove it. The list MUST NOT include
	// duplicated parameters. A unique parameter is defined by a combination
	// of a name and location
	Parameters Parameters

	RequestBody *RequestBodyRef `yaml:"requestBody" json:"requestBody"`

	// The list of possible responses as they are returned from executing this
	// operation.
	Responses Responses `yaml:"responses" json:"responses"`

	// The pipeline name used to identify the pipeline in the mokapi file.
	// If pipeline name is already defined at the Path Item, the new definition
	// will override it but can not set to empty pipeline name.
	Pipeline string `yaml:"x-mokapi-pipeline" yaml:"x-mokapi-pipeline"`

	Endpoint *Endpoint `yaml:"-" json:"-"`
}

type Parameters []*ParameterRef

type ParameterRef struct {
	Ref   string
	Value *Parameter
}

type Parameter struct {
	// The name of the parameter. Parameter names are case sensitive.
	Name string

	// The location of the parameter
	Type ParameterLocation `yaml:"in" json:"in"`

	// The schema defining the type used for the parameter
	Schema *SchemaRef

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

type ParameterLocation string

type HttpStatus int

func (s HttpStatus) String() string {
	return strconv.Itoa(int(s))
}

type SchemaRef struct {
	Ref   string `yaml:"$ref" json:"$ref"`
	Value *Schema
}

type SchemaRefs []*SchemaRef

type Schema struct {
	Type                 string
	Format               string
	Pattern              string
	Description          string
	Properties           *Schemas
	AdditionalProperties *SchemaRef // TODO custom marshal for bool, {} etc. Should it be a schema reference?
	Faker                string     `yaml:"x-faker" json:"x-faker"`
	Items                *SchemaRef
	Xml                  *Xml
	Required             []string
	Nullable             bool
	Example              interface{}
	Enum                 []interface{}
	Minimum              *float64   `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	Maximum              *float64   `yaml:"maximum,omitempty" json:"maximum,omitempty"`
	ExclusiveMinimum     *bool      `yaml:"exclusiveMinimum,omitempty" json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     *bool      `yaml:"exclusiveMaximum ,omitempty" json:"exclusiveMaximum,omitempty"`
	AnyOf                SchemaRefs `yaml:"anyOf" json:"anyOf"`
	AllOf                SchemaRefs `yaml:"allOf" json:"allOf"`
	OneOf                SchemaRefs `yaml:"oneOf" json:"oneOf"`
	UniqueItems          bool       `yaml:"uniqueItems" json:"uniqueItems"`
	MinItems             *int       `yaml:"minItems" json:"minItems"`
	MaxItems             *int       `yaml:"maxItems" json:"maxItems"`
	ShuffleItems         bool       `yaml:"x-shuffleItems" json:"x-shuffleItems"`
}

type AdditionalProperties struct {
	Schema *Schema
}

type RequestBodyRef struct {
	Ref   string `yaml:"$ref" json:"$ref"`
	Value *RequestBody
}

type RequestBody struct {
	// A brief description of the request body. This could contain
	// examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string

	// The content of the request body. The key is a media type or media type range
	// and the value describes it. For requests that match multiple keys, only the
	// most specific key is applicable. e.g. text/plain overrides text/*
	Content map[string]*MediaType

	// Determines if the request body is required in the request. Defaults to false.
	Required bool
}

type Responses map[HttpStatus]*ResponseRef

type ResponseRef struct {
	Ref   string `yaml:"$ref" json:"$ref"`
	Value *Response
}

type Response struct {
	// A short description of the response. CommonMark syntax
	// MAY be used for rich text representation.
	Description string

	// A map containing descriptions of potential response payloads.
	// The key is a media type or media type range and the value describes
	// it. For responses that match multiple keys, only the most specific
	// key is applicable. e.g. text/plain overrides text/*
	Content map[string]*MediaType

	// Maps a header name to its definition. RFC7230 states header names are
	// case insensitive. If a response header is defined with the name
	// "Content-Type", it SHALL be ignored.
	Headers map[string]*HeaderRef
}

type MediaType struct {
	// The schema defining the content of the request, response.
	Schema   *SchemaRef
	Example  interface{}
	Examples map[string]*ExampleRef
}

type HeaderRef struct {
	Ref   string `yaml:"$ref" json:"$ref"`
	Value *Header
}

type Header struct {
	Name        string
	Description string
	Schema      *SchemaRef
}

type Example struct {
	Summary     string
	Value       interface{}
	Description string
}

type ExampleRef struct {
	Ref   string `yaml:"$ref" json:"$ref"`
	Value *Example
}

type Components struct {
	Schemas       *Schemas
	Responses     *NamedResponses
	RequestBodies *RequestBodies `yaml:"requestBodies" json:"requestBodies"`
	Parameters    *NamedParameters
	Examples      *Examples
	Headers       *NamedHeaders
}

type Schemas struct {
	Ref   string
	Value map[string]*SchemaRef
}

type NamedResponses struct {
	Ref   string
	Value map[string]*ResponseRef
}

type NamedParameters struct {
	Ref   string
	Value map[string]*ParameterRef
}

type NamedHeaders struct {
	Ref   string
	Value map[string]*HeaderRef
}

type Examples struct {
	Ref   string
	Value map[string]*ExampleRef
}

type RequestBodies struct {
	Ref   string
	Value map[string]*RequestBodyRef
}

type Xml struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool `yaml:"x-cdata" json:"x-cdata"`
}

const (
	PathParameter   ParameterLocation = "path"
	QueryParameter  ParameterLocation = "query"
	HeaderParameter ParameterLocation = "header"
	CookieParameter ParameterLocation = "cookie"

	Invalid            HttpStatus = -1
	Undefined          HttpStatus = 0
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

func (s *Schemas) Get(name string) (*SchemaRef, bool) {
	if s.Value == nil {
		return nil, false
	}
	p, ok := s.Value[name]
	return p, ok
}

func (r *RequestBody) GetMedia(contentType *media.ContentType) (*MediaType, bool) {
	if c, ok := r.Content[contentType.String()]; ok {
		return c, true
	} else if c, ok := r.Content[contentType.Key()]; ok {
		return c, true
	}
	return nil, false
}

func (s *Server) GetPort() int {
	u, err := url.Parse(s.Url)
	if err != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return -1
	}
	portString := u.Port()
	if len(portString) == 0 {
		if u.Scheme == "https" {
			return 443
		}
		return 80
	} else {
		port, err := strconv.ParseInt(portString, 10, 32)
		if err != nil {
			log.WithField("url", s.Url).Error("Invalid port format in url found.")
		}
		return int(port)
	}
}

func (s *Server) GetHost() string {
	u, err := url.Parse(s.Url)
	if err != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	return u.Hostname()
}

func (s *Server) GetPath() string {
	u, err := url.Parse(s.Url)
	if err != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	if len(u.Path) == 0 {
		return "/"
	}
	return u.Path
}

func (e *Endpoint) Operations() map[string]*Operation {
	operations := make(map[string]*Operation, 4)
	if v := e.Get; v != nil {
		operations[http.MethodGet] = v
	}
	if v := e.Patch; v != nil {
		operations[http.MethodPatch] = v
	}
	if v := e.Post; v != nil {
		operations[http.MethodPost] = v
	}
	if v := e.Put; v != nil {
		operations[http.MethodPut] = v
	}
	if v := e.Delete; v != nil {
		operations[http.MethodDelete] = v
	}
	if v := e.Head; v != nil {
		operations[http.MethodHead] = v
	}
	if v := e.Options; v != nil {
		operations[http.MethodOptions] = v
	}
	if v := e.Trace; v != nil {
		operations[http.MethodTrace] = v
	}

	return operations
}
