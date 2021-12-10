package kafka_test

import (
	"fmt"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/test"
	"strings"
	"testing"
)

func TestMetadata(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"default",
			testMetadata,
		},
		{
			"with topic",
			testMetadataWithTopic,
		},
		{
			"with number of partitions",
			testMetadataPartition,
		},
		{
			"with invalid topic",
			testMetadataInvalidTopic,
		},
		{
			"with invalid topic name",
			testMetadataInvalidTopicName,
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

func testMetadata(t *testing.T, b *kafka.Binding) {
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
	r, err := client.Metadata(4, &metaData.Request{})
	test.Ok(t, err)

	// controller
	test.Equals(t, int32(0), r.ControllerId)

	// brokers
	test.Equals(t, 1, len(r.Brokers))
	test.Equals(t, int32(0), r.Brokers[0].NodeId)
	test.Equals(t, "127.0.0.1", r.Brokers[0].Host)
	test.Equals(t, int32(9092), r.Brokers[0].Port)
	test.Equals(t, "", r.Brokers[0].Rack)

	// topics
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
	test.Equals(t, 1, len(r.Topics[0].Partitions))
	test.Equals(t, int32(0), r.Topics[0].Partitions[0].PartitionIndex)
	test.Equals(t, int32(0), r.Topics[0].Partitions[0].LeaderId) // default broker id is 0
	test.Equals(t, 1, len(r.Topics[0].Partitions[0].ReplicaNodes))
	test.Equals(t, int32(1), r.Topics[0].Partitions[0].ReplicaNodes[0])
	test.Equals(t, 1, len(r.Topics[0].Partitions[0].IsrNodes))
	test.Equals(t, int32(1), r.Topics[0].Partitions[0].IsrNodes[0])
	test.Equals(t, false, r.Topics[0].IsInternal)
}

func testMetadataWithTopic(t *testing.T, b *kafka.Binding) {
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
	r, err := client.Metadata(4, &metaData.Request{
		Topics: []metaData.TopicName{{Name: "foo"}},
	})
	test.Ok(t, err)

	// controller
	test.Equals(t, int32(0), r.ControllerId)

	// brokers
	test.Equals(t, 1, len(r.Brokers))
	test.Equals(t, int32(0), r.Brokers[0].NodeId)
	test.Equals(t, "127.0.0.1", r.Brokers[0].Host)
	test.Equals(t, int32(9092), r.Brokers[0].Port)
	test.Equals(t, "", r.Brokers[0].Rack)

	// topics
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
	test.Equals(t, 1, len(r.Topics[0].Partitions))
	test.Equals(t, int32(0), r.Topics[0].Partitions[0].PartitionIndex)
	test.Equals(t, int32(0), r.Topics[0].Partitions[0].LeaderId) // default broker id is 0
	test.Equals(t, 1, len(r.Topics[0].Partitions[0].ReplicaNodes))
	test.Equals(t, int32(1), r.Topics[0].Partitions[0].ReplicaNodes[0])
	test.Equals(t, 1, len(r.Topics[0].Partitions[0].IsrNodes))
	test.Equals(t, int32(1), r.Topics[0].Partitions[0].IsrNodes[0])
	test.Equals(t, false, r.Topics[0].IsInternal)
}

func testMetadataPartition(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo",
			asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema()))),
			asyncapitest.WithChannelKafka("partitions", "3"),
		))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	r, err := client.Metadata(4, &metaData.Request{
		Topics: []metaData.TopicName{{Name: "foo"}},
	})
	test.Ok(t, err)
	test.Equals(t, 3, len(r.Topics[0].Partitions))
}

func testMetadataInvalidTopic(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo",
			asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema()))),
			asyncapitest.WithChannelKafka("partitions", "3"),
		))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	r, err := client.Metadata(4, &metaData.Request{
		Topics: []metaData.TopicName{{Name: "foo"}, {Name: "bar"}},
	})
	test.Ok(t, err)
	test.Equals(t, 2, len(r.Topics))
	test.Equals(t, protocol.None, r.Topics[0].ErrorCode)
	test.Equals(t, protocol.UnknownTopicOrPartition, r.Topics[1].ErrorCode)
}

func testMetadataInvalidTopicName(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo",
			asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema()))),
			asyncapitest.WithChannelKafka("partitions", "3"),
		))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()

	for _, name := range []string{"", ".", "..", "event?", strings.Repeat("a", 250)} {
		testName := name
		if len(name) > 10 {
			testName = testName[:10] + "..."
		}
		t.Run(fmt.Sprintf("name %q", testName), func(t *testing.T) {
			r, err := client.Metadata(4, &metaData.Request{
				Topics: []metaData.TopicName{{Name: name}},
			})
			test.Ok(t, err)
			test.Equals(t, 1, len(r.Topics))
			test.Equals(t, protocol.InvalidTopic, r.Topics[0].ErrorCode)
		})
	}
}
