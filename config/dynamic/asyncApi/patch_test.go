package asyncApi_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/asyncApi/kafka"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func TestConfig_Patch_Info(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncApi.Config
		test    func(t *testing.T, result *asyncApi.Config)
	}{
		{
			name: "patch description, version, terms and license",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "", "")),
				asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "bar", "1.0"),
					asyncapitest.WithInfoExt("term", "licName", "foo.bar")),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Equal(t, "bar", result.Info.Description)
				require.Equal(t, "1.0", result.Info.Version)
				require.NotNil(t, result.Info.License)
				require.Equal(t, "term", result.Info.TermsOfService)
				require.Equal(t, "licName", result.Info.License.Name)
				require.Equal(t, "foo.bar", result.Info.License.Url)
			},
		},
		{
			name: "patch description version is overwritten",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "bar", "1.0"),
					asyncapitest.WithInfoExt("term", "licName", "foo.bar")),
				asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "other", "2.0"),
					asyncapitest.WithInfoExt("foo", "otherName", "bar.foo")),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Equal(t, "other", result.Info.Description)
				require.Equal(t, "2.0", result.Info.Version)
				require.Equal(t, "foo", result.Info.TermsOfService)
				require.Equal(t, "otherName", result.Info.License.Name)
				require.Equal(t, "bar.foo", result.Info.License.Url)
			},
		},
		{
			name: "patch contact",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(),
				asyncapitest.NewConfig(asyncapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch contact",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithContact("", "", "")),
				asyncapitest.NewConfig(asyncapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch contact is overwrite",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithContact("foo", "foo.bar", "info@foo.bar")),
				asyncapitest.NewConfig(asyncapitest.WithContact("mokapi", "mokapi.io", "info@mokapi.io")),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "mokapi", result.Info.Contact.Name)
				require.Equal(t, "mokapi.io", result.Info.Contact.Url)
				require.Equal(t, "info@mokapi.io", result.Info.Contact.Email)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

func TestConfig_Patch_Server(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncApi.Config
		test    func(t *testing.T, result *asyncApi.Config)
	}{
		{
			name: "patch without server",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("description"))),
				asyncapitest.NewConfig(),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Url)
				require.Equal(t, "description", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("description"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Url)
				require.Equal(t, "description", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server url overwrite",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar")),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "bar.foo")),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "bar.foo", result.Servers["foo"].Value.Url)
			},
		},
		{
			name: "patch extend servers",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("description"))),
				asyncapitest.NewConfig(asyncapitest.WithServer("bar", "kafka", "bar.foo", asyncapitest.WithServerDescription("other"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Servers, 2)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Url)
				require.Equal(t, "description", result.Servers["foo"].Value.Description)
				require.Equal(t, "bar.foo", result.Servers["bar"].Value.Url)
				require.Equal(t, "other", result.Servers["bar"].Value.Description)
			},
		},
		{
			name: "patch server description",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar")),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("mokapi"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Url)
				require.Equal(t, "mokapi", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server description is overwritten",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("description"))),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar", asyncapitest.WithServerDescription("mokapi"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Url)
				require.Equal(t, "mokapi", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server bindings",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar")),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("foo", "bar"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				config := result.Servers["foo"].Value.Bindings.Kafka.Config
				require.Len(t, config, 1)
				require.Equal(t, "bar", config["foo"])
			},
		},
		{
			name: "patch server bindings empty value",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("foo", ""))),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("foo", "bar"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				config := result.Servers["foo"].Value.Bindings.Kafka.Config
				require.Len(t, config, 1)
				require.Equal(t, "bar", config["foo"])
			},
		},
		{
			name: "patch server bindings not overwrite",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("foo", "bar"))),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("foo", "12"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				config := result.Servers["foo"].Value.Bindings.Kafka.Config
				require.Len(t, config, 1)
				require.Equal(t, "bar", config["foo"])
			},
		},
		{
			name: "patch server bindings add",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("foo", "bar"))),
				asyncapitest.NewConfig(asyncapitest.WithServer("foo", "kafka", "foo.bar",
					asyncapitest.WithKafkaBinding("bar", "foo"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				config := result.Servers["foo"].Value.Bindings.Kafka.Config
				require.Len(t, config, 2)
				require.Equal(t, "bar", config["foo"])
				require.Equal(t, "foo", config["bar"])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

func TestConfig_Patch_Channel(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncApi.Config
		test    func(t *testing.T, result *asyncApi.Config)
	}{
		{
			name: "patch description",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithChannelDescription("bar"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Channels, 1)
				require.Equal(t, "bar", result.Channels["foo"].Value.Description)
			},
		},
		{
			name: "add channel",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithChannelDescription("bar"))),
				asyncapitest.NewConfig(asyncapitest.WithChannel("bar",
					asyncapitest.WithChannelDescription("foo"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Channels, 2)
				require.Equal(t, "bar", result.Channels["foo"].Value.Description)
				require.Equal(t, "foo", result.Channels["bar"].Value.Description)
			},
		},
		{
			name: "patch subscribe id, summary and description",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo")),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithSubscribe(
						asyncapitest.WithOperationInfo("c11", "foo", "bar")))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				ch := result.Channels["foo"]
				require.Equal(t, "c11", ch.Value.Subscribe.OperationId)
				require.Equal(t, "foo", ch.Value.Subscribe.Summary)
				require.Equal(t, "bar", ch.Value.Subscribe.Description)
			},
		},
		{
			name: "patch publish id, summary and description",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo")),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithPublish(
						asyncapitest.WithOperationInfo("c11", "foo", "bar")))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				ch := result.Channels["foo"]
				require.Equal(t, "c11", ch.Value.Publish.OperationId)
				require.Equal(t, "foo", ch.Value.Publish.Summary)
				require.Equal(t, "bar", ch.Value.Publish.Description)
			},
		},
		{
			name: "patch subscribe summary description",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithSubscribe())),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithSubscribe(
						asyncapitest.WithOperationInfo("c11", "foo", "bar")))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				ch := result.Channels["foo"]
				require.Equal(t, "c11", ch.Value.Subscribe.OperationId)
				require.Equal(t, "foo", ch.Value.Subscribe.Summary)
				require.Equal(t, "bar", ch.Value.Subscribe.Description)
			},
		},
		{
			name: "patch bindings",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithSubscribe())),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithChannelKafka(kafka.TopicBindings{Partitions: 10}))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				ch := result.Channels["foo"]
				require.Equal(t, 10, ch.Value.Bindings.Kafka.Partitions)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

