package runtime

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/kafka/store"
	"mokapi/kafka"
	"mokapi/runtime/monitor"
)

type KafkaInfo struct {
	*asyncApi.Config
	*store.Store
}

type KafkaHandler struct {
	kafka *monitor.Kafka
	next  kafka.Handler
}

func NewKafkaHandler(kafka *monitor.Kafka, next kafka.Handler) *KafkaHandler {
	return &KafkaHandler{kafka: kafka, next: next}
}

func (h *KafkaHandler) ServeMessage(rw kafka.ResponseWriter, req *kafka.Request) {
	ctx := monitor.NewKafkaContext(req.Context, h.kafka)

	req.WithContext(ctx)
	h.next.ServeMessage(rw, req)
}
