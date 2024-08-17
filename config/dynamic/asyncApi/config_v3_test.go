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

func TestConfig3_Schema(t *testing.T) {
	b := []byte(`asyncapi: 3.0.0
components:
  schemas:
    Foo:
      schemaFormat: 'application/vnd.apache.avro;version=1.9.0'
      schema:
        type: record
    Bar:
      type: object
`)
	var cfg *Config3
	err := yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)

	multi := cfg.Components.Schemas["Foo"].Value.(*MultiSchemaFormat)
	require.Equal(t, "application/vnd.apache.avro;version=1.9.0", multi.Format)
	require.Equal(t, map[string]interface{}{"type": "record"}, multi.Schema)

	single := cfg.Components.Schemas["Bar"].Value.(*schema.Schema)
	require.Equal(t, "object", single.Type[0])
}

func TestStreetlightKafka(t *testing.T) {
	b, err := os.ReadFile("./test/streetlight-kafka-3.0.yaml")
	require.NoError(t, err)

	var cfg *Config3
	err = yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)
	err = cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{})
	require.NoError(t, err)

	require.Equal(t, "3.0.0", cfg.Version)
	require.Equal(t, "Streetlights Kafka API", cfg.Info.Name)
	require.Equal(t, "1.0.0", cfg.Info.Version)
	require.True(t, strings.HasPrefix(cfg.Info.Description, "The Smartylighting"), "should have description")
	require.Equal(t, "Apache 2.0", cfg.Info.License.Name)
	require.Equal(t, "https://www.apache.org/licenses/LICENSE-2.0", cfg.Info.License.Url)
	require.Equal(t, "application/json", cfg.DefaultContentType)

	require.Len(t, cfg.Servers, 2)
	server := cfg.Servers["scram-connections"]
	require.Equal(t, "test.mykafkacluster.org:18092", server.Value.Host)
	require.Equal(t, "kafka-secure", server.Value.Protocol)
	require.Equal(t, "Test broker secured with scramSha256", server.Value.Description)

	require.Len(t, cfg.Channels, 4)
	channel := cfg.Channels["lightingMeasured"]
	require.Equal(t, "smartylighting.streetlights.1.0.event.{streetlightId}.lighting.measured", channel.Value.Address)
	require.Equal(t, "The topic on which measured values may be produced and consumed.", channel.Description)

	require.Len(t, channel.Value.Messages, 1)
	message := channel.Value.Messages["lightMeasured"]
	require.Equal(t, "lightMeasured", message.Value.Name)
	require.Equal(t, "Light measured", message.Value.Title)
	require.True(t, strings.HasPrefix(message.Value.Summary, "Inform about environmental"))
	require.Equal(t, "application/json", message.Value.ContentType)
	require.Equal(t, "#/components/messageTraits/commonHeaders", message.Value.Traits[0].Ref)
	trait := message.Value.Traits[0].Value
	s := trait.Headers.Value.(*schema.Schema)
	require.Equal(t, "integer", s.Properties.Get("my-app-header").Value.Type[0])
	payload := message.Value.Payload.Value.(*schema.Schema)
	require.Equal(t, "Light intensity measured in lumens.", payload.Properties.Get("lumens").Value.Description)

	param := channel.Value.Parameters["streetlightId"]
	require.Equal(t, "The ID of the streetlight.", param.Value.Description)

	require.Equal(t, "The ID of the streetlight.", cfg.Components.Parameters["streetlightId"].Value.Description)
}
