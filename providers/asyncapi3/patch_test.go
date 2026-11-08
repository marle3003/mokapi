package asyncapi3_test

import (
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Patch_Info(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "patch description, version, terms and license",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "", "")),
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithInfoExt("term", "licName", "foo.bar")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
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
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "bar", "1.0"),
					asyncapi3test.WithInfoExt("term", "licName", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithInfo("foo", "other", "2.0"),
					asyncapi3test.WithInfoExt("foo", "otherName", "bar.foo")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, "other", result.Info.Description)
				require.Equal(t, "2.0", result.Info.Version)
				require.Equal(t, "foo", result.Info.TermsOfService)
				require.Equal(t, "otherName", result.Info.License.Name)
				require.Equal(t, "bar.foo", result.Info.License.Url)
			},
		},
		{
			name: "patch contact",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithContact("foo", "foo.bar", "info@foo.bar")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch contact",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithContact("", "", "")),
				asyncapi3test.NewConfig(asyncapi3test.WithContact("foo", "foo.bar", "info@foo.bar")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.NotNil(t, result.Info.Contact)
				require.Equal(t, "foo", result.Info.Contact.Name)
				require.Equal(t, "foo.bar", result.Info.Contact.Url)
				require.Equal(t, "info@foo.bar", result.Info.Contact.Email)
			},
		},
		{
			name: "patch contact is overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithContact("foo", "foo.bar", "info@foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithContact("mokapi", "mokapi.io", "info@mokapi.io")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
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
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "patch without server",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
				asyncapi3test.NewConfig(),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, result.Servers.Len(), 1)
				require.Equal(t, "foo.bar", result.Servers.Lookup("foo").Value.Host)
				require.Equal(t, "description", result.Servers.Lookup("foo").Value.Description)
			},
		},
		{
			name: "patch server",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, result.Servers.Len(), 1)
				require.Equal(t, "foo.bar", result.Servers.Lookup("foo").Value.Host)
				require.Equal(t, "description", result.Servers.Lookup("foo").Value.Description)
			},
		},
		{
			name: "patch server url overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "bar.foo")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, result.Servers.Len(), 1)
				require.Equal(t, "bar.foo", result.Servers.Lookup("foo").Value.Host)
			},
		},
		{
			name: "patch extend servers",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("bar", "kafka", "bar.foo", asyncapi3test.WithServerDescription("other"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, result.Servers.Len(), 2)
				require.Equal(t, "foo.bar", result.Servers.Lookup("foo").Value.Host)
				require.Equal(t, "description", result.Servers.Lookup("foo").Value.Description)
				require.Equal(t, "bar.foo", result.Servers.Lookup("bar").Value.Host)
				require.Equal(t, "other", result.Servers.Lookup("bar").Value.Description)
			},
		},
		{
			name: "patch server description",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("mokapi"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, result.Servers.Len(), 1)
				require.Equal(t, "foo.bar", result.Servers.Lookup("foo").Value.Host)
				require.Equal(t, "mokapi", result.Servers.Lookup("foo").Value.Description)
			},
		},
		{
			name: "patch server description is overwritten",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("mokapi"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, result.Servers.Len(), 1)
				require.Equal(t, "foo.bar", result.Servers.Lookup("foo").Value.Host)
				require.Equal(t, "mokapi", result.Servers.Lookup("foo").Value.Description)
			},
		},
		{
			name: "patch server bindings LogRetentionBytes",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionBytes: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRetentionBytes)
			},
		},
		{
			name: "patch server bindings LogRetentionBytes overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionBytes: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionBytes: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRetentionBytes)
			},
		},
		{
			name: "patch server bindings LogRetentionMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionMs: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRetentionMs)
			},
		},
		{
			name: "patch server bindings LogRetentionMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRetentionMs)
			},
		},
		{
			name: "patch server bindings LogRetentionCheckIntervalMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionCheckIntervalMs: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRetentionCheckIntervalMs)
			},
		},
		{
			name: "patch server bindings LogRetentionCheckIntervalMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionCheckIntervalMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRetentionCheckIntervalMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRetentionCheckIntervalMs)
			},
		},
		{
			name: "patch server bindings LogSegmentDeleteDelayMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogSegmentDeleteDelayMs: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogSegmentDeleteDelayMs)
			},
		},
		{
			name: "patch server bindings LogSegmentDeleteDelayMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogSegmentDeleteDelayMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogSegmentDeleteDelayMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogSegmentDeleteDelayMs)
			},
		},
		{
			name: "patch server bindings LogRollMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRollMs: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRollMs)
			},
		},
		{
			name: "patch server bindings LogRollMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRollMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogRollMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogRollMs)
			},
		},
		{
			name: "patch server bindings LogSegmentBytes",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogSegmentBytes: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogSegmentBytes)
			},
		},
		{
			name: "patch server bindings LogSegmentBytes overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogSegmentBytes: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{LogSegmentBytes: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.LogSegmentBytes)
			},
		},
		{
			name: "patch server bindings GroupInitialRebalanceDelayMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{GroupInitialRebalanceDelayMs: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.GroupInitialRebalanceDelayMs)
			},
		},
		{
			name: "patch server bindings GroupInitialRebalanceDelayMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{GroupInitialRebalanceDelayMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{GroupInitialRebalanceDelayMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.GroupInitialRebalanceDelayMs)
			},
		},
		{
			name: "patch server bindings GroupMinSessionTimeoutMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{GroupMinSessionTimeoutMs: 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.GroupMinSessionTimeoutMs)
			},
		},
		{
			name: "patch server bindings GroupMinSessionTimeoutMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{GroupMinSessionTimeoutMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar",
					asyncapi3test.WithKafkaServerBinding(asyncapi3.BrokerBindings{GroupMinSessionTimeoutMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Servers.Lookup("foo").Value.Bindings.Kafka.GroupMinSessionTimeoutMs)
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
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "patch description",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithChannelDescription("bar"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Channels, 1)
				require.Equal(t, "bar", result.Channels["foo"].Value.Description)
			},
		},
		{
			name: "add channel",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithChannelDescription("bar"))),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("bar",
					asyncapi3test.WithChannelDescription("foo"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Channels, 2)
				require.Equal(t, "bar", result.Channels["foo"].Value.Description)
				require.Equal(t, "foo", result.Channels["bar"].Value.Description)
			},
		},
		{
			name: "add channel Partition",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 5}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, 5, result.Channels["foo"].Value.Bindings.Kafka.Partitions)
			},
		},
		{
			name: "add channel Partition",
			configs: []*asyncapi3.Config{
				{
					Channels: map[string]*asyncapi3.ChannelRef{
						"foo": {
							Value: &asyncapi3.Channel{},
						},
					},
				},
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 5}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, 5, result.Channels["foo"].Value.Bindings.Kafka.Partitions)
			},
		},
		{
			name: "add channel Partition overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, 1, result.Channels["foo"].Value.Bindings.Kafka.Partitions)
			},
		},
		{
			name: "add channel RetentionBytes",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{RetentionBytes: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.RetentionBytes)
			},
		},
		{
			name: "add channel RetentionBytes overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{RetentionBytes: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{RetentionBytes: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.RetentionBytes)
			},
		},
		{
			name: "add channel RetentionMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{RetentionMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.RetentionMs)
			},
		},
		{
			name: "add channel RetentionMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{RetentionMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{RetentionMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.RetentionMs)
			},
		},
		{
			name: "add channel SegmentBytes",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{SegmentBytes: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.SegmentBytes)
			},
		},
		{
			name: "add channel SegmentBytes overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{SegmentBytes: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{SegmentBytes: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.SegmentBytes)
			},
		},
		{
			name: "add channel SegmentMs",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{SegmentMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.SegmentMs)
			},
		},
		{
			name: "add channel SegmentMs overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{SegmentMs: 5}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{SegmentMs: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, int64(1), result.Channels["foo"].Value.Bindings.Kafka.SegmentMs)
			},
		},
		{
			name: "add channel ValueSchemaValidation",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo")),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{ValueSchemaValidation: true}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, true, result.Channels["foo"].Value.Bindings.Kafka.ValueSchemaValidation)
			},
		},
		{
			name: "add channel ValueSchemaValidation overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{ValueSchemaValidation: true}),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{ValueSchemaValidation: false}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, false, result.Channels["foo"].Value.Bindings.Kafka.ValueSchemaValidation)
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
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "patch info",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo"))),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithMessageInfo("name", "title", "summary", "description"),
						asyncapi3test.WithContentType("application/json")),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				msg := result.Channels["foo"].Value.Messages["foo"].Value
				require.Equal(t, "name", msg.Name)
				require.Equal(t, "title", msg.Title)
				require.Equal(t, "summary", msg.Summary)
				require.Equal(t, "description", msg.Description)
				require.Equal(t, "application/json", msg.ContentType)
			},
		},
		{
			name: "patch payload",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo"),
				)),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithMessage("foo",
						asyncapi3test.WithPayload(schematest.New("string")),
					))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				msg := result.Channels["foo"].Value.Messages["foo"].Value
				r := msg.Payload.Value.(*schema.Schema)
				require.Equal(t, "string", r.Type.String())
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
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "patch add server",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				{Components: &asyncapi3.Components{Servers: map[string]*asyncapi3.ServerRef{"foo": {Value: &asyncapi3.Server{Protocol: "kafka"}}}}},
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Servers, 1)
				s := result.Components.Servers["foo"]
				require.Equal(t, "kafka", s.Value.Protocol)
			},
		},
		{
			name: "patch server",
			configs: []*asyncapi3.Config{
				{Components: &asyncapi3.Components{Servers: map[string]*asyncapi3.ServerRef{"foo": {Value: &asyncapi3.Server{Protocol: "kafka"}}}}},
				{Components: &asyncapi3.Components{Servers: map[string]*asyncapi3.ServerRef{"foo": {Value: &asyncapi3.Server{Protocol: "kafka", ProtocolVersion: "1.0"}}}}},
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Servers, 1)
				s := result.Components.Servers["foo"].Value
				require.Equal(t, "kafka", s.Protocol)
				require.Equal(t, "1.0", s.ProtocolVersion)
			},
		},
		{
			name: "patch add schema",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.New("number"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Schemas, 1)
				s := result.Components.Schemas["foo"].Value.(*schema.Schema)
				require.Equal(t, "number", s.Type.String())
			},
		},
		{
			name: "patch add additional schema",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.New("number"))),
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("bar", schematest.New("string"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Schemas, 2)
				s := result.Components.Schemas["foo"].Value.(*schema.Schema)
				require.Equal(t, "number", s.Type.String())
				s = result.Components.Schemas["bar"].Value.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "patch schema",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.New("number"))),
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.New("number", schematest.WithFormat("double")))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Schemas, 1)
				s := result.Components.Schemas["foo"].Value.(*schema.Schema)
				require.Equal(t, "number", s.Type.String())
				require.Equal(t, "double", s.Format)
			},
		},
		{
			name: "patch add message",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithComponentMessage("foo", asyncapi3test.NewMessage(asyncapi3test.WithMessageInfo("name", "", "", "")))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Messages, 1)
				msg := result.Components.Messages["foo"].Value
				require.Equal(t, "name", msg.Name)
			},
		},
		{
			name: "patch message",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithComponentMessage("foo", asyncapi3test.NewMessage(asyncapi3test.WithMessageInfo("name", "", "", "")))),
				asyncapi3test.NewConfig(asyncapi3test.WithComponentMessage("foo", asyncapi3test.NewMessage(asyncapi3test.WithMessageInfo("", "title", "", "")))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Messages, 1)
				msg := result.Components.Messages["foo"].Value
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

