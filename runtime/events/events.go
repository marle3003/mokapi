package events

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

var stores []*store

type Event struct {
	Id     string      `json:"id"`
	Traits Traits      `json:"traits"`
	Data   interface{} `json:"data"`
	Time   time.Time   `json:"time"`
}

func SetStore(size int, traits Traits) {
	stores = append(stores, &store{
		size:   size,
		traits: traits,
	})
}

func Push(data interface{}, traits Traits) error {
	if len(traits) == 0 {
		return fmt.Errorf("empty traits not allowed")
	}
	for _, s := range stores {
		if s.traits.Match(traits) {
			s.Push(Event{
				Id:     uuid.New().String(),
				Traits: traits,
				Data:   data,
				Time:   time.Now(),
			})
			return nil
		}
	}
	return fmt.Errorf("no store found for %s", traits)
}

func Events(traits Traits) []Event {
	events := make([]Event, 0)
	for _, s := range stores {
		if len(traits) == 0 || s.traits.Match(traits) {
			events = append(events, s.Events(traits)...)
		}
	}
	return events
}

func Reset() {
	stores = make([]*store, 0)
}
