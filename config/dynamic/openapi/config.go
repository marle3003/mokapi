package openapi

import (
	"errors"
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
	"mokapi/sortedmap"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	common.Register("openapi", &Config{})
}

type Config struct {
	OpenApi string    `yaml:"openapi" json:"openapi"`
	Info    Info      `yaml:"info" json:"info"`
	Servers []*Server `yaml:"servers,omitempty" json:"servers,omitempty"`

	// A relative path to an individual endpoint. The path MUST begin
	// with a forward slash ('/'). The path is appended to the url from
	// server objects url field in order to construct the full URL
	Paths      EndpointsRef `yaml:"paths,omitempty" json:"paths,omitempty"`
	Components Components   `yaml:"components,omitempty" json:"components,omitempty"`
}

type Info struct {
	// The title of the service
	Name string `yaml:"title" json:"title"`

	// A short description of the API. CommonMark syntax MAY be
	// used for rich text representation.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	Contact *Contact `yaml:"contact,omitempty" json:"contact,omitempty"`

	// The version of the service
	Version string `yaml:"version" json:"version"`
}

type Contact struct {
	Name  string `yaml:"name" json:"name"`
	Url   string `yaml:"url" json:"url"`
	Email string `yaml:"email" json:"email"`
}

type Server struct {
	Url string

	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string
}

type EndpointsRef struct {
	ref.Reference
	Value map[string]*EndpointRef
}

type EndpointRef struct {
	ref.Reference
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
	Parameters parameter.Parameters
}

type Operation struct {
	// A list of tags for API documentation control. Tags can be used for
	// logical grouping of operations by resources or any other qualifier.
	Tags []string `yaml:"tags" json:"tags"`

	// A short summary of what the operation does.
	Summary string `yaml:"summary" json:"summary"`

	// A verbose explanation of the operation behavior.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `yaml:"description" json:"description"`

	Deprecated bool `yaml:"deprecated" json:"deprecated"`

	// Unique string used to identify the operation. The id MUST be unique
	// among all operations described in the API. The operationId value is
	// case-sensitive. Tools and libraries MAY use the operationId to uniquely
	// identify an operation, therefore, it is RECOMMENDED to follow common
	// programming naming conventions.
	OperationId string `yaml:"operationId" json:"operationId"`

	// A list of parameters that are applicable for this operation.
	// If a parameter is already defined at the Path Item, the new definition
	// will override it but can never remove it. The list MUST NOT include
	// duplicated parameters. A unique parameter is defined by a combination
	// of a name and location
	Parameters parameter.Parameters

	RequestBody *RequestBodyRef `yaml:"requestBody" json:"requestBody"`

	// The list of possible responses as they are returned from executing this
	// operation.
	Responses *Responses `yaml:"responses" json:"responses"`

	Endpoint *Endpoint `yaml:"-" json:"-"`
}

func (r EndpointsRef) Resolve(token string) (interface{}, error) {
	if v, ok := r.Value["/"+token]; ok {
		return v, nil
	}
	return nil, nil
}

func IsHttpStatusSuccess(status int) bool {
	return status == http.StatusOK ||
		status == http.StatusCreated ||
		status == http.StatusAccepted ||
		status == http.StatusNonAuthoritativeInfo ||
		status == http.StatusNoContent ||
		status == http.StatusResetContent ||
		status == http.StatusPartialContent ||
		status == http.StatusMultiStatus ||
		status == http.StatusAlreadyReported ||
		status == http.StatusIMUsed
}

type Responses struct {
	sortedmap.LinkedHashMap
} // map[HttpStatus]*ResponseRef

type ResponseRef struct {
	ref.Reference
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
	Content Content

	// Maps a header name to its definition. RFC7230 states header names are
	// case-insensitive. If a response header is defined with the name
	// "Content-Type", it SHALL be ignored.
	Headers map[string]*HeaderRef
}

type Content map[string]*MediaType

type MediaType struct {
	Schema   *schema.Ref
	Example  interface{}
	Examples map[string]*ExampleRef

	ContentType media.ContentType `yaml:"-" json:"-"`
}

type HeaderRef struct {
	ref.Reference
	Value *Header
}

type Header struct {
	Name        string
	Description string
	Schema      *schema.Ref
}