func TestConfig_Patch_HeaderFieldsOfFirstConfigWin(t *testing.T) {
	first := &asyncapi3.Config{Version: "3.0.0", Id: "first", DefaultContentType: "application/xml", Info: asyncapi3.Info{Name: "foo"}}
	second := &asyncapi3.Config{Version: "3.1.0", Id: "second", DefaultContentType: "application/json", Info: asyncapi3.Info{Name: "foo"}}

	result := &asyncapi3.Config{}
	result.Patch(first)
	result.Patch(second)

	require.Equal(t, "3.0.0", result.Version)
	require.Equal(t, "first", result.Id)
	require.Equal(t, "application/xml", result.DefaultContentType)
	require.Equal(t, "foo", result.Info.Name)
}

func TestConfig_Patch_OneSideWithoutValue(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "contact only in first config",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithContact("foo", "foo.bar", "info@foo.bar")),
				asyncapi3test.NewConfig(),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, "foo", result.Info.Contact.Name)
			},
		},
		{
			name: "license only in first config",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithInfoExt("", "MIT", "foo.bar")),
				asyncapi3test.NewConfig(),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, "MIT", result.Info.License.Name)
			},
		},
		{
			name: "kafka topic configs only in patch config",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 3}))),
				asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
					asyncapi3test.WithKafkaChannelBindingFrom(map[string]any{"retention.ms": 1}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				b := result.Channels["foo"].Value.Bindings.Kafka
				require.True(t, b.IsSet())
				require.Equal(t, int64(1), b.RetentionMs)
				require.Equal(t, 1, b.Partitions)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				require.NotPanics(t, func() { c.Patch(p) })
			}
			tc.test(t, c)
		})
	}
}

