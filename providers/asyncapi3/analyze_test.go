package asyncapi3_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAnalyze(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncapi3.Config
		test func(t *testing.T, a *asyncapi3.Analysis)
	}{
		{
			name: "no server no channels",
			cfg:  &asyncapi3.Config{},
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, false, a.HasServers)
				require.Equal(t, false, a.HasChannels)
				require.Len(t, a.Protocols, 0)
			},
		},
		{
			name: "server empty protocol",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithServer("", "", ""),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, true, a.HasServers)
				require.Equal(t, false, a.HasChannels)
				require.Len(t, a.Protocols, 0)
			},
		},
		{
			name: "server kafka protocol",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithServer("foo", "kafka", ""),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, true, a.HasServers)
				require.Equal(t, false, a.HasChannels)
				require.Contains(t, a.Protocols, "kafka")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Servers: []string{"foo"},
				}, a.Protocols["kafka"])
			},
		},
		{
			name: "server mqtt protocol",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithServer("foo", "mqtt", ""),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, true, a.HasServers)
				require.Equal(t, false, a.HasChannels)
				require.Contains(t, a.Protocols, "mqtt")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Servers: []string{"foo"},
				}, a.Protocols["mqtt"])
			},
		},
		{
			name: "server websocket protocol",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithServer("foo", "ws", ""),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, true, a.HasServers)
				require.Equal(t, false, a.HasChannels)
				require.Contains(t, a.Protocols, "ws")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Servers: []string{"foo"},
				}, a.Protocols["ws"])
			},
		},
		{
			name: "server kafka protocol and channel",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithServer("foo", "kafka", ""),
				asyncapi3test.WithChannel("c-1"),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, true, a.HasServers)
				require.Equal(t, true, a.HasChannels)
				require.Len(t, a.Protocols, 1)
				require.Contains(t, a.Protocols, "kafka")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Servers: []string{"foo"},
				}, a.Protocols["kafka"])
			},
		},
		{
			name: "channel",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("c-1"),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, false, a.HasServers)
				require.Equal(t, true, a.HasChannels)
				require.Len(t, a.Protocols, 0)
			},
		},
		{
			name: "channel with kafka binding",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("c-1",
					asyncapi3test.WithKafkaChannelBindingFrom(map[string]any{}),
				),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, false, a.HasServers)
				require.Equal(t, true, a.HasChannels)
				require.Len(t, a.Protocols, 1)
				require.Contains(t, a.Protocols, "kafka")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Channels: []string{"c-1"},
				}, a.Protocols["kafka"])
			},
		},
		{
			name: "channel with message binding",
			cfg: asyncapi3test.NewConfig(
				asyncapi3test.WithChannel("c-1",
					asyncapi3test.WithMessage("m-1",
						asyncapi3test.WithKafkaMessageBinding(asyncapi3.KafkaMessageBinding{SchemaIdLocation: "foo"}),
					),
				),
			),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, false, a.HasServers)
				require.Equal(t, true, a.HasChannels)
				require.Len(t, a.Protocols, 1)
				require.Contains(t, a.Protocols, "kafka")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Channels: []string{"c-1"},
				}, a.Protocols["kafka"])
			},
		},
		{
			name: "channel with operation binding",
			cfg: func() *asyncapi3.Config {
				ch := asyncapi3test.NewChannel()

				return asyncapi3test.NewConfig(
					asyncapi3test.AddChannel("c-1", ch),
					asyncapi3test.WithOperation("op",
						asyncapi3test.WithOperationAction("send"),
						asyncapi3test.WithOperationChannel(ch),
						asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{ClientId: schematest.New("string")}),
					),
				)
			}(),
			test: func(t *testing.T, a *asyncapi3.Analysis) {
				require.Equal(t, false, a.HasServers)
				require.Equal(t, true, a.HasChannels)
				require.Len(t, a.Protocols, 1)
				require.Contains(t, a.Protocols, "kafka")
				require.Equal(t, &asyncapi3.ProtocolUsage{
					Channels: []string{"c-1"},
				}, a.Protocols["kafka"])
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.cfg.Analyze()
			tc.test(t, a)
		})
	}
}

// Empty entries (`foo:` in YAML) are removed by Parse (Analyze always runs on parsed configs).
// An operation without channel is a missing field and must be skipped by Analyze itself.
func TestAnalyze_EmptyEntries(t *testing.T) {
	const head = "asyncapi: 3.0.0\ninfo:\n  title: foo\n  version: '1.0'\n"
	testcases := []struct {
		name string
		body string
	}{
		{name: "server", body: "servers:\n  dev:\n"},
		{name: "channel", body: "channels:\n  orders:\n"},
		{name: "operation", body: "operations:\n  send:\n"},
		{name: "operation without channel", body: "operations:\n  send:\n    action: send\n"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *asyncapi3.Config
			require.NoError(t, yaml.Unmarshal([]byte(head+tc.body), &cfg))
			require.NoError(t, cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{}))

			var a *asyncapi3.Analysis
			require.NotPanics(t, func() { a = cfg.Analyze() })
			require.False(t, a.HasServers)
			require.False(t, a.HasChannels)
			require.Empty(t, a.Protocols)
		})
	}
}
