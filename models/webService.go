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
	Models   []*Schema

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
	Invalid             HttpStatus = -1
	Ok                  HttpStatus = 200
	Created             HttpStatus = 201
	Accepted            HttpStatus = 202
	NoContent           HttpStatus = 204
	MovedPermanently    HttpStatus = 301
	MovedTemporarily    HttpStatus = 302
	NotModified         HttpStatus = 304
	BadRequest          HttpStatus = 400
	Unauthorized        HttpStatus = 401
	Forbidden           HttpStatus = 403
	NotFound            HttpStatus = 404
	MethodNotAllowed    HttpStatus = 405
	InternalServerError HttpStatus = 500
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
	case Ok, Created, Accepted, NoContent, MovedPermanently,
		MovedTemporarily, NotModified, BadRequest, Unauthorized,
		Forbidden, NotFound, MethodNotAllowed, InternalServerError:
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
	Type                 string
	Format               string
	Description          string
	Properties           map[string]*Schema
	Faker                string
	Items                *Schema
	Xml                  *XmlEncoding
	AdditionalProperties string
	Reference            string
	Required             []string
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
