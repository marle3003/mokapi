package server

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/kafka/kafkatest"
	"mokapi/kafka/metaData"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/schema/json/schema"
	"mokapi/try"
	"testing"
	"time"
)

func TestKafkaServer(t *testing.T) {
	port := try.GetFreePort()
	addr := fmt.Sprintf("127.0.0.1:%v", port)
	c := asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("kafka12", "kafka", addr),
		asyncapi3test.WithChannel("foo",
			asyncapi3test.WithMessage("foo",
				asyncapi3test.WithPayload(
					&schema.Schema{Type: schema.Types{"string"}},
				),
			),
		),
	)

	m := NewKafkaManager(nil, runtime.New())
	defer m.Stop()
	m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}, Data: c}})

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
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("foo", "kafka", addr),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				port = try.GetFreePort()
				addr = fmt.Sprintf("127.0.0.1:%v", port)
				cfg.Servers["bar"] = &asyncapi3.ServerRef{Value: &asyncapi3.Server{
					Host:     addr,
					Protocol: "kafka",
				}}

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

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
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				cfg.Servers["broker"] = &asyncapi3.ServerRef{Value: &asyncapi3.Server{
					Host:     addr,
					Protocol: "kafka",
				}}

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

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
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithServer("", "kafka", addr),
					asyncapi3test.WithTitle("foo"),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Brokers, 1)

				delete(cfg.Servers, "")
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				r, err = client.Metadata(0, &metaData.Request{})
				require.EqualError(t, err, "EOF")
			},
		},
		{
			"change broker name",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("kafka", "kafka", addr),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				delete(cfg.Servers, "kafka")
				cfg.Servers["broker"] = &asyncapi3.ServerRef{Value: &asyncapi3.Server{
					Host:     addr,
					Protocol: "kafka",
				}}

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

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
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("add topic", "kafka", addr),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
					),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				cfg.Channels["bar"] = &asyncapi3.ChannelRef{Value: asyncapi3test.NewChannel(
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(
							&schema.Schema{Type: schema.Types{"string"}},
						),
					))}

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

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
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("remove topic", "kafka", addr),
					asyncapi3test.WithChannel("foo",
						asyncapi3test.WithMessage("foo",
							asyncapi3test.WithPayload(
								&schema.Schema{Type: schema.Types{"string"}},
							),
						),
					),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				delete(cfg.Channels, "foo")

				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				// wait for update
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				r, err := client.Metadata(0, &metaData.Request{})
				require.NoError(t, err)
				require.Len(t, r.Topics, 0)
			},
		},
		{
			"remove cluster",
			func(t *testing.T, m *KafkaManager) {
				port := try.GetFreePort()
				addr := fmt.Sprintf("127.0.0.1:%v", port)
				cfg := asyncapi3test.NewConfig(
					asyncapi3test.WithTitle("foo"),
					asyncapi3test.WithServer("foo", "kafka", addr),
				)
				m.UpdateConfig(dynamic.ConfigEvent{Config: &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")}}})

				m.UpdateConfig(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: &dynamic.Config{
						Info: dynamic.ConfigInfo{Url: MustParseUrl("foo.yml")},
						Data: cfg,
					},
				})

				// wait for kafka start
				time.Sleep(500 * time.Millisecond)

				client := kafkatest.NewClient(addr, "test")
				defer client.Close()

				_, err := client.Metadata(0, &metaData.Request{})
				require.EqualError(t, err, fmt.Sprintf("dial tcp 127.0.0.1:%v: connectex: No connection could be made because the target machine actively refused it.", port))
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
