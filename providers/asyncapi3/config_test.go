package asyncapi3_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/providers/swagger"
	jsonSchema "mokapi/schema/json/schema"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfig3_Schema(t *testing.T) {
	b := []byte(`asyncapi: 3.0.0
components:
  schemas:
    Foo:
      schemaFormat: 'application/vnd.apache.avro;version=1.9.0'
      schema:
        type: record
    FooRef:
      schemaFormat: 'application/vnd.apache.avro;version=1.9.0'
      schema:
        $ref: 'npm://foo.bar'
    Bar:
      type: object
`)
	var cfg *asyncapi3.Config
	err := yaml.Unmarshal(b, &cfg)
	require.NoError(t, err)

	multi := cfg.Components.Schemas["Foo"].Value.(*asyncapi3.MultiSchemaFormat)
	require.Equal(t, "application/vnd.apache.avro;version=1.9.0", multi.Format)
	avroSchema := multi.Schema.Value.(*asyncapi3.AvroRef)
	require.Equal(t, "record", avroSchema.Type[0])

	multi = cfg.Components.Schemas["FooRef"].Value.(*asyncapi3.MultiSchemaFormat)
	require.Equal(t, "application/vnd.apache.avro;version=1.9.0", multi.Format)
	require.Equal(t, "npm://foo.bar", multi.Schema.Ref)

	jsSchema := cfg.Components.Schemas["Bar"].Value.(*jsonSchema.Schema)
	require.Equal(t, "object", jsSchema.Type.String())

	require.Equal(t, "application/json", cfg.DefaultContentType)
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
	require.Equal(t, cfg.Servers.Len(), 2)
	server := cfg.Servers.Lookup("scram-connections")
	require.Equal(t, "test.mykafkacluster.org:18092", server.Value.Host)
	require.Equal(t, "kafka-secure", server.Value.Protocol)
	require.Equal(t, "Test broker secured with scramSha256", server.Value.Description)

	server = cfg.Servers.Lookup("mtls-connections")
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
	require.Equal(t, "lightingMeasured", channel.Value.Name)
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
	s := message.Value.Headers.Value.(*jsonSchema.Schema)
	require.Equal(t, "integer", s.Properties.Get("my-app-header").Type[0])

	payload := message.Value.Payload.Value.(*jsonSchema.Schema)
	require.Equal(t, "Light intensity measured in lumens.", payload.Properties.Get("lumens").Description)

	// message trait
	require.Equal(t, "#/components/messageTraits/commonHeaders", message.Value.Traits[0].Ref)
	trait := message.Value.Traits[0].Value
	s = trait.Headers.Value.(*jsonSchema.Schema)
	require.Equal(t, "integer", s.Properties.Get("my-app-header").Type[0])

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

func TestConfig_Payload_YAML(t *testing.T) {
	testcases := []struct {
		name   string
		cfg    string
		reader dynamic.Reader
		test   func(t *testing.T, cfg *asyncapi3.Config)
	}{
		{
			name: "just a schema",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      type: string`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				s := msg.Value.Payload.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: format and schema",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      schemaFormat: application/vnd.aai.asyncapi;version=3.0.0
      schema: 
        type: string`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				ms := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, ms)
				require.Equal(t, "application/vnd.aai.asyncapi;version=3.0.0", ms.Format)
				s := ms.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: no format",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      schema: 
        type: string`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				ms := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, ms)
				require.Equal(t, "", ms.Format)
				s := ms.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "ref",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      $ref: '#/components/schemas/bar'
    bar:
      type: string`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				s := msg.Value.Payload.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: ref",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      schemaFormat: application/vnd.aai.asyncapi;version=3.0.0
      schema:
        $ref: '#/components/schemas/bar'
    bar:
      type: string`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				ms := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, ms)
				require.Equal(t, "application/vnd.aai.asyncapi;version=3.0.0", ms.Format)
				s := ms.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "ref to swagger file",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      $ref: 'swagger.json#/definitions/foo'`,
			reader: dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
				return &dynamic.Config{Data: &swagger.Config{Definitions: map[string]*schema.Schema{"foo": schematest.New("string")}}}, nil
			}),
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				s := msg.Value.Payload.Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: ref to swagger file",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      schema:
        $ref: 'swagger.json#/definitions/foo'`,
			reader: dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
				return &dynamic.Config{Data: &swagger.Config{Definitions: map[string]*schema.Schema{"foo": schematest.New("string")}}}, nil
			}),
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)

				require.Equal(t, "", msf.Format)
				s := msf.Schema.Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: OpenAPI",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/msg'
components:
  messages:
    msg:
      payload:
        $ref: '#/components/schemas/foo'
  schemas:
    foo:
      schemaFormat: application/vnd.oai.openapi;version=3.0.0
      schema:
        type: string`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", msf.Format)
				s := msf.Schema.Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "multiple refs",
			cfg: `asyncapi: 3.0.0
channels:
  foo:
    messages:
      msg:
        payload:
          $ref: '#/components/schemas/foo'
components:
  schemas:
    foo:
      schemaFormat: application/vnd.oai.openapi;version=3.0.0
      schema:
        $ref: '#/components/schemas/bar'
    bar:
      $ref: '#/components/schemas/zzz'
    zzz:
      $ref: '#/components/schemas/yuh'
    yuh:
      items:
        $ref: '#/components/schemas/items'
    items:
      type: string
`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", msf.Format)
				// referenced is schema defines no format => default JSON schema
				s := msf.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Items.Type.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg *asyncapi3.Config
			err := yaml.Unmarshal([]byte(tc.cfg), &cfg)
			require.NoError(t, err)

			err = cfg.Parse(&dynamic.Config{Info: dynamictest.NewConfigInfo(), Data: cfg}, tc.reader)
			require.NoError(t, err)

			tc.test(t, cfg)
		})
	}
}

