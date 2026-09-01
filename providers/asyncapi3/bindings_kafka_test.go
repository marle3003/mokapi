package asyncapi3_test

import (
	"encoding/json"
	"mokapi/providers/asyncapi3"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestServerBindings_Kafka_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.ServerBindings
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "kafka default",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{}`, json)
				require.Equal(t, "{}\n", yaml)
			},
		},
		{
			name: "log.retention.bytes",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					LogRetentionBytes: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"log.retention.bytes":5}}`, json)
				require.Equal(t, "kafka:\n    log.retention.bytes: 5\n", yaml)
			},
		},
		{
			name: "log.retention",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					LogRetentionMs: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"log.retention":5}}`, json)
				require.Equal(t, "kafka:\n    log.retention: 5\n", yaml)
			},
		},
		{
			name: "log.retention.check.interval.ms",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					LogRetentionCheckIntervalMs: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"log.retention.check.interval.ms":5}}`, json)
				require.Equal(t, "kafka:\n    log.retention.check.interval.ms: 5\n", yaml)
			},
		},
		{
			name: "log.segment.delete.delay.ms",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					LogSegmentDeleteDelayMs: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"log.segment.delete.delay.ms":5}}`, json)
				require.Equal(t, "kafka:\n    log.segment.delete.delay.ms: 5\n", yaml)
			},
		},
		{
			name: "log.roll",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					LogRollMs: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"log.roll":5}}`, json)
				require.Equal(t, "kafka:\n    log.roll: 5\n", yaml)
			},
		},
		{
			name: "log.segment.bytes",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					LogSegmentBytes: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"log.segment.bytes":5}}`, json)
				require.Equal(t, "kafka:\n    log.segment.bytes: 5\n", yaml)
			},
		},
		{
			name: "group.initial.rebalance.delay.ms",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					GroupInitialRebalanceDelayMs: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"group.initial.rebalance.delay.ms":5}}`, json)
				require.Equal(t, "kafka:\n    group.initial.rebalance.delay.ms: 5\n", yaml)
			},
		},
		{
			name: "group.min.session.timeout.ms",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					GroupMinSessionTimeoutMs: 5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"group.min.session.timeout.ms":5}}`, json)
				require.Equal(t, "kafka:\n    group.min.session.timeout.ms: 5\n", yaml)
			},
		},
		{
			name: "schemaRegistryUrl",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					SchemaRegistryUrl: "foo.bar",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"schemaRegistryUrl":"foo.bar"}}`, json)
				require.Equal(t, "kafka:\n    schemaRegistryUrl: foo.bar\n", yaml)
			},
		},
		{
			name: "schemaRegistryVendor",
			s: &asyncapi3.ServerBindings{
				Kafka: asyncapi3.BrokerBindings{
					SchemaRegistryVendor: "foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"schemaRegistryVendor":"foo"}}`, json)
				require.Equal(t, "kafka:\n    schemaRegistryVendor: foo\n", yaml)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb, err := json.Marshal(tc.s)
			if err != nil {
				tc.test(t, "", "", err)
			}
			yb, err := yaml.Marshal(tc.s)
			tc.test(t, string(jb), string(yb), err)
		})
	}
}

func TestChannelBindings_Kafka_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.ChannelBindings
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "kafka default",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					ValueSchemaValidation: true,
					KeySchemaValidation:   true,
					Partitions:            1,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{}`, json)
				require.Equal(t, "{}\n", yaml)
			},
		},
		{
			name: "partition",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					ValueSchemaValidation: true,
					KeySchemaValidation:   true,
					Partitions:            5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"partitions":5}}`, json)
				require.Equal(t, "kafka:\n    partitions: 5\n", yaml)
			},
		},
		{
			name: "default partition set by config",
			s: func() *asyncapi3.ChannelBindings {
				tb := &asyncapi3.TopicBindings{}
				err := json.Unmarshal([]byte(`{"partitions":1}`), &tb)
				require.NoError(t, err)
				return &asyncapi3.ChannelBindings{Kafka: *tb}
			}(),
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"partitions":1}}`, json)
				require.Equal(t, "kafka:\n    partitions: 1\n", yaml)
			},
		},
		{
			name: "retention.bytes",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					Partitions:            1,
					ValueSchemaValidation: true,
					KeySchemaValidation:   true,
					RetentionBytes:        5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"retention.bytes":5}}`, json)
				require.Equal(t, "kafka:\n    retention.bytes: 5\n", yaml)
			},
		},
		{
			name: "retention.ms",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					Partitions:            1,
					ValueSchemaValidation: true,
					KeySchemaValidation:   true,
					RetentionMs:           5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"retention.ms":5}}`, json)
				require.Equal(t, "kafka:\n    retention.ms: 5\n", yaml)
			},
		},
		{
			name: "segment.bytes",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					Partitions:            1,
					ValueSchemaValidation: true,
					KeySchemaValidation:   true,
					SegmentBytes:          5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"segment.bytes":5}}`, json)
				require.Equal(t, "kafka:\n    segment.bytes: 5\n", yaml)
			},
		},
		{
			name: "segment.ms",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					Partitions:            1,
					ValueSchemaValidation: true,
					KeySchemaValidation:   true,
					SegmentMs:             5,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"segment.ms":5}}`, json)
				require.Equal(t, "kafka:\n    segment.ms: 5\n", yaml)
			},
		},
		{
			name: "confluent.value.schema.validation",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					Partitions:            1,
					ValueSchemaValidation: false,
					KeySchemaValidation:   true,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"confluent.value.schema.validation":false}}`, json)
				require.Equal(t, "kafka:\n    confluent.value.schema.validation: false\n", yaml)
			},
		},
		{
			name: "confluent.key.schema.validation",
			s: &asyncapi3.ChannelBindings{
				Kafka: asyncapi3.TopicBindings{
					Partitions:            1,
					ValueSchemaValidation: true,
					KeySchemaValidation:   false,
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"kafka":{"confluent.key.schema.validation":false}}`, json)
				require.Equal(t, "kafka:\n    confluent.key.schema.validation: false\n", yaml)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb, err := json.Marshal(tc.s)
			if err != nil {
				tc.test(t, "", "", err)
			}
			yb, err := yaml.Marshal(tc.s)
			tc.test(t, string(jb), string(yb), err)
		})
	}
}
