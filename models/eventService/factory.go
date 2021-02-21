package event

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/asyncApi"
	"net/url"
	"strconv"
)

type context struct {
	service    *Service
	unresolved map[string]*asyncApi.Schema
	path       string
	error      func(msg string)
}

func (s *Service) Apply(config *asyncApi.Config, filePath string) {
	ctx := &context{
		service: s,
		error: func(msg string) {
			log.Errorf("error in config %v: %v", filePath, msg)
			s.Errors = append(s.Errors, msg)
		},
		path: filePath,
	}

	s.Name = config.Info.Name
	s.Description = config.Info.Description
	s.Version = config.Info.Version

	for _, c := range config.Servers {
		s.Servers = append(s.Servers, newServer(c, ctx))
	}

	for n, s := range config.Components.Schemas {
		buildSchemaFromComponents(n, s, ctx)
	}

	for n, m := range config.Components.Messages {
		buildMessageFromComponents(n, m, ctx)
	}

	for n, c := range config.Channels {
		s.Channels[n] = createChannel(c, ctx)
	}
}

func newServer(config asyncApi.Server, ctx *context) Server {
	s := Server{
		Host: getHost(config.Url, ctx),
		Port: getPort(config.Url, ctx),
	}

	switch config.Protocol {
	case "kafka":
		s.Type = Kafka
	}

	return s
}

func createChannel(config *asyncApi.Channel, ctx *context) *Channel {
	c := &Channel{
		Description: config.Description,
	}

	if config.Subscribe != nil {
		c.Subscribe = createOperation(config.Subscribe, ctx)
	}
	if config.Publish != nil {
		c.Publish = createOperation(config.Publish, ctx)
	}

	return c
}

func createOperation(config *asyncApi.Operation, ctx *context) *Operation {
	return &Operation{
		Id:          config.Id,
		Summary:     config.Summary,
		Description: config.Description,
		Message:     createMessage(config.Message, ctx),
	}
}

func getHost(s string, ctx *context) string {
	u, error := url.Parse(s)
	if error != nil {
		ctx.error(fmt.Sprintf("invalid format in url found: %v", s))
		return ""
	}
	return u.Hostname()
}

func getPort(s string, ctx *context) int {
	u, err := url.Parse(s)
	if err != nil {
		ctx.error(fmt.Sprintf("invalid format in url found: %v", s))
	}
	portString := u.Port()
	if len(portString) == 0 {
		return 9092
	} else {
		port, err := strconv.ParseInt(portString, 10, 32)
		if err != nil {
			ctx.error(fmt.Sprintf("invalid port format in url found: %v", err.Error()))
		}
		return int(port)
	}
}
