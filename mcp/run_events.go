package mcp

import (
	"fmt"
	"mokapi/js/util"
	"mokapi/runtime/events"
	"reflect"
	"strings"

	"github.com/dop251/goja"
)

func (m *mokapi) getEvents(vTraits goja.Value, vLimit goja.Value) ([]events.Event, error) {
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
	if len(evts) > limit {
		return evts[0:limit], nil
	} else {
		return evts, nil
	}
}

func (m *mokapi) getEvent(id string) (events.Event, error) {
	if id == "" {
		return events.Event{}, fmt.Errorf("expected id parameter in GUID format, got '%v'", id)
	}

	e := m.app.Events.GetEvent(id)
	if e.Id == "" {
		return e, fmt.Errorf("event %s not found. Use `mokapi.search('type:event ...')` to search for existing events", id)
	}
	return e, nil
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
