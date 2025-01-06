package asyncapi3_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
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
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Host)
				require.Equal(t, "description", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Host)
				require.Equal(t, "description", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server url overwrite",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "bar.foo")),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "bar.foo", result.Servers["foo"].Value.Host)
			},
		},
		{
			name: "patch extend servers",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("bar", "kafka", "bar.foo", asyncapi3test.WithServerDescription("other"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Servers, 2)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Host)
				require.Equal(t, "description", result.Servers["foo"].Value.Description)
				require.Equal(t, "bar.foo", result.Servers["bar"].Value.Host)
				require.Equal(t, "other", result.Servers["bar"].Value.Description)
			},
		},
		{
			name: "patch server description",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar")),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("mokapi"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Host)
				require.Equal(t, "mokapi", result.Servers["foo"].Value.Description)
			},
		},
		{
			name: "patch server description is overwritten",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("description"))),
				asyncapi3test.NewConfig(asyncapi3test.WithServer("foo", "kafka", "foo.bar", asyncapi3test.WithServerDescription("mokapi"))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Servers, 1)
				require.Equal(t, "foo.bar", result.Servers["foo"].Value.Host)
				require.Equal(t, "mokapi", result.Servers["foo"].Value.Description)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRetentionBytes)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRetentionBytes)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRetentionMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRetentionMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRetentionCheckIntervalMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRetentionCheckIntervalMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogSegmentDeleteDelayMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogSegmentDeleteDelayMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRollMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogRollMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogSegmentBytes)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.LogSegmentBytes)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.GroupInitialRebalanceDelayMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.GroupInitialRebalanceDelayMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.GroupMinSessionTimeoutMs)
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
				require.Equal(t, int64(1), result.Servers["foo"].Value.Bindings.Kafka.GroupMinSessionTimeoutMs)
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
					asyncapi3test.WithKafkaChannelBinding(asyncapi3.TopicBindings{Partitions: 1}),
				)),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Equal(t, 1, result.Channels["foo"].Value.Bindings.Kafka.Partitions)
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
				r := msg.Payload.Value.Schema.(*schema.Ref)
				require.Equal(t, "string", r.Value.Type.String())
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
				s := result.Components.Schemas["foo"].Value.Schema.(*schema.Schema)
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
				s := result.Components.Schemas["foo"].Value.Schema.(*schema.Schema)
				require.Equal(t, "number", s.Type.String())
				s = result.Components.Schemas["bar"].Value.Schema.(*schema.Schema)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "patch schema",
			configs: []*asyncapi3.Config{
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.NewRef("number"))),
				asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.NewRef("number", schematest.WithFormat("double")))),
			},
			test: func(t *testing.T, result *asyncapi3.Config) {
				require.Len(t, result.Components.Schemas, 1)
				s := result.Components.Schemas["foo"].Value.Schema.(*schema.Ref)
				require.Equal(t, "number", s.Type())
				require.Equal(t, "double", s.Value.Format)
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
