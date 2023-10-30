package events

import (
	"fmt"
	"github.com/google/uuid"
	"sort"
	"sync"
	"time"
)

var (
	stores []*store
	m      sync.Mutex
)

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
	score := 0
	var bestStore *store
	for _, s := range stores {
		if s.traits.Match(traits) && len(s.traits) > score {
			bestStore = s
			score = len(s.traits)
		}
	}

	if bestStore == nil {
		return fmt.Errorf("no store found for %s", traits)
	}

	bestStore.Push(Event{
		Id:     uuid.New().String(),
		Traits: traits,
		Data:   data,
		Time:   time.Now(),
	})
	return nil
}

func GetEvents(traits Traits) []Event {
	events := make([]Event, 0)

	for _, s := range stores {
		if len(traits) == 0 || traits.Match(s.traits) || s.traits.Match(traits) {
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
	m.Lock()
	defer m.Unlock()

	stores = make([]*store, 0)
}

func ResetStores(traits Traits) {
	m.Lock()
	defer m.Unlock()

	i := 0 // output index
	for _, s := range stores {
		if !traits.Match(s.traits) {
			// copy and increment index
			stores[i] = s
			i++
		}
	}
	// Prevent memory leak by erasing truncated values
	for j := i; j < len(stores); j++ {
		stores[j] = nil
	}
	stores = stores[:i]
}

func (e *Event) IsValid() bool {
	return len(e.Id) > 0
}
