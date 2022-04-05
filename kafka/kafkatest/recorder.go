package kafkatest

import "mokapi/kafka"

type ResponseRecorder struct {
	ApiKey        kafka.ApiKey
	Version       int
	CorrelationId int
	Message       kafka.Message
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

func (r *ResponseRecorder) WriteHeader(key kafka.ApiKey, version, correlationId int) {
	r.ApiKey = key
	r.Version = version
	r.CorrelationId = correlationId
}

func (r *ResponseRecorder) Write(msg kafka.Message) error {
	r.Message = msg
	return nil
}
