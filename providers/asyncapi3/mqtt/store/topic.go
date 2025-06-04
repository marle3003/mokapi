package store

import (
	"mokapi/providers/asyncapi3"
	"sync"
)

type Message struct {
	Topic  string
	Data   []byte
	QoS    byte
	Retain bool
}

type Topic struct {
	Name     string
	Retained *Message

	cfg *asyncapi3.Channel
	m   sync.RWMutex
}
