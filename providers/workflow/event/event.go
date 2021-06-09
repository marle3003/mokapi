package event

import (
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/workflow/path/urlpath"
	"strings"
)

type Handler func(trigger mokapi.Trigger) bool

type HttpEvent struct {
	Service string
	Method  string
	Path    string
}

func WithHttpEvent(evt HttpEvent) Handler {
	return func(t mokapi.Trigger) bool {
		if len(t.Service) > 0 && !strings.EqualFold(t.Service, evt.Service) {
			return false
		}
		if evt.isValid(t.Http) {
			return true
		}

		return false
	}
}

func (e HttpEvent) isValid(t *mokapi.HttpTrigger) bool {
	if t == nil {
		return false
	}
	if len(t.Method) > 0 && t.Method != strings.ToLower(e.Method) {
		return false
	}
	if len(t.Path) > 0 && !urlpath.Match(t.Path, e.Path) {
		return false
	}

	return true
}