func TestConfig_Patch_Operations(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*asyncapi3.Config
		test    func(t *testing.T, result *asyncapi3.Config)
	}{
		{
			name: "add operation",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("foo",
					asyncapi3test.WithOperationAction("send"), asyncapi3test.WithOperationTitle("bar"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Operations, 1)
				require.Equal(t, "bar", result.Operations["foo"].Value.Title)
			},
		},
		{
			name: "add additional operation",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("foo", asyncapi3test.WithOperationAction("send"))),
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("bar", asyncapi3test.WithOperationAction("receive"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Operations, 2)
				require.Equal(t, "send", result.Operations["foo"].Value.Action)
				require.Equal(t, "receive", result.Operations["bar"].Value.Action)
			},
		},
		{
			name: "patch operation",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("foo",
					asyncapi3test.WithOperationAction("send"), asyncapi3test.WithOperationTitle("title"))),
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("foo",
					asyncapi3test.WithOperationAction("send"),
					asyncapi3test.WithOperationSummary("summary"), asyncapi3test.WithOperationDescription("description"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Operations, 1)
				op := result.Operations["foo"].Value
				require.Equal(t, "title", op.Title)
				require.Equal(t, "summary", op.Summary)
				require.Equal(t, "description", op.Description)
			},
		},
		{
			name: "patch operation binding",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("foo", asyncapi3test.WithOperationAction("send"))),
				asyncapi3test.NewConfig(asyncapi3test.WithOperation("foo", asyncapi3test.WithOperationAction("send"),
					asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{ClientId: schematest.New("string")}))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, "string", result.Operations["foo"].Value.Bindings.Kafka.ClientId.Type.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

