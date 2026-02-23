package events

import (
	"fmt"
	"mokapi/runtime/search"
	"reflect"
	"time"
)

type eventIndex struct {
	Type          string `json:"type"`
	Discriminator string `json:"discriminator"`
	Api           string `json:"api"`
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
		Api:           getApiFromEvent(event),
		Event:         event,
		Time:          event.Time.Format(time.RFC3339),
	}
	if event.Data != nil {
		data.Title = event.Data.Title()
	}

	m.index.Add(event.Id, data)
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

func getApiFromEvent(event *Event) string {
	if event.Data == nil {
		return ""
	}
	f := reflect.ValueOf(event.Data).Elem().FieldByName("Api")
	if f.IsValid() {
		return f.String()
	}
	return ""
}
