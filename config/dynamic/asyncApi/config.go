package asyncApi

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/openapi"
	"net/url"
	"strconv"
)

func init() {
	dynamic.Register("asyncapi", &Config{}, func(path string, o dynamic.Config, r dynamic.ConfigReader) (bool, dynamic.Config) {
		eh := dynamic.NewEmptyEventHandler(o)
		switch c := o.(type) {
		case *Config:
			c.ConfigPath = path
			r := refResolver{reader: r, path: path, config: c, eh: eh}

			if err := r.resolveConfig(); err != nil {
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
	Servers    map[string]Server
	Channels   map[string]*ChannelRef
	Components Components
}

type Info struct {
	Name           string `yaml:"title" json:"title"`
	Description    string
	Version        string
	TermsOfService string
	Contact        *Contact
	License        *License
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
	Bindings        ServerBindings
}

type ServerBindings struct {
	Kafka Kafka
}

type ChannelRef struct {
	Ref   string
	Value *Channel
}

type Channel struct {
	Description string
	Subscribe   *Operation
	Publish     *Operation
	Bindings    ChannelBindings
}

type ChannelBindings struct {
	Kafka KafkaChannelBinding
}

type Operation struct {
	Id          string `yaml:"operationId" json:"operationId"`
	Summary     string
	Description string
	Message     *MessageRef
}

type MessageRef struct {
	Ref   string
	Value *Message
}

type Message struct {
	Name        string
	Title       string
	Summary     string
	Description string
	ContentType string
	Payload     *openapi.SchemaRef
	Bindings    MessageBinding
}

type MessageBinding struct {
	Kafka KafkaMessageBinding
}

type Components struct {
	Schemas  *openapi.Schemas
	Messages map[string]*Message
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
