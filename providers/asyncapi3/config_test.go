package asyncapi3_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
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
	var cfg *asyncapi3.Config
	err := yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)

	multi := cfg.Components.Schemas["Foo"].Value.(*asyncapi3.MultiSchemaFormat)
	require.Equal(t, "application/vnd.apache.avro;version=1.9.0", multi.Format)
	require.Equal(t, map[string]interface{}{"type": "record"}, multi.Schema)

	single := cfg.Components.Schemas["Bar"].Value.(*schema.Schema)
	require.Equal(t, "object", single.Type[0])
}

func TestStreetlightKafka(t *testing.T) {
	b, err := os.ReadFile("./test/streetlight-kafka-3.0.yaml")
	require.NoError(t, err)

	var cfg *asyncapi3.Config
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

	// Server
	require.Len(t, cfg.Servers, 2)
	server := cfg.Servers["scram-connections"]
	require.Equal(t, "test.mykafkacluster.org:18092", server.Value.Host)
	require.Equal(t, "kafka-secure", server.Value.Protocol)
	require.Equal(t, "Test broker secured with scramSha256", server.Value.Description)

	server = cfg.Servers["mtls-connections"]
	require.Equal(t, "test.mykafkacluster.org:28092", server.Value.Host)
	require.Equal(t, "kafka-secure", server.Value.Protocol)
	require.Equal(t, "Test broker secured with X509", server.Value.Description)

	// Channel
	require.Len(t, cfg.Channels, 4)
	require.Contains(t, cfg.Channels, "lightingMeasured")
	require.Contains(t, cfg.Channels, "lightTurnOn")
	require.Contains(t, cfg.Channels, "lightTurnOff")
	require.Contains(t, cfg.Channels, "lightsDim")

	channel := cfg.Channels["lightingMeasured"]
	require.Equal(t, "smartylighting.streetlights.1.0.event.{streetlightId}.lighting.measured", channel.Value.Address)
	require.Equal(t, "The topic on which measured values may be produced and consumed.", channel.Description)

	// Message lightMeasured
	require.Len(t, channel.Value.Messages, 1)
	message := channel.Value.Messages["lightMeasured"]
	require.Equal(t, "lightMeasured", message.Value.Name)
	require.Equal(t, "Light measured", message.Value.Title)
	require.True(t, strings.HasPrefix(message.Value.Summary, "Inform about environmental"))
	require.Equal(t, "application/json", message.Value.ContentType)
	// header from message trait should be applied
	s := asyncapi3.ConvertToJsonSchema(message.Value.Headers.Value)
	require.Equal(t, "integer", s.Properties.Get("my-app-header").Value.Type[0])

	payload := message.Value.Payload.Value.(*schema.Schema)
	require.Equal(t, "Light intensity measured in lumens.", payload.Properties.Get("lumens").Value.Description)

	// message trait
	require.Equal(t, "#/components/messageTraits/commonHeaders", message.Value.Traits[0].Ref)
	trait := message.Value.Traits[0].Value
	s = trait.Headers.Value.(*schema.Schema)
	require.Equal(t, "integer", s.Properties.Get("my-app-header").Value.Type[0])

	param := channel.Value.Parameters["streetlightId"]
	require.Equal(t, "The ID of the streetlight.", param.Value.Description)

	require.Equal(t, "The ID of the streetlight.", cfg.Components.Parameters["streetlightId"].Value.Description)

	// Operation
	require.Len(t, cfg.Operations, 4)
	require.Contains(t, cfg.Operations, "receiveLightMeasurement")
	require.Contains(t, cfg.Operations, "turnOn")
	require.Contains(t, cfg.Operations, "turnOff")
	require.Contains(t, cfg.Operations, "dimLight")

	op := cfg.Operations["receiveLightMeasurement"]
	require.Equal(t, "receive", op.Value.Action)
	require.Equal(t, "Inform about environmental lighting conditions of a particular streetlight.", op.Value.Summary)
	require.Equal(t, "The topic on which measured values may be produced and consumed.", op.Value.Channel.Value.Description)
	// Trait
	require.Equal(t, "string", op.Value.Bindings.Kafka.ClientId.Type[0])
	require.Contains(t, op.Value.Bindings.Kafka.ClientId.Enum, "my-app-id")

}
