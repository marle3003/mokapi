package models

import (
	"fmt"
	"mokapi/providers/parser"
	"strings"
)

type ServiceList []*Service

type Service struct {
	Name          string
	Description   string
	Version       string
	Servers       []Server
	Endpoint      map[string]*Endpoint
	DataProviders DataProviders
}

type DataProviders struct {
	File *FileDataProvider
}

type FileDataProvider struct {
	Path string
}

type Server struct {
	Host        string
	Port        int
	Path        string
	Description string
}

type Endpoint struct {
	Path       string
	Get        *Operation
	Post       *Operation
	Put        *Operation
	Patch      *Operation
	Delete     *Operation
	Head       *Operation
	Options    *Operation
	Trace      *Operation
	Parameters []*Parameter
}

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
	Summary     string
	Description string
	OperationId string
	Parameters  []*Parameter
	Responses   map[HttpStatus]*Response
	Middleware  *Middleware
	Resources   []*Resource
}

type HttpStatus int

const (
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

func IsValidHttpStatus(status HttpStatus) bool {
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
	Name        string
	Type        ParameterType
	Schema      *Schema
	Required    bool
	Description string
}

type ParameterType int

const (
	PathParameter   ParameterType = 1
	QueryParameter  ParameterType = 2
	HeaderParameter ParameterType = 3
	CookieParameter ParameterType = 4
)

func (p ParameterType) String() string {
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

type Schema struct {
	Type                 string
	Format               string
	Description          string
	Properties           map[string]*Schema
	Faker                string
	Items                *Schema
	Xml                  *XmlEncoding
	AdditionalProperties string
}

type Resource struct {
	If   *parser.FilterExp
	Name string
}

type Response struct {
	Description  string
	ContentTypes map[ContentType]*ResponseContent
}

type ResponseContent struct {
	Schema *Schema
}

type ContentType string

// do we need that as type? Change to simple string?
const (
	Json      ContentType = "application/json"
	Rss       ContentType = "application/rss+xml"
	JsonOData ContentType = "application/json;odata=verbose"
	TextXml   ContentType = "text/xml"
)

func (c ContentType) String() string {
	return string(c)
}

func ParseContentType(s string) (ContentType, error) {
	c := ContentType(s)
	switch c {
	case Json, Rss, JsonOData, TextXml:
		return c, nil
	default:
		return c, fmt.Errorf("Unknown content type %v", s)
	}
}

type XmlEncoding struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
