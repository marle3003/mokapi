package runtime_test

import (
	"encoding/json"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/runtime"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// The Kafka, MQTT and WebSocket stores merge files with the same title in the same way.
// These helpers build the files for their tests.

// twoFilesWithSameTitle returns a.yml and b.yml. b.yml overrides the host of server "a"
// of a.yml and adds server "b" and channel "fromB".
func twoFilesWithSameTitle(protocol string) (a, b *dynamic.Config) {
	a = getConfig("https://mokapi.io/a.yml", asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("a", protocol, ":1111"),
		asyncapi3test.WithChannel("fromA"),
	))
	b = getConfig("https://mokapi.io/b.yml", asyncapi3test.NewConfig(
		asyncapi3test.WithTitle("foo"),
		asyncapi3test.WithServer("a", protocol, ":2222"),
		asyncapi3test.WithServer("b", protocol, ":3333"),
		asyncapi3test.WithChannel("fromB"),
	))
	return
}

func requireMergedWithB(t *testing.T, merged *asyncapi3.Config) {
	t.Helper()
	require.Equal(t, ":2222", merged.Servers.Lookup("a").Value.Host, "b.yml overrides the host of a.yml")
	require.NotNil(t, merged.Servers.Lookup("b"))
	require.Contains(t, merged.Channels, "fromA")
	require.Contains(t, merged.Channels, "fromB")
}

// requireOnlyFileA fails if merged is not exactly the config of a.yml, in particular
// if a value of b.yml is left, e.g. an override of a nested value.
func requireOnlyFileA(t *testing.T, protocol string, merged *asyncapi3.Config) {
	t.Helper()
	fresh, _ := twoFilesWithSameTitle(protocol)
	require.JSONEq(t, marshal(t, fresh.Data), marshal(t, merged))
}

func marshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

// twoAsyncApi2FilesWithSameTitle returns two AsyncAPI 2.x files. The payload schema of
// the 2.x config is shared with the converted config. b.yml adds a property to it.
func twoAsyncApi2FilesWithSameTitle(t *testing.T) (a, b *dynamic.Config) {
	t.Helper()
	load := func(url string) *dynamic.Config {
		data, err := os.ReadFile("../config/dynamic/asyncApi/test/streetlight-kafka-2.0.yaml")
		require.NoError(t, err)
		var cfg *asyncApi.Config
		require.NoError(t, yaml.Unmarshal(data, &cfg))
		c := getConfig(url, nil)
		c.Data = cfg
		require.NoError(t, cfg.Parse(c, &dynamictest.Reader{}))
		return c
	}
	a = load("https://mokapi.io/a.yml")
	b = load("https://mokapi.io/b.yml")
	payload := b.Data.(*asyncApi.Config).Components.Schemas["lightMeasuredPayload"].Value.(*schema.Schema)
	payload.Properties.Set("extra", schematest.New("string"))
	return
}

func requireExtraPropertyOnlyIn(t *testing.T, merged *asyncapi3.Config, a, b *dynamic.Config) {
	t.Helper()
	extra := func(c *dynamic.Config) bool {
		s := c.Data.(*asyncApi.Config).Components.Schemas["lightMeasuredPayload"].Value.(*schema.Schema)
		return s.Properties.Get("extra") != nil
	}
	mergedSchema := merged.Components.Schemas["lightMeasuredPayload"].Value.(*schema.Schema)
	require.NotNil(t, mergedSchema.Properties.Get("extra"), "merged config has the property of b.yml")
	require.True(t, extra(b))
	require.False(t, extra(a), "a.yml must not change")
}

// asyncApiFile returns an AsyncAPI 3 file as the file provider reads it: unmarshaled from
// YAML and parsed, so empty entries and unresolved references look like in a real file.
func asyncApiFile(t *testing.T, url, content string) *dynamic.Config {
	t.Helper()
	var cfg *asyncapi3.Config
	require.NoError(t, yaml.Unmarshal([]byte(content), &cfg))
	c := getConfig(url, cfg)
	require.NoError(t, cfg.Parse(c, &dynamictest.Reader{}))
	return c
}

