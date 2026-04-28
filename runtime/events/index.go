package events

import (
	"fmt"
	"mokapi/runtime/search"
	"reflect"
	"strings"
	"time"
)

type IndexFieldsProvider interface {
	IndexFields() map[string]any
}

func (m *StoreManager) addToIndex(event *Event) {
	if m.index == nil {
		return
	}

	data := map[string]any{
		"api":           event.Traits.GetName(),
		"type":          "event",
		"discriminator": fmt.Sprintf("event_%s", event.Traits.String()),
		"domain":        getDomainFromEvent(event),
		"id":            event.Id,
		"traits":        event.Traits,
		"_time":         event.Time.Format(time.RFC3339),
	}

	if event.Data != nil {
		data["_title"] = event.Data.Title()

		if p, ok := event.Data.(IndexFieldsProvider); ok {
			for k, v := range p.IndexFields() {
				data[k] = v
			}
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
		"type": "event",
		"id":   fields["id"],
	}
	for k, v := range fields {
		if strings.HasPrefix(k, "traits.") {
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
	return event.Traits.GetName()
}
