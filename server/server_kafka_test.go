package server

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/protocol/metaData"
	"mokapi/test"
	"testing"
	"time"
)

func TestKafkaServer(t *testing.T) {
	c := asyncapitest.NewConfig(
		asyncapitest.WithTitle("foo"),
		asyncapitest.WithServer("kafka12", "kafka", "127.0.0.1:9092"),
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

	// wait for kafka start
	time.Sleep(500 * time.Millisecond)

	test.Equals(t, 1, len(clusters))
	_, ok := clusters["foo"]
	test.Assert(t, ok, "cluster exists")
}

func TestKafkaServer_Update(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, c KafkaClusters)
	}{
		{
			"add another broker",
			func(t *testing.T, c KafkaClusters) {
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("add topic", "kafka", "127.0.0.1:9092"),
				)
				c.UpdateConfig(&common.File{Data: cfg})

				cfg.Servers["broker"] = asyncApi.Server{
					Url:      "127.0.0.1:9093",
					Protocol: "kafka",
				}

				c.UpdateConfig(&common.File{Data: cfg})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient("127.0.0.1:9093", "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				test.Ok(t, err)
				test.Equals(t, 2, len(r.Brokers))
			},
		},
		{
			"add broker",
			func(t *testing.T, c KafkaClusters) {
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
				)
				c.UpdateConfig(&common.File{Data: cfg})

				cfg.Servers["broker"] = asyncApi.Server{
					Url:      "127.0.0.1:9092",
					Protocol: "kafka",
				}

				c.UpdateConfig(&common.File{Data: cfg})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient("127.0.0.1:9092", "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Brokers))
			},
		},
		{
			"change broker name",
			func(t *testing.T, c KafkaClusters) {
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("kafka", "kafka", "127.0.0.1:9092"),
				)
				c.UpdateConfig(&common.File{Data: cfg})

				delete(cfg.Servers, "kafka")
				cfg.Servers["broker"] = asyncApi.Server{
					Url:      "127.0.0.1:9092",
					Protocol: "kafka",
				}

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				c.UpdateConfig(&common.File{Data: cfg})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient("127.0.0.1:9092", "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				test.Ok(t, err)
				test.Equals(t, 1, len(r.Brokers))
			},
		},
		{
			"add topic",
			func(t *testing.T, c KafkaClusters) {
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("add topic", "kafka", "127.0.0.1:9092"),
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
				c.UpdateConfig(&common.File{Data: cfg})

				cfg.Channels["bar"] = &asyncApi.ChannelRef{Value: asyncapitest.NewChannel(asyncapitest.WithSubscribeAndPublish(
					asyncapitest.WithMessage(
						asyncapitest.WithPayload(
							&openapi.Schema{Type: "string"},
						),
					),
				))}

				c.UpdateConfig(&common.File{Data: cfg})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient("127.0.0.1:9092", "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				test.Ok(t, err)
				test.Equals(t, 2, len(r.Topics))
			},
		},
		{
			"remove topic",
			func(t *testing.T, c KafkaClusters) {
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("remove topic", "kafka", "127.0.0.1:9092"),
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
				c.UpdateConfig(&common.File{Data: cfg})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				delete(cfg.Channels, "foo")

				c.UpdateConfig(&common.File{Data: cfg})

				client := kafkatest.NewClient("127.0.0.1:9092", "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				test.Ok(t, err)
				test.Equals(t, 0, len(r.Topics))
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			c := KafkaClusters{}
			defer c.Stop()

			data.fn(t, c)
		})
	}
}
