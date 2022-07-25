package events

import (
	"fmt"
	"github.com/google/uuid"
	"sort"
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

func GetEvents(traits Traits) []Event {
	events := make([]Event, 0)

	for _, s := range stores {
		if len(traits) == 0 || s.traits.Match(traits) {
			events = append(events, s.Events(traits)...)
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Time.After(events[j].Time)
	})

	return events
}

func GetEvent(id string) Event {
	for _, s := range stores {
		for _, e := range s.events {
			if e.Id == id {
				return e
			}
		}
	}
	return Event{}
}

func Reset() {
	stores = make([]*store, 0)
}

func (e *Event) IsValid() bool {
	return len(e.Id) > 0
}
