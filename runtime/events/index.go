package events

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/runtime/search"
	"strings"
)

type eventIndex struct {
	Type          string `json:"type"`
	Discriminator string `json:"discriminator"`
	Title         string `json:"_title" index:"false" store:"true"`
	Event         *Event `json:"event"`
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
	}

	if err := m.index.Index(event.Id, data); err != nil {
		log.Errorf("add '%s' to search index failed: %v", event.Id, err)
	}
}

func (m *StoreManager) removeFromIndex(event *Event) {
	_ = m.index.Delete(event.Id)
}

func GetSearchResult(fields map[string]string, _ []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type:   "Event",
		Title:  fields["_title"],
		Domain: fields["event.data.api"],
	}
	result.Params = map[string]string{
		"type": strings.ToLower(result.Type),
		"id":   fields["event.id"],
	}

	return result, nil
}
