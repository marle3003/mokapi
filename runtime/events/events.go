package events

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

type Handler interface {
	Push(data EventData, traits Traits) error
	GetEvents(traits Traits) []Event
}

type EventData interface {
	Title() string
}

type StoreManager struct {
	stores []*store
	index  bleve.Index
	m      sync.RWMutex
}

func NewStoreManager(index bleve.Index) *StoreManager {
	return &StoreManager{index: index}
}

func (m *StoreManager) Push(data EventData, traits Traits) error {
	m.m.RLock()
	defer m.m.RUnlock()

	if len(traits) == 0 {
		return fmt.Errorf("empty traits not allowed")
	}
	score := 0
	var bestStore *store
	for _, s := range m.stores {
		if s.traits.Match(traits) && len(s.traits) > score {
			bestStore = s
			score = len(s.traits)
		}
	}

	if bestStore == nil {
		return fmt.Errorf("no store found for %s", traits)
	}

	evt := NewEvent(data, traits)

	m.addToIndex(&evt)

	removed := bestStore.Push(evt)
	if removed != nil {
		m.removeFromIndex(removed)
	}

	return nil
}

func (m *StoreManager) SetStore(size int, traits Traits) {
	m.stores = append(m.stores, &store{
		size:   size,
		traits: traits,
	})
}

func (m *StoreManager) GetStores(traits Traits) []StoreInfo {
	m.m.RLock()
	defer m.m.RUnlock()

	var result []StoreInfo
	for _, s := range m.stores {
		if s.traits.Match(traits) {
			result = append(result, StoreInfo{
				Traits:    s.traits,
				Size:      s.size,
				NumEvents: len(s.events),
			})
		}
	}
	return result
}

func (m *StoreManager) GetEvents(traits Traits) []Event {
	m.m.RLock()
	defer m.m.RUnlock()

	events := make([]Event, 0)

	for _, s := range m.stores {
		if len(traits) == 0 || traits.Match(s.traits) || s.traits.Match(traits) {
			events = append(events, s.Events(traits)...)
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Time.After(events[j].Time)
	})

	return events
}

func (m *StoreManager) GetEvent(id string) Event {
	m.m.RLock()
	defer m.m.RUnlock()

	for _, s := range m.stores {
		for _, e := range s.events {
			if e.Id == id {
				return e
			}
		}
	}
	return Event{}
}

func (m *StoreManager) ResetStores(traits Traits) {
	m.m.Lock()
	defer m.m.Unlock()

	i := 0 // output index
	for _, s := range m.stores {
		if !traits.Match(s.traits) {
			// copy and increment index
			m.stores[i] = s
			i++
		} else {
			log.Debugf("reset store %s", traits.String())
		}
	}
	// Prevent memory leak by erasing truncated values
	for j := i; j < len(m.stores); j++ {
		m.stores[j] = nil
	}
	m.stores = m.stores[:i]
}

type Event struct {
	Id     string    `json:"id"`
	Traits Traits    `json:"traits"`
	Data   EventData `json:"data"`
	Time   time.Time `json:"time"`
}

type StoreInfo struct {
	Traits    Traits  `json:"traits"`
	Events    []Event `json:"events,omitempty"`
	Size      int     `json:"size"`
	NumEvents int     `json:"numEvents"`
}

func (e *Event) IsValid() bool {
	return len(e.Id) > 0
}

func NewEvent(data EventData, traits Traits) Event {
	return Event{
		Id:     uuid.New().String(),
		Traits: traits,
		Data:   data,
		Time:   time.Now(),
	}
}
