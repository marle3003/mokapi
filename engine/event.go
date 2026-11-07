package engine

import (
	"mokapi/engine/common"
)

type keyedDispatcher interface {
	Has(key string) bool
	Clear(key string)
}

type EventHandlerDispatcher struct {
	HttpEventDispatcher
	KafkaEventDispatcher
	MqttEventDispatcher
	WebsocketEventDispatcher
	MailEventDispatcher
	LdapEventDispatcher
}

func (e *EventHandlerDispatcher) dispatcher() []keyedDispatcher {
	return []keyedDispatcher{
		&e.HttpEventDispatcher,
		&e.KafkaEventDispatcher,
		&e.MqttEventDispatcher,
		&e.WebsocketEventDispatcher,
		&e.MailEventDispatcher,
		&e.LdapEventDispatcher,
	}
}

func (e *EventHandlerDispatcher) Clear(key string) {
	if e == nil {
		return
	}
	for _, d := range e.dispatcher() {
		d.Clear(key)
	}
}

func (e *EventHandlerDispatcher) Has(key string) bool {
	for _, d := range e.dispatcher() {
		if d.Has(key) {
			return true
		}
	}
	return false
}

func addDefaultTags(args *common.EventArgs, sh *scriptHost, event string) {
	defaultTags := map[string]string{
		"name":    sh.name,
		"file":    sh.name,
		"fileKey": sh.file.Info.Key(),
		"event":   event,
	}
	if args.Tags == nil {
		args.Tags = defaultTags
		return
	}
	for k, v := range defaultTags {
		if _, ok := args.Tags[k]; !ok {
			args.Tags[k] = v
		}
	}
}

func matchesIfPresent(a, b string) bool {
	return a == "" || a == b
}
