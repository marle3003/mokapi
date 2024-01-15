package server

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/try"
	"testing"
	"time"
)

func TestKafkaServer(t *testing.T) {
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	c := asyncapitest.NewConfig(
		asyncapitest.WithTitle("foo"),
		asyncapitest.WithServer("kafka12", "kafka", addr),
		asyncapitest.WithChannel("foo",
			asyncapitest.WithSubscribeAndPublish(
				asyncapitest.WithMessage(
					asyncapitest.WithPayload(
						&schema.Schema{Type: "string"},
					),
				),
			),
		),
	)

	m := NewKafkaManager(nil, runtime.New())
	defer m.Stop()
	m.UpdateConfig(&dynamic.Config{Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}, Data: c})

	// wait for kafka start
	time.Sleep(500 * time.Millisecond)

	require.Len(t, m.clusters, 1)
	_, ok := m.clusters["foo"]
	require.True(t, ok, "cluster exists")
}

func TestKafkaServer_Update(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, m *KafkaManager)
	}{
		{
			"add another broker",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("add topic", "kafka", addr),
				)
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				port = try.GetFreePort()
				addr = fmt.Sprintf("127.0.0.1:%v", port)
				cfg.Servers["broker"] = asyncApi.Server{
					Url:      addr,
					Protocol: "kafka",
				}

				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Brokers, 2)
			},
		},
		{
			"add broker",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
				)
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				cfg.Servers["broker"] = asyncApi.Server{
					Url:      addr,
					Protocol: "kafka",
				}

				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Brokers, 1)
			},
		},
		{
			"remove broker",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithServer("", "kafka", addr),
					asyncapitest.WithTitle("foo"),
				)
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Brokers, 1)

				delete(cfg.Servers, "")
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				r, err = client.Metadata(0, &metaData.Request{})
				require.EqualError(t, err, "EOF")
			},
		},
		{
			"change broker name",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("kafka", "kafka", addr),
				)
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				delete(cfg.Servers, "kafka")
				cfg.Servers["broker"] = asyncApi.Server{
					Url:      addr,
					Protocol: "kafka",
				}

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Brokers, 1)
			},
		},
		{
			"add topic",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("add topic", "kafka", addr),
					asyncapitest.WithChannel("foo",
						asyncapitest.WithSubscribeAndPublish(
							asyncapitest.WithMessage(
								asyncapitest.WithPayload(
									&schema.Schema{Type: "string"},
								),
							),
						),
					),
				)
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				cfg.Channels["bar"] = &asyncApi.ChannelRef{Value: asyncapitest.NewChannel(asyncapitest.WithSubscribeAndPublish(
					asyncapitest.WithMessage(
						asyncapitest.WithPayload(
							&schema.Schema{Type: "string"},
						),
					),
				))}

				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Topics, 2)
			},
		},
		{
			"remove topic",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapitest.NewConfig(
					asyncapitest.WithTitle("foo"),
					asyncapitest.WithServer("remove topic", "kafka", addr),
					asyncapitest.WithChannel("foo",
						asyncapitest.WithSubscribeAndPublish(
							asyncapitest.WithMessage(
								asyncapitest.WithPayload(
									&schema.Schema{Type: "string"},
								),
							),
						),
					),
				)
				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				delete(cfg.Channels, "foo")

				m.UpdateConfig(&dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				// wait for update
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Topics, 0)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := NewKafkaManager(nil, runtime.New())
			defer m.Stop()

			tc.fn(t, m)

			events.Reset()
		})
	}
}
