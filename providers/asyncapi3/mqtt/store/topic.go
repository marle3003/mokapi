package store

import "sync"

type Message struct {
	Data   string
	QoS    byte
	Retain bool
}

type Topic struct {
	Name     string
	Clients  map[string]*Client // key clientId
	Retained *Message
	m        sync.RWMutex
}

func (t *Topic) AddMessage(msg *Message) {
	if msg.Retain {

		if msg.Data == "" {
			t.Retained = nil
			return
		} else {
			t.Retained = msg
		}
	}
}
