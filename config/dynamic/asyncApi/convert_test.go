package asyncApi

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/schema/json/schema"
	"os"
	"strings"
	"testing"
)

func TestConfig_Convert(t *testing.T) {
	b, err := os.ReadFile("./test/streetlight-kafka-2.0.yaml")
	require.NoError(t, err)

	var cfg *Config
	err = yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)

	err = cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{})
	require.NoError(t, err)

	cfg3, err := cfg.Convert()
	require.NoError(t, err)

	require.Equal(t, "3.0.0", cfg3.Version)
	require.Equal(t, "urn:example:com:smartylighting:streetlights:server", cfg3.Id)
	require.Equal(t, "application/json", cfg3.DefaultContentType)

	require.Equal(t, "Streetlights API", cfg3.Info.Name)
	require.Equal(t, "1.0.0", cfg3.Info.Version)
	require.True(t, strings.HasPrefix(cfg3.Info.Description, "The Smartylighting Streetlights API"))
	require.Equal(t, "Apache 2.0", cfg3.Info.License.Name)
	require.Equal(t, "https://www.apache.org/licenses/LICENSE-2.0", cfg3.Info.License.Url)

	// Server
	require.Len(t, cfg3.Servers, 1)
	require.Equal(t, "test.mosquitto.org:{port}", cfg3.Servers["production"].Value.Host)
	require.Equal(t, "mqtt", cfg3.Servers["production"].Value.Protocol)
	require.Equal(t, "Test broker", cfg3.Servers["production"].Value.Description)

	// Channel
	channel := cfg3.Channels["smartylighting/streetlights/1/0/event/{streetlightId}/lighting/measured"].Value
	require.Equal(t, "smartylighting/streetlights/1/0/event/{streetlightId}/lighting/measured", channel.Address)
	require.Equal(t, "The topic on which measured values may be produced and consumed.", channel.Description)
	param := channel.Parameters["streetlightId"].Value
	require.Equal(t, "The ID of the streetlight.", param.Description)
	require.Contains(t, channel.Messages, "receiveLightMeasurement")

	// message
	msg := channel.Messages["receiveLightMeasurement"].Value
	require.Equal(t, "lightMeasured", msg.Name)
	require.Equal(t, "Light measured", msg.Title)
	require.Equal(t, "Inform about environmental lighting conditions of a particular streetlight.", msg.Summary)

	// traits
	require.Len(t, msg.Traits, 1)
	require.IsType(t, &schema.Schema{}, msg.Traits[0].Value.Headers.Value.Schema)
	headers := msg.Traits[0].Value.Headers.Value.Schema.(*schema.Schema)
	require.Equal(t, float64(100), *headers.Properties.Get("my-app-header").Maximum)

	// payload
	payload := msg.Payload.Value
	require.IsType(t, &schema.Schema{}, payload.Schema)
	s := payload.Schema.(*schema.Schema)
	require.Equal(t, "object", s.Type[0])
}