// Patch adds the entries of every Components map, not only servers, schemas and messages.
func TestConfig_Patch_ComponentsOfAllKinds(t *testing.T) {
	newConfig := func(key string) *asyncapi3.Config {
		return asyncapi3test.NewConfig(
			asyncapi3test.WithComponentServer(key, &asyncapi3.Server{Host: key}),
			asyncapi3test.WithComponentTag(key, &asyncapi3.Tag{Name: key}),
			asyncapi3test.WithComponentChannel(key, &asyncapi3.Channel{Address: key}),
			asyncapi3test.WithComponentSchema(key, schematest.New("string")),
			asyncapi3test.WithComponentMessage(key, &asyncapi3.Message{Name: key}),
			asyncapi3test.WithComponentOperation(key, &asyncapi3.Operation{Title: key}),
			asyncapi3test.WithComponentParameter(key, &asyncapi3.Parameter{Default: key}),
			asyncapi3test.WithComponentCorrelationId(key, &asyncapi3.CorrelationId{Location: key}),
			asyncapi3test.WithComponentExternalDoc(key, &asyncapi3.ExternalDoc{Url: key}),
			asyncapi3test.WithComponentOperationTrait(key, &asyncapi3.OperationTrait{Title: key}),
			asyncapi3test.WithComponentMessageTrait(key, &asyncapi3.MessageTrait{Name: key}),
		)
	}
	result := &asyncapi3.Config{}

	result.Patch(newConfig("a"))
	result.Patch(newConfig("b"))

	c := result.Components
	require.Equal(t, "b", c.Servers["b"].Value.Host)
	require.Equal(t, "b", c.Tags["b"].Value.Name)
	require.Equal(t, "b", c.Channels["b"].Value.Address)
	require.NotNil(t, c.Schemas["b"].Value)
	require.Equal(t, "b", c.Messages["b"].Value.Name)
	require.Equal(t, "b", c.Operations["b"].Value.Title)
	require.Equal(t, "b", c.Parameters["b"].Value.Default)
	require.Equal(t, "b", c.CorrelationIds["b"].Value.Location)
	require.Equal(t, "b", c.ExternalDocs["b"].Value.Url)
	require.Equal(t, "b", c.OperationTraits["b"].Value.Title)
	require.Equal(t, "b", c.MessageTraits["b"].Value.Name)
	for _, m := range []int{len(c.Servers), len(c.Tags), len(c.Channels), len(c.Schemas), len(c.Messages), len(c.Operations),
		len(c.Parameters), len(c.CorrelationIds), len(c.ExternalDocs), len(c.OperationTraits), len(c.MessageTraits)} {
		require.Equal(t, 2, m)
	}
}

