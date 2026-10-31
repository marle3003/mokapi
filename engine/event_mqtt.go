package engine

import (
	"mokapi/engine/common"
)

type MqttEventHandler = eventHandler[common.MqttFilter]

type MqttEventDispatcher struct {
	dispatcher[common.MqttFilter]
}

func (sh *scriptHost) OnMqtt(filter common.MqttFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.MqttEventDispatcher.Register(sh.name, newEventHandler(sh, "mqtt", filter, do, args))
}

func (e *MqttEventDispatcher) EmitMqtt(message *common.MqttMessageEvent) []*common.Action {
	return e.emit(emitOptions[common.MqttFilter]{
		shouldEmit: func(f common.MqttFilter) bool {
			if !matchesIfPresent(f.Api, message.Api) {
				return false
			}
			if !matchesIfPresent(f.Topic, message.Topic) {
				return false
			}
			return true
		},
	}, message)
}
