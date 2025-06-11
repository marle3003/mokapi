package mqtttest

import (
	"mokapi/mqtt"
)

type MessageRecorder struct {
	Message *mqtt.Message
}

func NewRecorder() *MessageRecorder {
	return &MessageRecorder{}
}

func (r *MessageRecorder) Write(msg *mqtt.Message) error {
	r.Message = msg
	return nil
}
