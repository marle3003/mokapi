package server

import (
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/test"
	"testing"
)

func TestKafkaServer(t *testing.T) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithTitle("foo"),
		asyncapitest.WithServer("kafka", "kafka", "127.0.0.1:9092"),
		asyncapitest.WithChannel("foo",
			asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(
						&openapi.Schema{Type: "string"},
					),
				),
			),
		),
	)

	clusters := KafkaClusters{}
	defer clusters.Stop()
	clusters.UpdateConfig(&common.File{Data: c})

	test.Equals(t, 1, len(clusters))
	_, ok := clusters["foo"]
	test.Assert(t, ok, "cluster exists")
}
