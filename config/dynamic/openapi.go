package dynamic

import (
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type OpenApi struct {
	Info       Info
	Servers    []*Server
	EndPoints  map[string]*Endpoint `yaml:"paths"`
	Components Components
}

type Info struct {
	Name        string `yaml:"title"`
	Description string
	Version     string
	MokapiFile  string `yaml:"x-mokapifile"`
}

type Server struct {
	Url         string
	Description string
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

func (s *Server) GetPort() int {
	u, err := url.Parse(s.Url)
	if err != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return -1
	}
	portString := u.Port()
	if len(portString) == 0 {
		return 80
	} else {
		port, err := strconv.ParseInt(portString, 10, 32)
		if err != nil {
			log.WithField("url", s.Url).Error("Invalid port format in url found.")
		}
		return int(port)
	}
}

type Endpoint struct {
	Summary     string
	Description string
	Get         *Operation
	Post        *Operation
	Put         *Operation
	Patch       *Operation
	Delete      *Operation
	Head        *Operation
	Options     *Operation
	Trace       *Operation
	Parameters  []*Parameter
	Pipeline    string `yaml:"x-mokapi-pipeline"`
}

type Operation struct {
	Summary     string
	Description string
	OperationId string
	Parameters  []*Parameter
	RequestBody *RequestBody `yaml:"requestBody"`
	Responses   map[string]*Response
	Pipeline    string `yaml:"x-mokapi-pipeline"`
}

type Parameter struct {
	Name        string
	Type        string `yaml:"in"`
	Schema      *Schema
	Required    bool
	Description string
}

type Schema struct {
	Type                 string
	Format               string
	Reference            string `yaml:"$ref"`
	Description          string
	Properties           map[string]*Schema
	AdditionalProperties string `yaml:"additionalProperties"` // TODO custom marshal for bool, {} etc. Should it be a schema reference?
	Faker                string `yaml:"x-faker"`
	Items                *Schema
	Xml                  *Xml
	Required             []string
}

type RequestBody struct {
	Description string
	Content     map[string]*MediaType
	Required    bool
	Reference   string `yaml:"$ref"`
}

type Response struct {
	Description string
	Content     map[string]*MediaType
	Reference   string `yaml:"$ref"`
}

type MediaType struct {
	Schema *Schema
}

type Components struct {
	Schemas       map[string]*Schema
	Responses     map[string]*Response
	RequestBodies map[string]*RequestBody
}

type Xml struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
