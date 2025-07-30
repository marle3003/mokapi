package events

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/runtime/search"
	"time"
)

type eventIndex struct {
	Type          string `json:"type"`
	Discriminator string `json:"discriminator"`
	Title         string `json:"_title" index:"false" store:"true"`
	Event         *Event `json:"event"`
	Time          string `json:"_time"`
}

func (m *StoreManager) addToIndex(event *Event) {
	if m.index == nil {
		return
	}

	data := eventIndex{
		Type:          "event",
		Discriminator: fmt.Sprintf("event_%s", event.Traits.String()),
		Event:         event,
		Title:         event.Data.Title(),
		Time:          event.Time.Format(time.RFC3339),
	}

	if err := m.index.Index(event.Id, data); err != nil {
		log.Errorf("add '%s' to search index failed: %v", event.Id, err)
	}
}

func (m *StoreManager) removeFromIndex(event *Event) {
	if m.index == nil {
		return
	}
	_ = m.index.Delete(event.Id)
}

func GetSearchResult(fields map[string]string, _ []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type:   "Event",
		Title:  fields["_title"],
		Domain: fields["event.data.api"],
		Time:   fields["_time"],
	}
	if fields["event.traits.namespace"] == "kafka" {
		result.Domain = fmt.Sprintf("%s - %s", fields["event.data.api"], fields["event.traits.topic"])
	}
	result.Params = map[string]string{
		"namespace": fields["event.traits.namespace"],
		"id":        fields["event.id"],
	}

	return result, nil
}
