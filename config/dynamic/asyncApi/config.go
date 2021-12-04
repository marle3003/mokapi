package asyncApi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"net/url"
	"strconv"
)

func init() {
	common.Register("asyncapi", &Config{})
}

//func init() {
//	dynamic.Register("asyncapi", &Config{}, func(o *dynamic.Config, r dynamic.ConfigReader) bool {
//		eh := dynamic.NewEmptyEventHandler()
//		switch c := o.Data.(type) {
//		case *Config:
//			r := ReferenceResolver{reader: r, url: o.Url, config: c, eh: eh}
//
//			if err := r.ResolveConfig(); err != nil {
//				log.Errorf("error in resolving references in config %q: %v", o.Url.String(), err)
//			}
//
//			return true
//		}
//		return false
//	})
//}

type Config struct {
	AsyncApi   string
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
	Kafka kafka.BrokerBindings
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
	Kafka kafka.TopicBindings
}

type Operation struct {
	Id          string `yaml:"operationId" json:"operationId"`
	Summary     string
	Description string
	Message     *MessageRef
	Bindings    OperationBindings
}

type OperationBindings struct {
	Kafka kafka.Operation
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
	Headers     *openapi.SchemaRef
}

type MessageBinding struct {
	Kafka KafkaMessageBinding
}

type Components struct {
	Schemas  *openapi.Schemas
	Messages map[string]*Message
}

func (c *Config) Validate() error {
	if len(c.AsyncApi) == 0 {
		return fmt.Errorf("no version defined")
	}
	if c.AsyncApi != "2.0.0" {
		return fmt.Errorf("unsupported version: %v", c.AsyncApi)
	}
	return nil
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
