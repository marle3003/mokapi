package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/server/kafka/protocol/offsetCommit"
	"mokapi/test"
	"testing"
)

func TestOffsetCommit(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(*testing.T, *kafka.Binding)
	}{
		{
			"group not exists",
			testOffsetCommitGroupNotExists,
		},
		{
			"offset out of range",
			testOffsetCommitOutOfRange,
		},
		{
			"offset commit successfully",
			testOffsetCommit,
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

func testOffsetCommitGroupNotExists(t *testing.T, b *kafka.Binding) {
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
	r, err := client.OffsetCommit(2, &offsetCommit.Request{
		GroupId:       "TestGroup",
		GenerationId:  0,
		MemberId:      "foo",
		RetentionTime: 0,
		Topics: []offsetCommit.Topic{
			{
				Name: "foo",
				Partitions: []offsetCommit.Partition{
					{
						Index:  0,
						Offset: 0,
					},
				},
			},
		},
	})
	test.Ok(t, err)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, 1, len(r.Topics[0].Partitions))
	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.UnknownMemberId, p.ErrorCode)
}

func testOffsetCommitOutOfRange(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())),
				asyncapitest.WithOperationBinding(&openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	err = client.JoinSyncGroup("foo", "TestGroup", 3, 3)
	test.Ok(t, err)
	r, err := client.OffsetCommit(2, &offsetCommit.Request{
		GroupId:       "TestGroup",
		GenerationId:  0,
		MemberId:      "foo",
		RetentionTime: 0,
		Topics: []offsetCommit.Topic{
			{
				Name: "foo",
				Partitions: []offsetCommit.Partition{
					{
						Index:    0,
						Offset:   99999,
						Metadata: "",
					},
				},
			},
		},
	})
	test.Ok(t, err)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, 1, len(r.Topics[0].Partitions))
	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.OffsetOutOfRange, p.ErrorCode)
}

func testOffsetCommit(t *testing.T, b *kafka.Binding) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())),
				asyncapitest.WithOperationBinding(&openapi.Schema{Type: "string", Enum: []interface{}{"TestGroup"}}))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient("127.0.0.1:9092", "kafkatest")
	defer client.Close()
	err = client.JoinSyncGroup("foo", "TestGroup", 3, 3)
	test.Ok(t, err)
	r, err := client.OffsetCommit(2, &offsetCommit.Request{
		GroupId:       "TestGroup",
		GenerationId:  0,
		MemberId:      "foo",
		RetentionTime: 0,
		Topics: []offsetCommit.Topic{
			{
				Name: "foo",
				Partitions: []offsetCommit.Partition{
					{
						Index:    0,
						Offset:   0,
						Metadata: "",
					},
				},
			},
		},
	})
	test.Ok(t, err)
	test.Equals(t, 1, len(r.Topics))
	test.Equals(t, "foo", r.Topics[0].Name)
	test.Equals(t, 1, len(r.Topics[0].Partitions))
	p := r.Topics[0].Partitions[0]
	test.Equals(t, protocol.None, p.ErrorCode)
}