func TestConfig_Patch_NilEntriesInPatchConfig(t *testing.T) {
	result := asyncapi3test.NewConfig(
		asyncapi3test.WithServer("foo", "kafka", "foo.bar"),
		asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("foo")),
		asyncapi3test.WithOperation("foo", asyncapi3test.WithOperationAction("send")),
	)
	patch := asyncapi3test.NewConfig()
	patch.Servers.Set("foo", nil)
	patch.Channels = map[string]*asyncapi3.ChannelRef{"foo": {}}
	patch.Operations = map[string]*asyncapi3.OperationRef{"foo": nil}

	require.NotPanics(t, func() { result.Patch(patch) })

	require.Equal(t, "foo.bar", result.Servers.Lookup("foo").Value.Host)
	require.Contains(t, result.Channels["foo"].Value.Messages, "foo")
	require.Equal(t, "send", result.Operations["foo"].Value.Action)
}

func TestConfig_Patch_PayloadOfOtherSchemaFormat(t *testing.T) {
	multi := func() *asyncapi3.SchemaRef {
		return &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{
			Format: "application/schema+json;version=draft-07",
			Schema: &asyncapi3.SchemaRef{Value: schematest.New("string")},
		}}
	}
	first := asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
		asyncapi3test.WithMessage("foo", asyncapi3test.WithPayload(schematest.New("number")))))
	patch := asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo",
		asyncapi3test.WithMessage("foo", asyncapi3test.UsePayload(multi()))))

	first.Patch(patch)

	payload := first.Channels["foo"].Value.Messages["foo"].Value.Payload
	require.Equal(t, multi(), payload)
	require.NotSame(t, patch.Channels["foo"].Value.Messages["foo"].Value.Payload, payload)
}

