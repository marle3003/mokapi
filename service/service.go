package service

import (
	"fmt"
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
	Get     *Operation
	Post    *Operation
	Put     *Operation
	Patch   *Operation
	Delete  *Operation
	Head    *Operation
	Options *Operation
	Trace   *Operation
}

type Operation struct {
	Summary     string
	Description string
	OperationId string
	Parameters  []*Parameter
	Responses   map[HttpStatus]*Response
}

type HttpStatus int

const (
	Ok      HttpStatus = 200
	Created HttpStatus = 201
)

func IsValidHttpStatus(status HttpStatus) bool {
	switch status {
	case Ok, Created:
		return true
	default:
		return false
	}
}

type Parameter struct {
	Name        string
	Type        string
	Schema      *Schema
	Required    bool
	Description string
}

type Schema struct {
	Type        string
	Format      string
	Description string
	Properties  map[string]*Schema
	Faker       string
	Resource    string
	Items       *Schema
	Xml         *XmlEncoding
}

type Response struct {
	Description  string
	ContentTypes map[ContentType]*ResponseContent
}

type ResponseContent struct {
	Schema *Schema
}

type ContentType string

const (
	Json      ContentType = "application/json"
	Rss       ContentType = "application/rss+xml"
	JsonOData ContentType = "application/json;odata=verbose"
)

func (c ContentType) String() string {
	return string(c)
}

func ParseContentType(s string) (ContentType, error) {
	c := ContentType(s)
	switch c {
	case Json, Rss, JsonOData:
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
