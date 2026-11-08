package asyncapi3

import (
	"mokapi/config/dynamic"
	openapi "mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"mokapi/sortedmap"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type otherSchema string

func (otherSchema) Parse(*dynamic.Config, dynamic.Reader) error { return nil }

func TestClone(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "nil stays nil",
			test: func(t *testing.T) {
				c := copies{}
				require.Nil(t, (*ServerRef)(nil).clone(c))
				require.Nil(t, (*ChannelRef)(nil).clone(c))
				require.Nil(t, (*MessageRef)(nil).clone(c))
				require.Nil(t, (*OperationRef)(nil).clone(c))
				require.Nil(t, (*SchemaRef)(nil).clone(c))
				require.Nil(t, (*Components)(nil).clone(c))
				require.Nil(t, (*ServerBindings)(nil).clone())
				require.Nil(t, cloneServers(nil, c))
				require.Nil(t, cloneMap[*ChannelRef](nil, c, (*ChannelRef).clone))
				require.Nil(t, cloneSlice[*TagRef](nil, c, (*TagRef).clone))
			},
		},
		{
			name: "reference keeps $ref, origin and value",
			test: func(t *testing.T) {
				origin := &dynamic.Config{}
				r := &ChannelRef{Value: &Channel{Title: "foo"}}
				r.Reference.Ref = "#/components/channels/foo"
				require.NoError(t, r.Reference.Parse(origin, nil))

				d := r.clone(copies{})

				require.Equal(t, r, d)
				require.NotSame(t, r, d)
				require.NotSame(t, r.Value, d.Value)

				d.Value.Title = "changed"
				require.Equal(t, "foo", r.Value.Title)
			},
		},
		{
			name: "channel is copied deep",
			test: func(t *testing.T) {
				ch := &Channel{
					Name:    "foo",
					Address: "foo",
					Servers: []*ServerRef{{Value: &Server{Host: "localhost"}}},
					Messages: map[string]*MessageRef{
						"msg": {Value: &Message{Name: "msg", Examples: []any{"a"}}},
					},
					Parameters:   map[string]*ParameterRef{"id": {Value: &Parameter{Enum: []string{"a"}}}},
					Bindings:     &ChannelBindings{Kafka: TopicBindings{configs: map[string]any{"partitions": 1}, Partitions: 1}},
					Tags:         []*TagRef{{Value: &Tag{Name: "tag", ExternalDocs: []*ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}}}}},
					ExternalDocs: []*ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}},
				}

				d := ch.clone(copies{})

				require.Equal(t, ch, d)
				d.Servers[0].Value.Host = "changed"
				d.Messages["msg"].Value.Name = "changed"
				d.Messages["msg"].Value.Examples[0] = "changed"
				d.Messages["added"] = &MessageRef{}
				d.Parameters["id"].Value.Enum[0] = "changed"
				d.Bindings.Kafka.configs["partitions"] = 2
				d.Bindings.Kafka.Partitions = 2
				d.Tags[0].Value.Name = "changed"
				d.Tags[0].Value.ExternalDocs[0].Value.Url = "changed"
				d.ExternalDocs[0].Value.Url = "changed"

				require.Equal(t, "localhost", ch.Servers[0].Value.Host)
				require.Equal(t, "msg", ch.Messages["msg"].Value.Name)
				require.Equal(t, []any{"a"}, ch.Messages["msg"].Value.Examples)
				require.Len(t, ch.Messages, 1)
				require.Equal(t, []string{"a"}, ch.Parameters["id"].Value.Enum)
				require.Equal(t, map[string]any{"partitions": 1}, ch.Bindings.Kafka.configs)
				require.Equal(t, 1, ch.Bindings.Kafka.Partitions)
				require.Equal(t, "tag", ch.Tags[0].Value.Name)
				require.Equal(t, "a", ch.Tags[0].Value.ExternalDocs[0].Value.Url)
				require.Equal(t, "a", ch.ExternalDocs[0].Value.Url)
			},
		},
		{
			name: "an operation and its channel map to the same copied channel",
			test: func(t *testing.T) {
				ch := &Channel{Name: "foo"}
				cfg := &Config{
					Channels: map[string]*ChannelRef{"foo": {Value: ch}},
					Operations: map[string]*OperationRef{
						"op": {Value: &Operation{
							Action:   "send",
							Channel:  &ChannelRef{Value: ch},
							Messages: []*MessageRef{{Value: &Message{Name: "msg"}}},
						}},
					},
				}
				c := copies{}

				channel := cfg.Channels["foo"].clone(c)
				op := cfg.Operations["op"].clone(c)

				require.NotSame(t, ch, channel.Value)
				require.Same(t, channel.Value, op.Value.Channel.Value)
			},
		},
		{
			name: "a server of the config and of a channel map to the same copy",
			test: func(t *testing.T) {
				server := &Server{Host: "localhost"}
				c := copies{}

				fromConfig := (&ServerRef{Value: server}).clone(c)
				fromChannel := (&Channel{Servers: []*ServerRef{{Value: server}}}).clone(c)

				require.NotSame(t, server, fromConfig.Value)
				require.Same(t, fromConfig.Value, fromChannel.Servers[0].Value)
			},
		},
		{
			name: "the back pointer of a channel to its config is not followed",
			test: func(t *testing.T) {
				ch := &Channel{Name: "foo"}
				cfg := &Config{Channels: map[string]*ChannelRef{"foo": {Value: ch}}}
				ch.Config = cfg

				d := ch.clone(copies{})

				require.Same(t, cfg, d.Config)
			},
		},
		{
			name: "operation is copied deep",
			test: func(t *testing.T) {
				op := &Operation{
					Action:   "send",
					Channel:  &ChannelRef{Value: &Channel{Name: "foo"}},
					Bindings: &OperationBindings{Kafka: KafkaOperationBinding{ClientId: schematest.New("string")}},
					Traits: []*OperationTraitRef{{Value: &OperationTrait{
						Channel:      &ChannelRef{Value: &Channel{Name: "bar"}},
						Bindings:     &OperationBindings{Kafka: KafkaOperationBinding{GroupId: schematest.New("string")}},
						ExternalDocs: []*ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}},
					}}},
					Messages:     []*MessageRef{{Value: &Message{Name: "msg"}}},
					ExternalDocs: []*ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}},
				}

				d := op.clone(copies{})

				require.Equal(t, op, d)
				d.Channel.Value.Name = "changed"
				d.Bindings.Kafka.ClientId.Type[0] = "integer"
				d.Traits[0].Value.Channel.Value.Name = "changed"
				d.Traits[0].Value.Bindings.Kafka.GroupId.Type[0] = "integer"
				d.Traits[0].Value.ExternalDocs[0].Value.Url = "changed"
				d.Messages[0].Value.Name = "changed"
				d.ExternalDocs[0].Value.Url = "changed"

				require.Equal(t, "foo", op.Channel.Value.Name)
				require.Equal(t, jsonSchema.Types{"string"}, op.Bindings.Kafka.ClientId.Type)
				require.Equal(t, "bar", op.Traits[0].Value.Channel.Value.Name)
				require.Equal(t, jsonSchema.Types{"string"}, op.Traits[0].Value.Bindings.Kafka.GroupId.Type)
				require.Equal(t, "a", op.Traits[0].Value.ExternalDocs[0].Value.Url)
				require.Equal(t, "msg", op.Messages[0].Value.Name)
				require.Equal(t, "a", op.ExternalDocs[0].Value.Url)
			},
		},
		{
			name: "message is copied deep",
			test: func(t *testing.T) {
				m := &Message{
					CorrelationId: &CorrelationIdRef{Value: &CorrelationId{Location: "a"}},
					Headers:       &SchemaRef{Value: schematest.New("object")},
					Payload:       &SchemaRef{Value: schematest.New("string")},
					Bindings:      &MessageBinding{Kafka: KafkaMessageBinding{Key: &SchemaRef{Value: schematest.New("string")}}},
					Traits: []*MessageTraitRef{{Value: &MessageTrait{
						CorrelationId: &CorrelationIdRef{Value: &CorrelationId{Location: "a"}},
						Headers:       &SchemaRef{Value: schematest.New("object")},
						Bindings:      &MessageBinding{Mqtt: MqttMessageBindings{configs: map[string]any{"a": 1}}},
						Examples:      []*MessageExample{{Name: "a", Headers: map[string]any{"a": 1}}},
						ExternalDocs:  []*ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}},
					}}},
					ExternalDocs: []*ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}},
				}

				d := m.clone(copies{})

				require.Equal(t, m, d)
				d.CorrelationId.Value.Location = "changed"
				d.Headers.Value.(*jsonSchema.Schema).Type[0] = "changed"
				d.Payload.Value.(*jsonSchema.Schema).Type[0] = "changed"
				d.Bindings.Kafka.Key.Value.(*jsonSchema.Schema).Type[0] = "changed"
				trait := d.Traits[0].Value
				trait.CorrelationId.Value.Location = "changed"
				trait.Headers.Value.(*jsonSchema.Schema).Type[0] = "changed"
				trait.Bindings.Mqtt.configs["a"] = 2
				trait.Examples[0].Headers["a"] = 2
				trait.ExternalDocs[0].Value.Url = "changed"
				d.ExternalDocs[0].Value.Url = "changed"

				require.Equal(t, "a", m.CorrelationId.Value.Location)
				require.Equal(t, jsonSchema.Types{"object"}, m.Headers.Value.(*jsonSchema.Schema).Type)
				require.Equal(t, jsonSchema.Types{"string"}, m.Payload.Value.(*jsonSchema.Schema).Type)
				require.Equal(t, jsonSchema.Types{"string"}, m.Bindings.Kafka.Key.Value.(*jsonSchema.Schema).Type)
				orig := m.Traits[0].Value
				require.Equal(t, "a", orig.CorrelationId.Value.Location)
				require.Equal(t, jsonSchema.Types{"object"}, orig.Headers.Value.(*jsonSchema.Schema).Type)
				require.Equal(t, map[string]any{"a": 1}, orig.Bindings.Mqtt.configs)
				require.Equal(t, map[string]any{"a": 1}, orig.Examples[0].Headers)
				require.Equal(t, "a", orig.ExternalDocs[0].Value.Url)
				require.Equal(t, "a", m.ExternalDocs[0].Value.Url)
			},
		},
		{
			name: "schema formats",
			test: func(t *testing.T) {
				openApi := &openapi.Schema{Type: jsonSchema.Types{"string"}}
				multi := &MultiSchemaFormat{Format: "f", Schema: &SchemaRef{Value: schematest.New("string")}}
				avro := &AvroRef{}
				nested := &SchemaRef{Value: schematest.New("string")}
				c := copies{}

				o := (&SchemaRef{Value: openApi}).clone(c).Value.(*openapi.Schema)
				m := (&SchemaRef{Value: multi}).clone(c).Value.(*MultiSchemaFormat)
				a := (&SchemaRef{Value: avro}).clone(c).Value.(*AvroRef)

				require.Equal(t, openApi, o)
				require.NotSame(t, openApi, o)
				o.Type[0] = "changed"
				require.Equal(t, jsonSchema.Types{"string"}, openApi.Type)

				require.Equal(t, multi, m)
				require.NotSame(t, multi, m)
				m.Schema.Value.(*jsonSchema.Schema).Type[0] = "changed"
				require.Equal(t, jsonSchema.Types{"string"}, multi.Schema.Value.(*jsonSchema.Schema).Type)

				require.Equal(t, avro, a)
				require.NotSame(t, avro, a)

				n := (&SchemaRef{Value: nested}).clone(c).Value.(*SchemaRef)
				require.Equal(t, nested, n)
				require.NotSame(t, nested, n)

				// unknown schema kinds are taken over
				require.Equal(t, otherSchema("other"), (&SchemaRef{Value: otherSchema("other")}).clone(c).Value)
			},
		},
		{
			name: "schema ref copies its reference and its value",
			test: func(t *testing.T) {
				r := &SchemaRef{Value: schematest.New("string")}
				r.Ref = "#/components/schemas/foo"

				d := r.clone(copies{})

				require.Equal(t, "#/components/schemas/foo", d.Ref)
				require.Equal(t, r.Value, d.Value)
				require.NotSame(t, r.Value, d.Value)
			},
		},
		{
			name: "servers keep their order and key normalizer",
			test: func(t *testing.T) {
				servers := &sortedmap.LinkedHashMap[string, *ServerRef]{KeyNormalizer: strings.ToLower}
				servers.Set("B", &ServerRef{Value: &Server{Host: "b"}})
				servers.Set("a", &ServerRef{Value: &Server{Host: "a"}})

				d := cloneServers(servers, copies{})

				require.Equal(t, []string{"b", "a"}, d.Keys())
				s, ok := d.Get("B")
				require.True(t, ok, "key normalizer is kept")
				require.NotSame(t, servers.Values()[0], s)

				s.Value.Host = "changed"
				require.Equal(t, "b", servers.Values()[0].Value.Host)
			},
		},
		{
			name: "server is copied deep",
			test: func(t *testing.T) {
				s := &Server{
					Variables:    map[string]*ServerVariableRef{"v": {Value: &ServerVariable{Enum: []string{"a"}, Examples: []string{"a"}}}},
					Tags:         []*TagRef{{Value: &Tag{Name: "tag"}}},
					Bindings:     &ServerBindings{Kafka: BrokerBindings{configs: map[string]any{"a": 1}}},
					ExternalDocs: []ExternalDocRef{{Value: &ExternalDoc{Url: "a"}}},
				}

				d := s.clone(copies{})

				require.Equal(t, s, d)
				d.Variables["v"].Value.Enum[0] = "changed"
				d.Variables["v"].Value.Examples[0] = "changed"
				d.Tags[0].Value.Name = "changed"
				d.Bindings.Kafka.configs["a"] = 2
				d.ExternalDocs[0].Value.Url = "changed"

				require.Equal(t, []string{"a"}, s.Variables["v"].Value.Enum)
				require.Equal(t, []string{"a"}, s.Variables["v"].Value.Examples)
				require.Equal(t, "tag", s.Tags[0].Value.Name)
				require.Equal(t, map[string]any{"a": 1}, s.Bindings.Kafka.configs)
				require.Equal(t, "a", s.ExternalDocs[0].Value.Url)
			},
		},
		{
			name: "channel and operation bindings",
			test: func(t *testing.T) {
				b := &ChannelBindings{
					Mqtt:      MqttChannelBindings{configs: map[string]any{"a": 1}},
					Websocket: WebsocketChannelBindings{Method: "GET", Query: schematest.New("object"), Headers: schematest.New("object")},
				}
				ob := &OperationBindings{
					Mqtt:      MqttOperationBindings{configs: map[string]any{"a": 1}},
					Websocket: WebsocketOperationBindings{configs: map[string]any{"a": 1}},
				}

				d := b.clone()
				od := ob.clone()

				require.Equal(t, b, d)
				require.Equal(t, ob, od)
				d.Mqtt.configs["a"] = 2
				d.Websocket.Query.Type[0] = "changed"
				d.Websocket.Headers.Type[0] = "changed"
				od.Mqtt.configs["a"] = 2
				od.Websocket.configs["a"] = 2

				require.Equal(t, map[string]any{"a": 1}, b.Mqtt.configs)
				require.Equal(t, jsonSchema.Types{"object"}, b.Websocket.Query.Type)
				require.Equal(t, jsonSchema.Types{"object"}, b.Websocket.Headers.Type)
				require.Equal(t, map[string]any{"a": 1}, ob.Mqtt.configs)
				require.Equal(t, map[string]any{"a": 1}, ob.Websocket.configs)
			},
		},
		{
			name: "components",
			test: func(t *testing.T) {
				cmp := &Components{
					Servers:         map[string]*ServerRef{"a": {Value: &Server{Host: "a"}}},
					Tags:            map[string]*TagRef{"a": {Value: &Tag{Name: "a"}}},
					Channels:        map[string]*ChannelRef{"a": {Value: &Channel{Name: "a"}}},
					Schemas:         map[string]*SchemaRef{"a": {Value: schematest.New("string")}},
					Messages:        map[string]*MessageRef{"a": {Value: &Message{Name: "a"}}},
					Operations:      map[string]*OperationRef{"a": {Value: &Operation{Title: "a"}}},
					Parameters:      map[string]*ParameterRef{"a": {Value: &Parameter{Default: "a"}}},
					CorrelationIds:  map[string]*CorrelationIdRef{"a": {Value: &CorrelationId{Location: "a"}}},
					ExternalDocs:    map[string]*ExternalDocRef{"a": {Value: &ExternalDoc{Url: "a"}}},
					OperationTraits: map[string]*OperationTraitRef{"a": {Value: &OperationTrait{Title: "a"}}},
					MessageTraits:   map[string]*MessageTraitRef{"a": {Value: &MessageTrait{Name: "a"}}},
					ServerVariables: map[string]*ServerVariableRef{"a": {Value: &ServerVariable{Default: "a"}}},
				}

				d := cmp.clone(copies{})

				require.Equal(t, cmp, d)
				d.Servers["a"].Value.Host = "changed"
				d.Tags["a"].Value.Name = "changed"
				d.Channels["a"].Value.Name = "changed"
				d.Schemas["a"].Value.(*jsonSchema.Schema).Type[0] = "changed"
				d.Messages["a"].Value.Name = "changed"
				d.Operations["a"].Value.Title = "changed"
				d.Parameters["a"].Value.Default = "changed"
				d.CorrelationIds["a"].Value.Location = "changed"
				d.ExternalDocs["a"].Value.Url = "changed"
				d.OperationTraits["a"].Value.Title = "changed"
				d.MessageTraits["a"].Value.Name = "changed"
				d.ServerVariables["a"].Value.Default = "changed"

				require.Equal(t, "a", cmp.Servers["a"].Value.Host)
				require.Equal(t, "a", cmp.Tags["a"].Value.Name)
				require.Equal(t, "a", cmp.Channels["a"].Value.Name)
				require.Equal(t, jsonSchema.Types{"string"}, cmp.Schemas["a"].Value.(*jsonSchema.Schema).Type)
				require.Equal(t, "a", cmp.Messages["a"].Value.Name)
				require.Equal(t, "a", cmp.Operations["a"].Value.Title)
				require.Equal(t, "a", cmp.Parameters["a"].Value.Default)
				require.Equal(t, "a", cmp.CorrelationIds["a"].Value.Location)
				require.Equal(t, "a", cmp.ExternalDocs["a"].Value.Url)
				require.Equal(t, "a", cmp.OperationTraits["a"].Value.Title)
				require.Equal(t, "a", cmp.MessageTraits["a"].Value.Name)
				require.Equal(t, "a", cmp.ServerVariables["a"].Value.Default)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, tc.test)
	}
}