func TestConfig_Payload_JSON(t *testing.T) {
	testcases := []struct {
		name   string
		cfg    string
		reader dynamic.Reader
		test   func(t *testing.T, cfg *asyncapi3.Config)
	}{
		{
			name: "just a schema",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "type": "string"
     }
  }
}}`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				s := msg.Value.Payload.Value.(*jsonSchema.Schema)
				require.NotNil(t, s)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: format and schema",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "schemaFormat": "application/vnd.aai.asyncapi;version=3.0.0",
      "schema": {
        "type": "string"
      }
    }
  }
}}`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, msf)
				require.Equal(t, "application/vnd.aai.asyncapi;version=3.0.0", msf.Format)
				s := msf.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: no format",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "schema": { 
        "type": "string"
      }
    }
  }
}}`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, msf)
				require.Equal(t, "", msf.Format)
				s := msf.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "ref",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "$ref": "#/components/schemas/bar"
    },
    "bar": {
      "type": "string"
    }
  }
}}`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				s := msg.Value.Payload.Value.(*jsonSchema.Schema)
				require.NotNil(t, s)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: ref",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "schemaFormat": "application/vnd.aai.asyncapi;version=3.0.0",
      "schema": {
        "$ref": "#/components/schemas/bar"
      }
    },
    "bar": {
      "type": "string"
    }
  }
}}`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.Equal(t, "application/vnd.aai.asyncapi;version=3.0.0", msf.Format)
				s := msf.Schema.Value.(*jsonSchema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "ref to swagger file",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "$ref": "swagger.json#/definitions/foo"
    }
  }
}}`,
			reader: dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
				return &dynamic.Config{Data: &swagger.Config{Definitions: map[string]*schema.Schema{"foo": schematest.New("string")}}}, nil
			}),
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				s := msg.Value.Payload.Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: ref to swagger file",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "schema": {
        "$ref": "swagger.json#/definitions/foo"
      }
    }
  }
}}`,
			reader: dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
				return &dynamic.Config{Data: &swagger.Config{Definitions: map[string]*schema.Schema{"foo": schematest.New("string")}}}, nil
			}),
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, msf)
				s := msf.Schema.Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "MultiSchema: OpenAPI",
			cfg: `{"asyncapi": "3.0.0",
"channels": {
  "foo": {
    "messages": {
      "msg": {
        "$ref": "#/components/messages/msg"
      }
    }
  }
},
"components": {
  "messages": {
    "msg": {
      "payload": {
        "$ref": "#/components/schemas/foo"
      }
    }
  },
  "schemas": {
    "foo": {
      "schemaFormat": "application/vnd.oai.openapi;version=3.0.0",
      "schema": {
        "type": "string"
      }
    }
  }
}}`,
			test: func(t *testing.T, cfg *asyncapi3.Config) {
				ch := cfg.Channels["foo"]
				require.NotNil(t, ch)
				msg := ch.Value.Messages["msg"]
				require.NotNil(t, msg)
				msf := msg.Value.Payload.Value.(*asyncapi3.MultiSchemaFormat)
				require.NotNil(t, msf)
				require.Equal(t, "application/vnd.oai.openapi;version=3.0.0", msf.Format)
				s := msf.Schema.Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg *asyncapi3.Config
			err := json.Unmarshal([]byte(tc.cfg), &cfg)
			require.NoError(t, err)

			err = cfg.Parse(&dynamic.Config{Info: dynamictest.NewConfigInfo(), Data: cfg}, tc.reader)
			require.NoError(t, err)

			tc.test(t, cfg)
		})
	}
}
