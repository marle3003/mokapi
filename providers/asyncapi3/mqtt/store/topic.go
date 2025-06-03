package store

import (
	"mokapi/providers/asyncapi3"
	"sync"
)

type Message struct {
	Data []byte
	QoS  byte
}

type Topic struct {
	Name     string
	Retained *Message

	cfg *asyncapi3.Channel
	m   sync.RWMutex
}
