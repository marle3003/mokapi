package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/runtime/monitor"
)

type KafkaInfo struct {
	*asyncApi.Config
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func NewKafkaMonitor(kafka *monitor.Kafka, next kafka.Handler) *KafkaHandler {
	return &KafkaHandler{kafka: kafka, next: next}
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	switch req.Header.ApiKey {
	case kafka.Produce:
		h.kafka.Messages.Add(1)
	}

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}
