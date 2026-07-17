package engine

import (
	"mokapi/engine/common"
	"time"

	log "github.com/sirupsen/logrus"
)

type EventHandler struct {
	HttpEventDispatcher
	KafkaEventDispatcher
	MqttEventDispatcher
	WebsocketEventDispatcher
	MailEventDispatcher
	LdapEventDispatcher
}

func (e *EventHandler) Clear(key string) {
	if e == nil {
		return
	}
	e.HttpEventDispatcher.Clear(key)
	e.KafkaEventDispatcher.Clear(key)
	e.MqttEventDispatcher.Clear(key)
	e.WebsocketEventDispatcher.Clear(key)
	e.MailEventDispatcher.Clear(key)
	e.LdapEventDispatcher.Clear(key)
}

func runEventHandler(h common.EventHandler, args common.EventArgs, params ...interface{}) *common.Action {
	action := &common.Action{
		Tags: args.Tags,
	}
	start := time.Now()
	logs := len(action.Logs)

	ctx := &common.EventContext{
		EventLogger: action.AppendLog,
		Args:        params,
	}

	if b, err := h(ctx); err != nil {
		log.Errorf("unable to execute event handler: %v", err)
		action.Error = &common.Error{Message: err.Error()}
	} else if !b && logs == len(action.Logs) {
		return nil
	}
	log.WithField("handler", action).Debug("processed event handler")

	action.Parameters = getDeepCopy(params)
	action.Duration = time.Now().Sub(start).Milliseconds()
	return action
}

func (e *EventHandler) Has(key string) bool {
	if e == nil {
		return false
	}
	if e.HttpEventDispatcher.Has(key) {
		return true
	}
	if e.KafkaEventDispatcher.Has(key) {
		return true
	}
	if e.MqttEventDispatcher.Has(key) {
		return true
	}
	if e.WebsocketEventDispatcher.Has(key) {
		return true
	}
	if e.MailEventDispatcher.Has(key) {
		return true
	}
	if e.LdapEventDispatcher.Has(key) {
		return true
	}
	return false
}

func addDefaultTags(args *common.EventArgs, sh *scriptHost) {
	defaultTags := map[string]string{
		"name":    sh.name,
		"file":    sh.name,
		"fileKey": sh.file.Info.Key(),
		"event":   "http",
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
