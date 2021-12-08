package kafka_test

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/kafkatest"
	"mokapi/test"
	"testing"
)

func TestBindingApiVersion(t *testing.T) {
	b := kafka.NewBinding(func(topic string, key []byte, message []byte, partition int) {

	})
	c := asyncapitest.NewConfig(
		asyncapitest.WithServer("foo", "kafka", ":9092"),
		asyncapitest.WithChannel(
			"foo", asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(openapitest.NewSchema())))))
	err := b.Apply(c)
	test.Ok(t, err)

	client := kafkatest.NewClient(":9092", "kafkatest")
	r, err := client.Send(kafkatest.NewRequest("kafkatest", 3, &apiVersion.Request{}))
	test.Ok(t, err)
	test.Equals(t, protocol.ApiVersions, r.Header.ApiKey)
	b.Stop()
}
