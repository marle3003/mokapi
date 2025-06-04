package mqtttest

import (
	"mokapi/mqtt"
)

type ResponseRecorder struct {
	Message mqtt.Message
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

func (r *ResponseRecorder) Write(messageType mqtt.Type, msg mqtt.Message) {
	r.Message = msg
}
