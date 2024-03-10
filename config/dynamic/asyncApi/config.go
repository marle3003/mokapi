package asyncApi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/providers/openapi/schema"
	"mokapi/version"
	"net/url"
	"strconv"
)

var supportedVersions = []*version.Version{
	version.New("2.0.0"),
	version.New("2.1.0"),
	version.New("2.2.0"),
	version.New("2.3.0"),
}

type Config struct {
	AsyncApi   string                 `yaml:"asyncapi" json:"asyncapi"`
	Info       Info                   `yaml:"info" json:"info"`
	Servers    map[string]*ServerRef  `yaml:"servers,omitempty" json:"servers,omitempty"`
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
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Url   string `yaml:"url,omitempty" json:"url,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type License struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
}

type Server struct {
	Url             string         `yaml:"url" json:"url"`
	Protocol        string         `yaml:"protocol" json:"protocol"`
	ProtocolVersion string         `yaml:"protocolVersion" json:"protocolVersion"`
	Description     string         `yaml:"description" json:"description"`
	Bindings        ServerBindings `yaml:"bindings" json:"bindings"`
}

type ServerRef struct {
	Ref   string
	Value *Server
}

type ServerBindings struct {
	Kafka kafka.BrokerBindings `yaml:"kafka" json:"kafka"`
}

type ChannelRef struct {
	Ref   string
	Value *Channel
}

type Channel struct {
	Description string          `yaml:"description" json:"description"`
	Subscribe   *Operation      `yaml:"subscribe" json:"subscribe"`
	Publish     *Operation      `yaml:"publish" json:"publish"`
	Bindings    ChannelBindings `yaml:"bindings" json:"bindings"`
	Servers     []string        `yaml:"servers" json:"servers"`
}

type ChannelBindings struct {
	Kafka kafka.TopicBindings `yaml:"kafka" json:"kafka"`
}

type Operation struct {
	OperationId string            `yaml:"operationId" json:"operationId"`
	Summary     string            `yaml:"summary" json:"summary"`
	Description string            `yaml:"description" json:"description"`
	Message     *MessageRef       `yaml:"message" json:"message"`
	Bindings    OperationBindings `yaml:"bindings" json:"bindings"`
}

type OperationBindings struct {
	Kafka kafka.Operation `yaml:"kafka" json:"kafka"`
}

type MessageRef struct {
	Ref   string
	Value *Message
}

type Message struct {
	Name        string         `yaml:"name" json:"name"`
	Title       string         `yaml:"title" json:"title"`
	Summary     string         `yaml:"summary" json:"summary"`
	Description string         `yaml:"description" json:"description"`
	ContentType string         `yaml:"contentType" json:"contentType"`
	Payload     *schema.Ref    `yaml:"payload" json:"payload"`
	Bindings    MessageBinding `yaml:"bindings" json:"bindings"`
	Headers     *schema.Ref    `yaml:"headers" json:"headers"`
}

type MessageBinding struct {
	Kafka kafka.MessageBinding `yaml:"kafka" json:"kafka"`
}

type Components struct {
	Servers  map[string]*Server  `yaml:"servers" json:"servers"`
	Channels map[string]*Channel `yaml:"channels" json:"channels"`
	Schemas  *schema.Schemas     `yaml:"schemas" json:"schemas"`
	Messages map[string]*Message `yaml:"messages" json:"messages"`
}

func (c *Config) Validate() error {
	if len(c.AsyncApi) == 0 {
		return fmt.Errorf("no version defined")
	}

	v := version.New(c.AsyncApi)
	if !v.IsSupported(supportedVersions...) {
		return fmt.Errorf("not supported version: %v", c.AsyncApi)
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
