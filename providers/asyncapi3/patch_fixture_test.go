package asyncapi3

import (
	"fmt"
	"mokapi/config/dynamic"
	openapi "mokapi/providers/openapi/schema"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"mokapi/sortedmap"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// maximalConfig returns a config in which every field of the asyncapi3 types is set
// (see requireAllFieldsSet). The key is used for all map keys and names, so configs
// with different keys have nothing in common that Patch could merge.
// Values stored in interfaces (examples, binding values) are scalars, Patch takes them over.
func maximalConfig(key string) *Config {
	server := fullServer(key)
	channel := fullChannel(key, server)
	message := fullMessage(key)
	servers := &sortedmap.LinkedHashMap[string, *ServerRef]{}
	servers.Set(key, &ServerRef{Value: server})

	return &Config{
		Version: "3.0.0",
		Id:      "urn:" + key,
		Info: Info{
			Name:           "title",
			Description:    key + " description",
			Version:        "1.0",
			TermsOfService: "https://terms.io",
			Contact:        &Contact{Name: key, Url: "https://contact.io", Email: "contact@foo.bar"},
			License:        &License{Name: "MIT", Url: "https://license.io"},
			ExternalDocs:   []ExternalDocRef{{Value: fullExternalDoc(key)}},
		},
		DefaultContentType: "application/json",
		Servers:            servers,
		Channels: map[string]*ChannelRef{
			key: {Reference: reference[*ChannelRef](), Value: channel},
		},
		Operations: map[string]*OperationRef{key: {Value: fullOperation(key, "send", channel, message)}},
		Components: &Components{
			Servers:         map[string]*ServerRef{key: {Value: fullServer(key)}},
			Tags:            map[string]*TagRef{key: {Value: fullTag(key)}},
			Channels:        map[string]*ChannelRef{key: {Value: fullChannel(key, server)}},
			Schemas:         map[string]*SchemaRef{key: {Reference: reference[*SchemaRef](), Value: schematest.New("string")}},
			Messages:        map[string]*MessageRef{key: {Value: message}},
			Operations:      map[string]*OperationRef{key: {Value: fullOperation(key, "receive", channel, message)}},
			Parameters:      map[string]*ParameterRef{key: {Value: fullParameter()}},
			CorrelationIds:  map[string]*CorrelationIdRef{key: {Value: &CorrelationId{Description: key, Location: "$message.header#/id"}}},
			ExternalDocs:    map[string]*ExternalDocRef{key: {Value: fullExternalDoc(key)}},
			OperationTraits: map[string]*OperationTraitRef{key: {Value: fullOperationTrait(key, channel)}},
			MessageTraits:   map[string]*MessageTraitRef{key: {Value: fullMessageTrait(key)}},
			ServerVariables: map[string]*ServerVariableRef{key: {Value: fullServerVariable()}},
		},
	}
}

func reference[T any]() dynamic.Reference[T] {
	return dynamic.Reference[T]{Ref: "#/components/foo", Summary: "summary", Description: "description"}
}

func fullServer(key string) *Server {
	return &Server{
		Host:            "localhost:9092",
		Pathname:        "/" + key,
		Title:           key + " title",
		Summary:         key + " summary",
		Description:     key + " description",
		Protocol:        "kafka",
		ProtocolVersion: "3.0",
		Variables:       map[string]*ServerVariableRef{key: {Value: fullServerVariable()}},
		Tags:            []*TagRef{{Value: fullTag(key)}},
		Bindings: &ServerBindings{Kafka: BrokerBindings{
			configs:                      map[string]any{"log.retention.bytes": 1},
			LogRetentionBytes:            1,
			LogRetentionMs:               2,
			LogRetentionCheckIntervalMs:  3,
			LogSegmentDeleteDelayMs:      4,
			LogRollMs:                    5,
			LogSegmentBytes:              6,
			GroupInitialRebalanceDelayMs: 7,
			GroupMinSessionTimeoutMs:     8,
			SchemaRegistryUrl:            "https://registry.io",
			SchemaRegistryVendor:         "confluent",
		}},
		ExternalDocs: []ExternalDocRef{{Value: fullExternalDoc(key)}},
	}
}

func fullServerVariable() *ServerVariable {
	return &ServerVariable{Description: "description", Enum: []string{"a", "b"}, Default: "a", Examples: []string{"a"}}
}

func fullTag(key string) *Tag {
	return &Tag{Name: key, Description: key + " description", ExternalDocs: []*ExternalDocRef{{Value: fullExternalDoc(key)}}}
}

func fullExternalDoc(key string) *ExternalDoc {
	return &ExternalDoc{Description: key + " docs", Url: "https://docs.io"}
}

func fullParameter() *Parameter {
	return &Parameter{Description: "description", Enum: []string{"a", "b"}, Default: "a", Examples: []string{"a"}, Location: "$message.payload#/id"}
}

func fullChannel(key string, server *Server) *Channel {
	return &Channel{
		Name:        key,
		Title:       key + " title",
		Address:     key + "/{id}",
		Summary:     key + " summary",
		Description: key + " description",
		Servers:     []*ServerRef{{Reference: reference[*ServerRef](), Value: server}},
		Messages:    map[string]*MessageRef{key: {Reference: reference[*MessageRef](), Value: fullMessage(key)}},
		Parameters:  map[string]*ParameterRef{"id": {Value: fullParameter()}},
		Bindings: &ChannelBindings{
			Kafka: TopicBindings{
				configs:               map[string]any{"partitions": 3},
				Partitions:            3,
				RetentionBytes:        1,
				RetentionMs:           2,
				SegmentBytes:          3,
				SegmentMs:             4,
				ValueSchemaValidation: true,
				KeySchemaValidation:   true,
			},
			Mqtt: MqttChannelBindings{configs: map[string]any{"qos": 1}},
			Websocket: WebsocketChannelBindings{
				Method:  "GET",
				Query:   schematest.New("object", schematest.WithProperty("q", schematest.New("string"))),
				Headers: schematest.New("object", schematest.WithProperty("h", schematest.New("string"))),
			},
		},
		Tags:         []*TagRef{{Value: fullTag(key)}},
		ExternalDocs: []*ExternalDocRef{{Value: fullExternalDoc(key)}},
	}
}

func fullMessage(key string) *Message {
	return &Message{
		Title:         key + " title",
		Name:          key,
		Summary:       key + " summary",
		Description:   key + " description",
		Deprecated:    true,
		CorrelationId: &CorrelationIdRef{Value: &CorrelationId{Description: key, Location: "$message.header#/id"}},
		ContentType:   "application/json",
		Headers:       &SchemaRef{Value: &openapi.Schema{Type: jsonSchema.Types{"object"}}},
		Payload: &SchemaRef{Reference: reference[*SchemaRef](), Value: schematest.New("object",
			schematest.WithProperty("id", schematest.New("string")),
			schematest.WithRequired("id"))},
		Bindings: &MessageBinding{
			Kafka: KafkaMessageBinding{
				Key:                     &SchemaRef{Value: &MultiSchemaFormat{Format: "application/schema+json;version=draft-07", Schema: &SchemaRef{Value: schematest.New("string")}}},
				SchemaIdLocation:        "payload",
				SchemaIdPayloadEncoding: "4",
			},
			Mqtt:      MqttMessageBindings{configs: map[string]any{"a": 1}},
			Websocket: WebsocketMessageBindings{configs: map[string]any{"a": 1}},
		},
		Traits:       []*MessageTraitRef{{Value: fullMessageTrait(key)}},
		Examples:     []any{"example"},
		ExternalDocs: []*ExternalDocRef{{Value: fullExternalDoc(key)}},
	}
}

func fullMessageTrait(key string) *MessageTrait {
	return &MessageTrait{
		Name:          key,
		Title:         key + " title",
		Summary:       key + " summary",
		Description:   key + " description",
		CorrelationId: &CorrelationIdRef{Value: &CorrelationId{Description: key, Location: "$message.header#/id"}},
		ContentType:   "application/json",
		Headers:       &SchemaRef{Value: schematest.New("object")},
		Bindings: &MessageBinding{
			Kafka:     KafkaMessageBinding{Key: &SchemaRef{Value: schematest.New("string")}, SchemaIdLocation: "header", SchemaIdPayloadEncoding: "0"},
			Mqtt:      MqttMessageBindings{configs: map[string]any{"a": 1}},
			Websocket: WebsocketMessageBindings{configs: map[string]any{"a": 1}},
		},
		Examples:     []*MessageExample{{Name: key, Summary: "summary", Headers: map[string]any{"a": 1}, Payload: "payload"}},
		ExternalDocs: []*ExternalDocRef{{Value: fullExternalDoc(key)}},
	}
}

func fullOperation(key, action string, channel *Channel, message *Message) *Operation {
	return &Operation{
		Action:       action,
		Channel:      &ChannelRef{Reference: reference[*ChannelRef](), Value: channel},
		Title:        key + " title",
		Summary:      key + " summary",
		Description:  key + " description",
		Bindings:     fullOperationBindings(),
		Traits:       []*OperationTraitRef{{Value: fullOperationTrait(key, channel)}},
		Messages:     []*MessageRef{{Value: message}},
		ExternalDocs: []*ExternalDocRef{{Value: fullExternalDoc(key)}},
	}
}

func fullOperationBindings() *OperationBindings {
	return &OperationBindings{
		Kafka:     KafkaOperationBinding{GroupId: schematest.New("string"), ClientId: schematest.New("string")},
		Mqtt:      MqttOperationBindings{configs: map[string]any{"qos": 1}},
		Websocket: WebsocketOperationBindings{configs: map[string]any{"a": 1}},
	}
}

func fullOperationTrait(key string, channel *Channel) *OperationTrait {
	return &OperationTrait{
		Channel:      &ChannelRef{Value: channel},
		Title:        key + " title",
		Summary:      key + " summary",
		Description:  key + " description",
		Bindings:     fullOperationBindings(),
		ExternalDocs: []*ExternalDocRef{{Value: fullExternalDoc(key)}},
	}
}

// requireAllFieldsSet fails if a field of an asyncapi3 type reachable from v has its zero value,
// so that a field added later is covered by the tests that use maximalConfig.
// Reference is not checked (every kind of reference has one) and neither is Channel.Config,
// the back pointer that Parse sets.
func requireAllFieldsSet(t *testing.T, v any) {
	t.Helper()
	checkFieldsSet(t, reflect.ValueOf(v), "Config", map[uintptr]bool{})
}

func checkFieldsSet(t *testing.T, v reflect.Value, path string, seen map[uintptr]bool) {
	t.Helper()
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() || seen[v.Pointer()] {
			return
		}
		seen[v.Pointer()] = true
		checkFieldsSet(t, v.Elem(), path, seen)
	case reflect.Interface:
		if !v.IsNil() {
			checkFieldsSet(t, v.Elem(), path, seen)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			checkFieldsSet(t, v.Index(i), fmt.Sprintf("%s[%d]", path, i), seen)
		}
	case reflect.Map:
		for it := v.MapRange(); it.Next(); {
			checkFieldsSet(t, it.Value(), fmt.Sprintf("%s[%v]", path, it.Key()), seen)
		}
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(sortedmap.LinkedHashMap[string, *ServerRef]{}) {
			for _, s := range v.Addr().Interface().(*sortedmap.LinkedHashMap[string, *ServerRef]).Values() {
				checkFieldsSet(t, reflect.ValueOf(s), path+"[server]", seen)
			}
			return
		}
		if !strings.HasPrefix(v.Type().PkgPath(), "mokapi/providers/asyncapi3") {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			name := path + "." + f.Name
			if f.Name == "Reference" || f.Type == reflect.TypeOf(&Config{}) {
				continue
			}
			require.False(t, v.Field(i).IsZero(), "%s.%s is not set in maximalConfig", v.Type().Name(), f.Name)
			checkFieldsSet(t, v.Field(i), name, seen)
		}
	}
}
