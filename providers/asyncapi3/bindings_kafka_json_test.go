package asyncapi3_test

import (
	"encoding/json"
	"mokapi/providers/asyncapi3"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKafkaBindingsServer_JSON(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, config *asyncapi3.ServerBindings, err error)
	}{
		{
			name: "log.retention.bytes",
			config: `{
"kafka": { 
  "log.retention.bytes": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.LogRetentionBytes)
			},
		},
		{
			name: "log.retention.bytes error",
			config: `{
"kafka": {
  "log.retention.bytes": "foo"
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.retention.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.ms",
			config: `{
"kafka": {
  "log.retention.ms": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.LogRetentionMs)
			},
		},
		{
			name: "log.retention.ms error",
			config: `{
"kafka":{
  "log.retention.ms": "foo"
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.retention.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.minutes",
			config: `{
"kafka":{
  "log.retention.minutes": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(600000), config.Kafka.LogRetentionMs)
			},
		},
		{
			name: "log.retention.minutes error",
			config: `{
"kafka":{
  "log.retention.minutes": "foo"
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.retention.minutes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.hours",
			config: `{
"kafka":{
  "log.retention.hours": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(36000000), config.Kafka.LogRetentionMs)
			},
		},
		{
			name: "log.retention.hours error",
			config: `{
"kafka":{
  "log.retention.hours": "foo"
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.retention.hours: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.check.interval.ms",
			config: `{
"kafka":{
  "log.retention.check.interval.ms": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.LogRetentionCheckIntervalMs)
			},
		},
		{
			name: "log.retention.check.interval.ms error",
			config: `{
"kafka":{
  "log.retention.check.interval.ms": "foo"
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.retention.check.interval.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.segment.delete.delay.ms",
			config: `{
"kafka":{
  "log.segment.delete.delay.ms": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.LogSegmentDeleteDelayMs)
			},
		},
		{
			name: "log.segment.delete.delay.ms error",
			config: `{
"kafka":{
  "log.segment.delete.delay.ms": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.segment.delete.delay.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.roll.ms",
			config: `{
"kafka":{
  "log.roll.ms": 10
}}
`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.LogRollMs)
			},
		},
		{
			name: "log.roll.ms error",
			config: `{
"kafka":{
  "log.roll.ms": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.roll.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.roll.minutes",
			config: `{
"kafka":{
  "log.roll.minutes": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(600000), config.Kafka.LogRollMs)
			},
		},
		{
			name: "log.roll.minutes error",
			config: `{
"kafka":{
  "log.roll.minutes": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.roll.minutes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.roll.hours",
			config: `{
"kafka":{
  "log.roll.hours": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(36000000), config.Kafka.LogRollMs)
			},
		},
		{
			name: "log.roll.hours error",
			config: `{
"kafka":{
  "log.roll.hours": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.roll.hours: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.segment.bytes",
			config: `{
"kafka":{
  "log.segment.bytes": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.LogSegmentBytes)
			},
		},
		{
			name: "log.segment.bytes error",
			config: `{
"kafka":{
  "log.segment.bytes": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid log.segment.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "group.initial.rebalance.delay.ms",
			config: `{
"kafka":{
  "group.initial.rebalance.delay.ms": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.GroupInitialRebalanceDelayMs)
			},
		},
		{
			name: "group.initial.rebalance.delay.ms error",
			config: `{
"kafka":{
  "group.initial.rebalance.delay.ms": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid group.initial.rebalance.delay.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "group.min.session.timeout.ms",
			config: `{
"kafka":{
  "group.min.session.timeout.ms": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.GroupMinSessionTimeoutMs)
			},
		},
		{
			name: "group.min.session.timeout.ms error",
			config: `{
"kafka":{
  "group.min.session.timeout.ms": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.EqualError(t, err, "invalid group.min.session.timeout.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "schemaRegistryUrl",
			config: `{
"kafka":{
  "schemaRegistryUrl": "foo.bar"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.Equal(t, "foo.bar", config.Kafka.SchemaRegistryUrl)
			},
		},
		{
			name: "schemaRegistryVendor",
			config: `{
"kafka":{
  "schemaRegistryVendor": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ServerBindings, err error) {
				require.Equal(t, "foo", config.Kafka.SchemaRegistryVendor)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &asyncapi3.ServerBindings{}
			err := json.Unmarshal([]byte(tc.config), &cfg)

			tc.test(t, cfg, err)
		})
	}
}

func TestKafkaBindingsTopic_JSON(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, config *asyncapi3.ChannelBindings, err error)
	}{
		{
			name: "partitions",
			config: `{
"kafka":{
  "partitions": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, 10, config.Kafka.Partitions)
			},
		},
		{
			name: "partition not set",
			config: `{
"kafka":{
  "retention.bytes": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, config.Kafka.Partitions)
			},
		},
		{
			name: "partitions error",
			config: `{
"kafka":{
  "partitions": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid partition: cannot unmarshal string to int: foo")
			},
		},
		{
			name: "retention.bytes",
			config: `{
"kafka":{
  "retention.bytes": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.RetentionBytes)
			},
		},
		{
			name: "retention.bytes error",
			config: `{
"kafka":{
  "retention.bytes": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid retention.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "retention.ms",
			config: `{
"kafka":{
  "retention.ms": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.RetentionMs)
			},
		},
		{
			name: "retention.ms error",
			config: `{
"kafka":{
  "retention.ms": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid retention.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "segment.bytes",
			config: `{
"kafka":{
  "segment.bytes": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.SegmentBytes)
			},
		},
		{
			name: "segment.bytes error",
			config: `{
"kafka":{
  "segment.bytes": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid segment.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "segment.ms",
			config: `{
"kafka":{
  "segment.ms": 10
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Kafka.SegmentMs)
			},
		},
		{
			name: "segment.ms error",
			config: `{
"kafka":{
  "segment.ms": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid segment.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "confluent.value.schema.validation",
			config: `{
"kafka":{
  "confluent.value.schema.validation": false
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.False(t, config.Kafka.ValueSchemaValidation)
			},
		},
		{
			name: "confluent.value.schema.validation error",
			config: `{
"kafka":{
  "confluent.value.schema.validation": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid confluent.value.schema.validation: cannot unmarshal string to bool: foo")
			},
		},
		{
			name: "confluent.key.schema.validation",
			config: `{
"kafka":{
  "confluent.key.schema.validation": false
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.NoError(t, err)
				require.False(t, config.Kafka.KeySchemaValidation)
			},
		},
		{
			name: "confluent.key.schema.validation error",
			config: `{
"kafka":{
  "confluent.key.schema.validation": "foo"
}}`,
			test: func(t *testing.T, config *asyncapi3.ChannelBindings, err error) {
				require.EqualError(t, err, "invalid confluent.key.schema.validation: cannot unmarshal string to bool: foo")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &asyncapi3.ChannelBindings{}
			err := json.Unmarshal([]byte(tc.config), &cfg)

			tc.test(t, cfg, err)
		})
	}
}

func TestKafkaBindingsMessage_JSON(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, config *asyncapi3.MessageBinding, err error)
	}{
		{
			name: "schemaIdLocation",
			config: `{
"kafka":{
  "schemaIdLocation": "payload"
}}`,
			test: func(t *testing.T, config *asyncapi3.MessageBinding, err error) {
				require.NoError(t, err)
				require.Equal(t, "payload", config.Kafka.SchemaIdLocation)
			},
		},
		{
			name: "schemaIdPayloadEncoding",
			config: `{
"kafka":{
  "schemaIdPayloadEncoding": "4"
}}`,
			test: func(t *testing.T, config *asyncapi3.MessageBinding, err error) {
				require.NoError(t, err)
				require.Equal(t, "4", config.Kafka.SchemaIdPayloadEncoding)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &asyncapi3.MessageBinding{}
			err := json.Unmarshal([]byte(tc.config), &cfg)

			tc.test(t, cfg, err)
		})
	}
}
