package asyncApi

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mokapi"
	"net/url"
	"strconv"
)

func init() {
	dynamic.Register("asyncapi", &Config{}, func(path string, o dynamic.Config, r dynamic.ConfigReader) (bool, dynamic.Config) {
		eh := dynamic.NewEmptyEventHandler(o)
		switch c := o.(type) {
		case *Config:
			r := refResolver{reader: r, path: path, config: c, eh: eh}

			if err := r.resolveConfig(); err != nil {
				log.Errorf("error in resolving references in config %q: %v", path, err)
			}
		}
		return false, nil
	})
}

type Config struct {
	Info       Info
	Servers    map[string]Server
	Channels   map[string]*Channel
	Components Components
}

type Info struct {
	Name           string `yaml:"title" json:"title"`
	Description    string
	Version        string
	TermsOfService string
	Contact        *Contact
	License        *License
	Mokapi         *MokapiRef `yaml:"x-mokapi" json:"x-mokapi"`
}

type MokapiRef struct {
	Ref   string
	Value *mokapi.Config
}

type Contact struct {
	Name  string
	Url   string
	Email string
}

type License struct {
	Name string
	Url  string
}

type Server struct {
	Url             string
	Protocol        string
	ProtocolVersion string
	Description     string
	Bindings        Bindings
	Variables       map[string]*ServerVariable
}

type ServerVariable struct {
	Enum        []string
	Default     string
	Description string
	Examples    []string
}

type Channel struct {
	Description string
	Subscribe   *Operation
	Publish     *Operation
}

type Operation struct {
	Id          string `yaml:"operationId" json:"operationId"`
	Summary     string
	Description string
	Message     *Message
}
type Message struct {
	Name        string
	Title       string
	Summary     string
	Description string
	ContentType string
	Payload     *Schema
	Reference   string `yaml:"$ref" json:"$ref"`
}

type Schema struct {
	Type        string
	Reference   string `yaml:"$ref" json:"$ref"`
	Description string
	Properties  map[string]*Schema
	Items       *Schema
	Faker       string `yaml:"x-faker" json:"x-faker"`
}

type Components struct {
	Schemas  map[string]*Schema
	Messages map[string]*Message
}

type Bindings struct {
	Kafka KafkaBinding
}

func (s *Server) GetPort() int {
	u, err := url.Parse("//" + s.Url)
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

func (s *Server) GetHost() string {
	u, err := url.Parse("//" + s.Url)
	if err != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	return u.Hostname()
}