func TestConfig_Patch_Message(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncApi.Config
		test    func(t *testing.T, result *asyncApi.Config)
	}{
		{
			name: "patch info",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithPublish(
						asyncapitest.WithMessage(),
					))),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithMessageInfo("name", "title", "summary", "description"),
							asyncapitest.WithMessageId("foo"),
							asyncapitest.WithContentType("application/json")),
					))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				msg := result.Channels["foo"].Value.Publish.Message.Value
				require.Equal(t, "foo", msg.MessageId)
				require.Equal(t, "name", msg.Name)
				require.Equal(t, "title", msg.Title)
				require.Equal(t, "summary", msg.Summary)
				require.Equal(t, "description", msg.Description)
				require.Equal(t, "application/json", msg.ContentType)
			},
		},
		{
			name: "patch payload",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithPublish(
						asyncapitest.WithMessage(),
					))),
				asyncapitest.NewConfig(asyncapitest.WithChannel("foo",
					asyncapitest.WithPublish(
						asyncapitest.WithMessage(
							asyncapitest.WithPayload(schematest.New("string")),
						)))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				msg := result.Channels["foo"].Value.Publish.Message.Value
				require.Equal(t, "string", msg.Payload.Value.Type.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

func TestConfig_Patch_Components(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncApi.Config
		test    func(t *testing.T, result *asyncApi.Config)
	}{
		{
			name: "patch add server",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(),
				{Components: &asyncApi.Components{Servers: map[string]*asyncApi.Server{"foo": {Protocol: "kafka"}}}},
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Components.Servers, 1)
				s := result.Components.Servers["foo"]
				require.Equal(t, "kafka", s.Protocol)
			},
		},
		{
			name: "patch server",
			configs: []*asyncApi.Config{
				{Components: &asyncApi.Components{Servers: map[string]*asyncApi.Server{"foo": {Protocol: "kafka"}}}},
				{Components: &asyncApi.Components{Servers: map[string]*asyncApi.Server{"foo": {Protocol: "kafka", ProtocolVersion: "1.0"}}}},
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Components.Servers, 1)
				s := result.Components.Servers["foo"]
				require.Equal(t, "kafka", s.Protocol)
				require.Equal(t, "1.0", s.ProtocolVersion)
			},
		},
		{
			name: "patch add schema",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(),
				asyncapitest.NewConfig(asyncapitest.WithSchemas("foo", schematest.New("number"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Equal(t, 1, result.Components.Schemas.Len())
				s := result.Components.Schemas.Get("foo")
				require.Equal(t, "number", s.Value.Type.String())
			},
		},
		{
			name: "patch add additional schema",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithSchemas("foo", schematest.New("number"))),
				asyncapitest.NewConfig(asyncapitest.WithSchemas("bar", schematest.New("string"))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Equal(t, 2, result.Components.Schemas.Len())
				s := result.Components.Schemas.Get("foo")
				require.Equal(t, "number", s.Value.Type.String())
				s = result.Components.Schemas.Get("bar")
				require.Equal(t, "string", s.Value.Type.String())
			},
		},
		{
			name: "patch schema",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithSchemas("foo", schematest.New("number"))),
				asyncapitest.NewConfig(asyncapitest.WithSchemas("foo", schematest.New("number", schematest.WithFormat("double")))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Equal(t, 1, result.Components.Schemas.Len())
				s := result.Components.Schemas.Get("foo")
				require.Equal(t, "number", s.Value.Type.String())
				require.Equal(t, "double", s.Value.Format)
			},
		},
		{
			name: "patch add message",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(),
				asyncapitest.NewConfig(asyncapitest.WithMessages("foo", asyncapitest.NewMessage(asyncapitest.WithMessageInfo("name", "", "", "")))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Components.Messages, 1)
				msg := result.Components.Messages["foo"]
				require.Equal(t, "name", msg.Name)
			},
		},
		{
			name: "patch message",
			configs: []*asyncApi.Config{
				asyncapitest.NewConfig(asyncapitest.WithMessages("foo", asyncapitest.NewMessage(asyncapitest.WithMessageInfo("name", "", "", "")))),
				asyncapitest.NewConfig(asyncapitest.WithMessages("foo", asyncapitest.NewMessage(asyncapitest.WithMessageInfo("", "title", "", "")))),
			},
			test: func(t *testing.T, result *asyncApi.Config) {
				require.Len(t, result.Components.Messages, 1)
				msg := result.Components.Messages["foo"]
				require.Equal(t, "title", msg.Title)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}
