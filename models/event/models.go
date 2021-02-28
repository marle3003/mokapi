package event

import (
	"mokapi/models/media"
	"mokapi/models/schemas"
)

type Service struct {
	Name        string
	Description string
	Version     string
	Servers     []Server
	Channels    map[string]*Channel
	Messages    map[string]*Message
	Models      map[string]*schemas.Schema
	Errors      []string
}

func NewService() *Service {
	return &Service{
		Channels: make(map[string]*Channel),
		Messages: make(map[string]*Message),
		Models:   make(map[string]*schemas.Schema),
	}
}

type Protocol int

const (
	Kafka Protocol = 1
)

func (p Protocol) String() string {
	switch p {
	case Kafka:
		return "kafka"
	}

	return "unknown protocol"
}

type Server struct {
	Host        string
	Port        int
	Description string

	Type Protocol
}

type Channel struct {
	Description string
	Subscribe   *Operation
	Publish     *Operation
}

type Operation struct {
	Id          string
	Summary     string
	Description string
	Message     *Message
}

type Message struct {
	Name        string
	Title       string
	Summary     string
	Description string
	ContentType *media.ContentType
	Payload     *schemas.Schema
	Reference   string
	isResolved  bool
}
