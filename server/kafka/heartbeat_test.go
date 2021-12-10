package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
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
			"group not exists",
			testHeartbeatGroupNotExists,
		},
		{
			"group exists",
			testHeartbeatGroupExists,
		},
		{
			"group balancing",
			testHeartbeatGroupBalancing,
		},
		{
			"ok",
			testHeartbeat,
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

func testHeartbeatGroupNotExists(t *testing.T, b *kafka.Binding) {
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
	test.Equals(t, protocol.UnknownMemberId, r.ErrorCode)
}

func testHeartbeatGroupExists(t *testing.T, b *kafka.Binding) {
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
	r, err := client.Heartbeat(3, &heartbeat.Request{
		GroupId:      "TestGroup",
		GenerationId: 0,
	})
	test.Ok(t, err)
	test.Equals(t, protocol.UnknownMemberId, r.ErrorCode)
}

func testHeartbeatGroupBalancing(t *testing.T, b *kafka.Binding) {
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
	_, err = client.JoinGroup(3, &joinGroup.Request{
		GroupId:      "TestGroup",
		MemberId:     "foo",
		ProtocolType: "consumer",
		Protocols: []joinGroup.Protocol{{
			Name: "range",
		}},
	})
	r, err := client.Heartbeat(3, &heartbeat.Request{
		GroupId:         "TestGroup",
		GenerationId:    0,
		MemberId:        "foo",
		GroupInstanceId: "",
	})
	test.Ok(t, err)
	test.Equals(t, protocol.RebalanceInProgress, r.ErrorCode)
}

func testHeartbeat(t *testing.T, b *kafka.Binding) {
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
	r, err := client.Heartbeat(3, &heartbeat.Request{
		GroupId:         "TestGroup",
		GenerationId:    0,
		MemberId:        "foo",
		GroupInstanceId: "",
	})
	test.Ok(t, err)
	test.Equals(t, protocol.None, r.ErrorCode)
}
