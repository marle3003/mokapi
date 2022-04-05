package asyncApi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/schema"
	"net/url"
	"strconv"
)

func init() {
	common.Register("asyncapi", &Config{})
}

type Config struct {
	AsyncApi   string                 `yaml:"asyncapi" json:"asyncapi"`
	Info       Info                   `yaml:"info" json:"info"`
	Servers    map[string]Server      `yaml:"servers,omitempty" json:"servers,omitempty"`
	Channels   map[string]*ChannelRef `yaml:"channels" json:"channels"`
	Components *Components            `yaml:"components,omitempty" json:"components,omitempty"`
}

type Info struct {
	Name           string   `yaml:"title" json:"title"`
	Description    string   `yaml:"description,omitempty" json:"description,omitempty"`
	Version        string   `yaml:"version" json:"version"`
	TermsOfService string   `yaml:"termsOfService,omitempty" json:"termsOfService,omitempty"`
	Contact        *Contact `yaml:"contact,omitempty" json:"contact,omitempty"`
	License        *License `yaml:"license,omitempty" json:"license,omitempty"`
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
	Payload     *schema.Ref
	Bindings    MessageBinding
	Headers     *schema.Ref
}

type MessageBinding struct {
	Kafka kafka.MessageBinding
}

type Components struct {
	Schemas  *schema.Schemas
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
