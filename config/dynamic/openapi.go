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
	Name          string `yaml:"title"`
	Description   string
	Version       string
	DataProviders *DataProviders `yaml:"x-mokapi-resource"`
}

type ServerConfiguration struct {
	DataProviders *DataProviders `yaml:"data"`
}

type DataProviders struct {
	File *FileDataProvider
}

type FileDataProvider struct {
	Filename  string
	Directory string
}

type Server struct {
	Url         string
	Description string
}

func (s *Server) GetHost() string {
	u, error := url.Parse(s.Url)
	if error != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	return u.Hostname()
}

func (s *Server) GetPath() string {
	u, error := url.Parse(s.Url)
	if error != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	if len(u.Path) == 0 {
		return "/"
	}
	return u.Path
}

func (s *Server) GetPort() int {
	u, error := url.Parse(s.Url)
	if error != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return -1
	}
	portString := u.Port()
	if len(portString) == 0 {
		return 80
	} else {
		port, error := strconv.ParseInt(portString, 10, 32)
		if error != nil {
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
}

type Operation struct {
	Summary     string
	Description string
	OperationId string
	Parameters  []*Parameter
	Responses   map[string]*Response
	Middlewares []map[string]interface{} `yaml:"x-mokapi-middlewares"`
	Resources   []*Resource              `yaml:"x-mokapi-resources"`
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
}

type Resource struct {
	Name string
	If   string
}

type Response struct {
	Description string
	Content     map[string]*MediaType
}

type MediaType struct {
	Schema *Schema
}

type Components struct {
	Schemas map[string]*Schema
}

type Xml struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
