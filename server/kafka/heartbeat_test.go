package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/test"
	"testing"
)

func TestHeartbeat(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"empty",
			testHeartbeatEmpty,
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			b := kafka.NewBinding(func(topic string, key []byte, message []byte, partition int) {})
			defer b.Stop()
			data.fn(t, b)
		})
	}
}

func testHeartbeatEmpty(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	r, err := client.Heartbeat(3, &heartbeat.Request{
		GroupId:         "",
		GenerationId:    0,
		MemberId:        "",
		GroupInstanceId: "",
	})
	test.Ok(t, err)
	test.Equals(t, protocol.GroupIdNotFound, r.ErrorCode)
}
