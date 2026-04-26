package events

import (
	"fmt"
	"mokapi/runtime/search"
	"reflect"
	"strings"
	"time"
)

type eventIndex struct {
	Api           string            `json:"api"`
	Type          string            `json:"type"`
	Discriminator string            `json:"discriminator"`
	Domain        string            `json:"domain"`
	Title         string            `json:"_title" index:"false" store:"true"`
	Event         *Event            `json:"event"`
	Metadata      map[string]string `json:"metadata"`
	Time          string            `json:"_time"`
}

func (m *StoreManager) addToIndex(event *Event) {
	if m.index == nil {
		return
	}

	data := eventIndex{
		Api:           event.Traits.GetName(),
		Type:          "event",
		Discriminator: fmt.Sprintf("event_%s", event.Traits.String()),
		Domain:        getDomainFromEvent(event),
		Event:         event,
		Time:          event.Time.Format(time.RFC3339),
	}
	if event.Data != nil {
		data.Title = event.Data.Title()

		for k, v := range event.Data.Metadata() {
			if data.Metadata == nil {
				data.Metadata = make(map[string]string)
			}
			data.Metadata[k] = v
		}
	}

	m.index.Add(event.Id, data)
}

func GetSearchResult(fields map[string]string, _ []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type:   "Event",
		Title:  fields["_title"],
		Domain: fields["domain"],
		Time:   fields["_time"],
	}
	result.Params = map[string]string{
		"id": fields["event.id"],
	}
	for k, v := range fields {
		if strings.HasPrefix(k, "event.traits.") {
			k = strings.Replace(k, "event.", "", 1)
			result.Params[k] = v
		} else if strings.HasPrefix(k, "metadata.") {
			k = strings.Replace(k, "metadata.", "", 1)
			result.Params[k] = v
		}
	}

	return result, nil
}

func getDomainFromEvent(event *Event) string {
	if d, ok := event.Data.(DomainData); ok {
		return d.Domain()
	}
	return getDataField(event, "Api")
}

type DomainData interface {
	Domain() string
}

func getDataField(event *Event, field string) string {
	if event.Data == nil {
		return ""
	}
	f := reflect.ValueOf(event.Data).Elem().FieldByName(field)
	if f.IsValid() {
		return f.String()
	}
	return ""
}
