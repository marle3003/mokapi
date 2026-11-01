package engine

import (
	"mokapi/engine/common"
)

type KafkaEventHandler = eventHandler[common.KafkaFilter]

type KafkaEventDispatcher struct {
	dispatcher[common.KafkaFilter]
}

func (sh *scriptHost) OnKafka(filter common.KafkaFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.KafkaEventDispatcher.Register(sh.name, newEventHandler(sh, "kafka", filter, do, args))
}

func (e *KafkaEventDispatcher) EmitKafka(record *common.KafkaEventRecord) []*common.Action {
	return e.emit(emitOptions[common.KafkaFilter]{
		shouldEmit: func(f common.KafkaFilter) bool {
			if !matchesIfPresent(f.Api, record.Api) {
				return false
			}
			if !matchesIfPresent(f.Topic, record.Topic) {
				return false
			}
			return true
		},
	}, record)
}