// Two files with the same title are merged into one config. An entry that is empty in one
// file (`foo:` without value) must not crash Mokapi when the other file defines the same entry.
func TestKafkaStore_MergeFilesWithIncompleteEntries(t *testing.T) {
	logrus.SetOutput(io.Discard)

	const head = "asyncapi: 3.0.0\ninfo:\n  title: foo\n  version: '1.0'\n"
	testcases := []struct {
		name  string
		a, b  string
		check func(t *testing.T, merged *asyncapi3.Config)
	}{
		{
			name: "component message is empty in first file",
			a:    head + "channels:\n  orders: {}\ncomponents:\n  messages:\n    order:\n",
			b:    head + "channels:\n  orders: {}\ncomponents:\n  messages:\n    order:\n      name: order\n",
			check: func(t *testing.T, merged *asyncapi3.Config) {
				require.Equal(t, "order", merged.Components.Messages["order"].Value.Name)
			},
		},
		{
			name: "component schema is empty in first file",
			a:    head + "channels:\n  orders: {}\ncomponents:\n  schemas:\n    order:\n",
			b:    head + "channels:\n  orders: {}\ncomponents:\n  schemas:\n    order:\n      type: string\n",
			check: func(t *testing.T, merged *asyncapi3.Config) {
				require.NotNil(t, merged.Components.Schemas["order"], "schema of second file is not lost")
			},
		},
		{
			name: "component channel is empty in first file",
			a:    head + "channels:\n  orders: {}\ncomponents:\n  channels:\n    order:\n",
			b:    head + "channels:\n  orders: {}\ncomponents:\n  channels:\n    order:\n      description: foo\n",
			check: func(t *testing.T, merged *asyncapi3.Config) {
				require.Equal(t, "foo", merged.Components.Channels["order"].Value.Description)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New(&static.Config{}, &dynamictest.Reader{})
			_, err := app.Kafka.Add(asyncApiFile(t, "https://mokapi.io/a.yml", tc.a), enginetest.NewEngine())
			require.NoError(t, err)

			var info *runtime.KafkaInfo
			require.NotPanics(t, func() {
				info, err = app.Kafka.Add(asyncApiFile(t, "https://mokapi.io/b.yml", tc.b), enginetest.NewEngine())
			})
			require.NoError(t, err)
			tc.check(t, info.Config)
		})
	}
}

// A single file with an empty entry (`foo:` without value) must not crash Mokapi, whichever
// protocol serves the file.
func TestAsyncApiStores_FileWithEmptyEntries(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		body string
	}{
		{name: "server", body: "  x:\n"},
		{name: "channel", body: "channels:\n  orders:\n"},
		{name: "server of channel", body: "channels:\n  orders:\n    address: orders\n    servers:\n      - \n"},
		{name: "parameter of channel", body: "channels:\n  orders:\n    address: orders\n    parameters:\n      id:\n"},
		{name: "message of channel", body: "channels:\n  orders:\n    address: orders\n    messages:\n      order:\n"},
		{name: "operation", body: "operations:\n  send:\n"},
		{name: "operation without channel", body: "operations:\n  send:\n    action: send\n"},
		{name: "message of operation", body: "channels:\n  orders:\n    address: orders\noperations:\n  send:\n    action: send\n    channel:\n      $ref: '#/channels/orders'\n    messages:\n      - \n"},
	}
	stores := map[string]func(app *runtime.App, c *dynamic.Config) error{
		"kafka": func(app *runtime.App, c *dynamic.Config) error {
			_, err := app.Kafka.Add(c, enginetest.NewEngine())
			return err
		},
		"mqtt": func(app *runtime.App, c *dynamic.Config) error {
			_, err := app.Mqtt.Add(c, enginetest.NewEngine())
			return err
		},
		"ws": func(app *runtime.App, c *dynamic.Config) error {
			_, err := app.Websocket.Add(c, enginetest.NewEngine())
			return err
		},
	}

	for protocol, add := range stores {
		for _, tc := range testcases {
			t.Run(protocol+" "+tc.name, func(t *testing.T) {
				content := "asyncapi: 3.0.0\ninfo:\n  title: foo\n  version: '1.0'\nservers:\n  dev:\n    host: ':1234'\n    protocol: " + protocol + "\n" + tc.body
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})

				require.NotPanics(t, func() {
					require.NoError(t, add(app, asyncApiFile(t, "https://mokapi.io/a.yml", content)))
				})
			})
		}
	}
}
