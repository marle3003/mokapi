package mcp

import (
	"fmt"
	"mokapi/js/util"
	"mokapi/providers/asyncapi3/kafka/store"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"reflect"
	"strings"
	"time"

	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Id     string        `json:"id"`
	Type   string        `json:"type"`
	Time   time.Time     `json:"time"`
	Traits events.Traits `json:"traits"`
}

type HttpEvent struct {
	Event
	Api        string                   `json:"api"`
	Path       string                   `json:"path"`
	Method     string                   `json:"method"`
	StatusCode int                      `json:"statusCode"`
	Request    *openapi.HttpRequestLog  `json:"request"`
	Response   *openapi.HttpResponseLog `json:"response"`
}

type KafkaEvent struct {
	Event
	Api       string `json:"api"`
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Key       string `json:"key"`
	Message   string `json:"message"`
}

type LogEvent struct {
	Event
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (m *mokapi) getEvents(vTraits goja.Value, vLimit goja.Value) ([]any, error) {
	traits, err := parseTraits(vTraits, m.vm)
	if err != nil {
		return nil, err
	}

	evts := m.app.Events.GetEvents(traits)

	limit := 10
	if vLimit != nil {
		if vLimit.ExportType().Kind() != reflect.Float64 {
			return nil, fmt.Errorf("unexpected type for apiType: %s", util.JsType(vLimit.ExportType()))
		}
		limit = int(vLimit.ToInteger())
	}
	var result []any
	for i, evt := range evts {
		if i >= limit {
			break
		}
		e := convertEvent(evt)
		result = append(result, e)
	}

	return result, nil
}

func (m *mokapi) getEvent(id string) (any, error) {
	if id == "" {
		return events.Event{}, fmt.Errorf("expected id parameter in GUID format, got '%v'", id)
	}

	e := m.app.Events.GetEvent(id)
	if e.Id == "" {
		return e, fmt.Errorf("event %s not found. Use `mokapi.search('type:event ...')` to search for existing events", id)
	}
	return convertEvent(e), nil
}

func parseTraits(v goja.Value, vm *goja.Runtime) (events.Traits, error) {
	traits := events.Traits{}

	if v == nil {
		return traits, nil
	}

	if v.ExportType().Kind() != reflect.Map {
		return nil, fmt.Errorf("expect object but got: %v", util.JsType(v.Export()))
	}

	obj := v.ToObject(vm)
	for _, k := range obj.Keys() {
		switch k {
		case "type":
			val := obj.Get(k)
			if val.ExportType().Kind() != reflect.String {
				return nil, fmt.Errorf("unexpected type for type: %s", util.JsType(val.ExportType()))
			}
			traits.WithNamespace(val.String())
		case "name":
			val := obj.Get(k)
			if val.ExportType().Kind() != reflect.String {
				return nil, fmt.Errorf("unexpected type for name: %s", util.JsType(val.ExportType()))
			}
			traits.WithName(val.String())
		case "method":
			val := obj.Get(k)
			if val.ExportType().Kind() != reflect.String {
				return nil, fmt.Errorf("unexpected type for method: %s", util.JsType(val.ExportType()))
			}
			traits.With("method", strings.ToUpper(val.String()))
		default:
			val := obj.Get(k)
			if val.ExportType().Kind() != reflect.String {
				return nil, fmt.Errorf("unexpected type for %s: %s", k, util.JsType(val.ExportType()))
			}
			traits.With(k, val.String())
		}
	}

	return traits, nil
}

func convertEvent(evt events.Event) any {
	var item any
	switch e := evt.Data.(type) {
	case *openapi.HttpLog:
		httpEvent := &HttpEvent{
			Event: Event{
				Id:   evt.Id,
				Type: "http",
				Time: evt.Time,
			},
			Api:      e.Api,
			Path:     e.Path,
			Request:  e.Request,
			Response: e.Response,
		}
		if e.Request != nil {
			httpEvent.Method = e.Request.Method
		}
		if e.Response != nil {
			httpEvent.StatusCode = e.Response.StatusCode
		}
		item = httpEvent
	case *store.KafkaMessageLog:
		kafkaEvent := &KafkaEvent{
			Event: Event{
				Id:     evt.Id,
				Type:   "kafka",
				Time:   evt.Time,
				Traits: evt.Traits,
			},
			Api:       e.Api,
			Topic:     evt.Traits.Get("topic"),
			Partition: e.Partition,
			Offset:    e.Offset,
		}
		if e.Key.Value != "" {
			kafkaEvent.Key = e.Key.Value
		} else {
			kafkaEvent.Key = string(e.Key.Binary)
		}
		if e.Message.Value != "" {
			kafkaEvent.Message = e.Message.Value
		} else {
			kafkaEvent.Message = string(e.Message.Value)
		}
		item = kafkaEvent
	case *runtime.LogData:
		item = &LogEvent{
			Event: Event{
				Id:     evt.Id,
				Type:   "log",
				Time:   evt.Time,
				Traits: evt.Traits,
			},
			Level:   e.Level,
			Message: e.Message,
		}
	default:
		log.Errorf("mcp: event type %T not supported", e)
		item = &Event{
			Id:     evt.Id,
			Type:   evt.Traits.GetNamespace(),
			Time:   evt.Time,
			Traits: evt.Traits,
		}
	}
	return item
}