func TestMultiSchemaFormat_Patch(t *testing.T) {
	newPatch := func() *asyncapi3.MultiSchemaFormat {
		return &asyncapi3.MultiSchemaFormat{
			Format: "application/schema+json;version=draft-07",
			Schema: &asyncapi3.SchemaRef{Value: schematest.New("object", schematest.WithProperty("foo", schematest.New("string")))},
		}
	}
	patch := newPatch()

	m := &asyncapi3.MultiSchemaFormat{}
	m.Patch(patch)
	m.Patch(&asyncapi3.MultiSchemaFormat{Schema: &asyncapi3.SchemaRef{Value: schematest.New("object",
		schematest.WithProperty("bar", schematest.New("string")))}})
	m.Schema.Value.(*schema.Schema).Properties.Get("foo").Type[0] = "changed"

	require.Equal(t, patch.Format, m.Format)
	require.Equal(t, 2, m.Schema.Value.(*schema.Schema).Properties.Len())
	require.Equal(t, newPatch(), patch, "the patch is not changed")
}

func TestConfig_Patch_RefWithoutValue(t *testing.T) {
	result := asyncapi3test.NewConfig()
	result.Channels = map[string]*asyncapi3.ChannelRef{"foo": {}}
	patch := asyncapi3test.NewConfig(asyncapi3test.WithChannel("foo", asyncapi3test.WithChannelDescription("bar")))
	result.Components = &asyncapi3.Components{Messages: map[string]*asyncapi3.MessageRef{"foo": {}}}
	patch.Components = &asyncapi3.Components{Messages: map[string]*asyncapi3.MessageRef{"foo": {Value: &asyncapi3.Message{Name: "name"}}}}

	result.Patch(patch)

	require.Equal(t, "bar", result.Channels["foo"].Value.Description)
	require.NotSame(t, patch.Channels["foo"].Value, result.Channels["foo"].Value)
	require.Equal(t, "name", result.Components.Messages["foo"].Value.Name)
	require.NotSame(t, patch.Components.Messages["foo"].Value, result.Components.Messages["foo"].Value)
}

func TestConfig_Patch_KafkaBindingSchemas(t *testing.T) {
	first := asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("foo")),
		asyncapi3test.WithOperation("foo", asyncapi3test.WithOperationAction("send")),
	)
	key := schematest.New("string")
	patch := asyncapi3test.NewConfig(
		asyncapi3test.WithChannel("foo", asyncapi3test.WithMessage("foo",
			asyncapi3test.WithKafkaMessageBinding(asyncapi3.KafkaMessageBinding{Key: &asyncapi3.SchemaRef{Value: key}}))),
		asyncapi3test.WithOperation("foo", asyncapi3test.WithOperationAction("send"),
			asyncapi3test.WithOperationBinding(asyncapi3.KafkaOperationBinding{GroupId: schematest.New("string"), ClientId: schematest.New("string")})),
	)
	first.Channels["foo"].Value.Messages["foo"].Value.Bindings = &asyncapi3.MessageBinding{}
	first.Operations["foo"].Value.Bindings = &asyncapi3.OperationBindings{}

	first.Patch(patch)

	msgKey := first.Channels["foo"].Value.Messages["foo"].Value.Bindings.Kafka.Key.Value.(*schema.Schema)
	require.Equal(t, key, msgKey)
	require.NotSame(t, key, msgKey)
	op := first.Operations["foo"].Value.Bindings.Kafka
	require.Equal(t, "string", op.GroupId.Type.String())
	require.Equal(t, "string", op.ClientId.Type.String())
	require.NotSame(t, patch.Operations["foo"].Value.Bindings.Kafka.GroupId, op.GroupId)
}
