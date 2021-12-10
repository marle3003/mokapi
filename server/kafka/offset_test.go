package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/server/kafka/protocol/offset"
	"mokapi/test"
	"testing"
)

func TestOffsetsFetch(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"empty",
			testListOffsetsFetchEmpty,
		},
		{
			"with single",
			testListOffsetsFetchWithSingle,
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

func testListOffsetsFetchEmpty(t *testing.T, b *kafka.Binding) {
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
	r, err := client.Offset(3, &offset.Request{Topics: []offset.RequestTopic{
		{
			Name: "foo",
			Partitions: []offset.RequestPartition{
				{
					Index:     0,
					Timestamp: protocol.Earliest,
				},
			},
		},
	}})
	test.Ok(t, err)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, 1, len(r.Topics[0].Partitions))

	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
	test.Equals(t, int64(-1), p.Offset)
}

func testListOffsetsFetchWithSingle(t *testing.T, b *kafka.Binding) {
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

	testProduce(t, b)

	r, err := client.Offset(3, &offset.Request{Topics: []offset.RequestTopic{
		{
			Name: "foo",
			Partitions: []offset.RequestPartition{
				{
					Index:     0,
					Timestamp: protocol.Earliest,
				},
			},
		},
	}})

	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
	test.Equals(t, protocol.Earliest, p.Timestamp)
	test.Equals(t, int64(-1), p.Offset)

	r, err = client.Offset(3, &offset.Request{Topics: []offset.RequestTopic{
		{
			Name: "foo",
			Partitions: []offset.RequestPartition{
				{
					Index:     0,
					Timestamp: protocol.Latest,
				},
			},
		},
	}})

	p = r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
	test.Equals(t, protocol.Latest, p.Timestamp)
	test.Equals(t, int64(0), p.Offset)
}
