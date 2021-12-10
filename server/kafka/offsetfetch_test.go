package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/test"
	"testing"
)

func TestOffsetFetch(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"empty",
			testOffsetFetchEmpty,
		},
		{
			"invalid partition request",
			testOffsetFetchInvalidPartition,
		},
		{
			"unknown topic request",
			testOffsetFetchUnknownTopic,
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

func testOffsetFetchEmpty(t *testing.T, b *kafka.Binding) {
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
	r, err := client.OffsetFetch(3, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
		{
			Name:             "foo",
			PartitionIndexes: []int32{0},
		},
	}})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, 1, len(r.Topics[0].Partitions))

	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
	test.Equals(t, int64(0), p.CommittedOffset)
}

func testOffsetFetchInvalidPartition(t *testing.T, b *kafka.Binding) {
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
	r, err := client.OffsetFetch(3, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
		{
			Name:             "foo",
			PartitionIndexes: []int32{9999},
		},
	}})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, 1, len(r.Topics[0].Partitions))

	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
	test.Equals(t, int64(-1), p.CommittedOffset)

	r, err = client.OffsetFetch(0, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
		{
			Name:             "foo",
			PartitionIndexes: []int32{9999},
		},
	}})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, 1, len(r.Topics[0].Partitions))

	p = r.Topics[0].Partitions[0]
	test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
}

func testOffsetFetchUnknownTopic(t *testing.T, b *kafka.Binding) {
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
	r, err := client.OffsetFetch(3, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
		{
			Name:             "unknown",
			PartitionIndexes: []int32{9999},
		},
	}})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, 1, len(r.Topics[0].Partitions))

	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
	test.Equals(t, int64(-1), p.CommittedOffset)

	r, err = client.OffsetFetch(0, &offsetFetch.Request{Topics: []offsetFetch.RequestTopic{
		{
			Name:             "unknown",
			PartitionIndexes: []int32{9999},
		},
	}})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, 1, len(r.Topics[0].Partitions))

	p = r.Topics[0].Partitions[0]
	test.Equals(t, protocol.UnknownTopicOrPartition, p.ErrorCode)
}