type Example struct {
	Summary     string
	Value       interface{}
	Description string
}

type ExampleRef struct {
	ref.Reference
	Value *Example
}

type Components struct {
	Schemas       *schema.SchemasRef         `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Responses     *NamedResponses            `yaml:"responses,omitempty" json:"responses,omitempty"`
	RequestBodies *RequestBodies             `yaml:"requestBodies,omitempty" json:"requestBodies,omitempty"`
	Parameters    *parameter.NamedParameters `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Examples      *Examples                  `yaml:"examples,omitempty" json:"examples,omitempty"`
	Headers       *NamedHeaders              `yaml:"headers,omitempty" json:"headers,omitempty"`
}

type NamedResponses struct {
	ref.Reference
	Value map[string]*ResponseRef
}

type NamedHeaders struct {
	ref.Reference
	Value map[string]*HeaderRef
}

type Examples struct {
	ref.Reference
	Value map[string]*ExampleRef
}

type RequestBodies struct {
	ref.Reference
	Value map[string]*RequestBodyRef
}

func (c *Config) Validate() error {
	if len(c.OpenApi) == 0 {
		return fmt.Errorf("no OpenApi version defined")
	}
	v := parseVersion(c.OpenApi)
	if v.major != 3 {
		return fmt.Errorf("unsupported version: %v", c.OpenApi)
	}

	if len(c.Info.Name) == 0 {
		return errors.New("an openapi title is required")
	}

	return nil
}

func (r *Responses) GetResponse(httpStatus int) *Response {
	i := r.Get(httpStatus)
	if i == nil {
		// 0 as default
		i = r.Get(0)
	}

	if i == nil {
		return nil
	}

	rr := i.(*ResponseRef)
	if rr == nil {
		return nil
	}
	return rr.Value
}

func (r *RequestBody) GetMedia(contentType media.ContentType) *MediaType {
	if c, ok := r.Content[contentType.String()]; ok {
		return c
	} else if c, ok := r.Content[contentType.Key()]; ok {
		return c
	}
	return nil
}

func (r *Response) GetContent(contentType media.ContentType) *MediaType {
	for _, v := range r.Content {
		if v.ContentType.Match(contentType) {
			return v
		}
	}

	return nil
}

func (e *Endpoint) Operations() map[string]*Operation {
	operations := make(map[string]*Operation, 7)
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

func (e *Endpoint) SetOperation(method string, o *Operation) {
	switch method {
	case http.MethodDelete:
		e.Delete = o
	case http.MethodGet:
		e.Get = o
	case http.MethodHead:
		e.Head = o
	case http.MethodOptions:
		e.Options = o
	case http.MethodPatch:
		e.Patch = o
	case http.MethodPost:
		e.Post = o
	case http.MethodPut:
		e.Put = o
	case http.MethodTrace:
		e.Trace = o
	default:
		panic(fmt.Errorf("unsupported HTTP method %q", method))
	}
}

func (op *Operation) getFirstSuccessResponse() (int, *Response, error) {
	var successStatus int
	for it := op.Responses.Iter(); it.Next(); {
		status := it.Key().(int)
		if IsHttpStatusSuccess(status) {
			successStatus = status
			break
		}
	}

	if successStatus == 0 {
		return 0, nil, fmt.Errorf("no success response (HTTP 2xx) in configuration")
	}

	r := op.Responses.GetResponse(successStatus)
	return successStatus, r, nil
}

func (op *Operation) getResponse(statusCode int) *Response {
	return op.Responses.GetResponse(statusCode)
}

type version struct {
	major int
	minor int
	build int
}

func parseVersion(s string) (v version) {
	numbers := strings.Split(s, ".")
	if len(numbers) == 0 {
		return
	}
	if len(numbers) > 0 {
		i, err := strconv.Atoi(numbers[0])
		if err != nil {
			return
		}
		v.major = i
	}
	if len(numbers) > 1 {
		i, err := strconv.Atoi(numbers[1])
		if err != nil {
			return
		}
		v.minor = i
	}
	if len(numbers) > 2 {
		i, err := strconv.Atoi(numbers[2])
		if err != nil {
			return
		}
		v.build = i
	}
	return
}
