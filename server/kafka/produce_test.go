package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/test"
	"testing"
	"time"
)

func TestProduce(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"default",
			testProduce,
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

func testProduce(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", ":9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient(":9092", "kafkatest")
	defer client.Close()
	r, err := client.Produce(3, &produce.Request{Topics: []produce.RequestTopic{
		{Name: "foo", Data: produce.RequestPartition{
			Partition: 0,
			Record: protocol.RecordBatch{
				Offset: 0,
				Records: []protocol.Record{
					{
						Offset:  0,
						Time:    time.Now(),
						Key:     []byte("foo"),
						Value:   []byte("bar"),
						Headers: nil,
					},
				},
			},
		},
		}},
	})
	test.Ok(t, err)
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
}
