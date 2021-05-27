package event

import (
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/workflow/runtime"
	"strings"
)

type WorkflowHandler func(events EventHandler, options ...runtime.WorkflowOptions)

type EventHandler func(trigger mokapi.Trigger) bool

type HttpEvent struct {
	Service string
	Method  string
	Path    string
}

func WithHttpEvent(evt HttpEvent) EventHandler {
	return func(t mokapi.Trigger) bool {
		if len(t.Service) > 0 && t.Service != evt.Service {
			return false
		}
		if evt.isValid(t.Http) {
			return true
		}

		return false
	}
}

func (e HttpEvent) isValid(t mokapi.HttpTrigger) bool {
	if len(t.Method) > 0 && t.Method != strings.ToLower(e.Method) {
		return false
	}
	if len(t.Path) > 0 && t.Path != strings.ToLower(e.Path) {
		return false
	}

	return true
}
