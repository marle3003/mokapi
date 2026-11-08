package runtime_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"os"
	"testing"

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
