package dynamic

import (
	"net/url"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type OpenApi struct {
	Parts map[string]*OpenApiPart
}

type OpenApiPart struct {
	Info       Info
	Servers    []*Server
	EndPoints  map[string]*Endpoint `yaml:"paths"`
	Components Components
}

type Info struct {
	Name                string `yaml:"title"`
	Description         string
	Version             string
	ServerConfiguration ServerConfiguration `yaml:"x-server-configuration"`
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

func (p *FileDataProvider) UpdatePath(path string) {
	if !filepath.IsAbs(p.Filename) {
		p.Filename = filepath.Join(path, p.Filename)
	}
	if !filepath.IsAbs(p.Directory) {
		p.Directory = filepath.Join(path, p.Directory)
	}

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
	Responses   map[string]*Response
}

type Parameter struct {
	Name        string
	Type        string `yaml:"in"`
	Schema      *Schema
	Required    bool
	Description string
}

type Schema struct {
	Type        string
	Format      string
	Reference   string `yaml:"$ref"`
	Description string
	Properties  map[string]*Schema
	Faker       string `yaml:"x-faker"`
	Resource    string `yaml:"x-resource"`
	Items       *Schema
	Xml         *Xml
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
