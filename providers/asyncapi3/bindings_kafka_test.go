package asyncapi3_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/providers/asyncapi3"
	"testing"
)

func TestKafkaBindingsServer_Yaml(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, config *asyncapi3.Config, err error)
	}{
		{
			name: "log.retention.bytes",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.bytes: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.LogRetentionBytes)
			},
		},
		{
			name: "log.retention.bytes error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.bytes: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.retention.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.ms",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.LogRetentionMs)
			},
		},
		{
			name: "log.retention.ms error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.retention.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.minutes",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.minutes: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(600000), config.Servers["test"].Value.Bindings.Kafka.LogRetentionMs)
			},
		},
		{
			name: "log.retention.minutes error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.minutes: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.retention.minutes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.hours",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.hours: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(36000000), config.Servers["test"].Value.Bindings.Kafka.LogRetentionMs)
			},
		},
		{
			name: "log.retention.hours error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.hours: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.retention.hours: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.retention.check.interval.ms",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.check.interval.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.LogRetentionCheckIntervalMs)
			},
		},
		{
			name: "log.retention.check.interval.ms error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.retention.check.interval.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.retention.check.interval.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.segment.delete.delay.ms",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.segment.delete.delay.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.LogSegmentDeleteDelayMs)
			},
		},
		{
			name: "log.segment.delete.delay.ms error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.segment.delete.delay.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.segment.delete.delay.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.roll.ms",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.roll.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.LogRollMs)
			},
		},
		{
			name: "log.roll.ms error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.roll.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.roll.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.roll.minutes",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.roll.minutes: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(600000), config.Servers["test"].Value.Bindings.Kafka.LogRollMs)
			},
		},
		{
			name: "log.roll.minutes error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.roll.minutes: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.roll.minutes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.roll.hours",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.roll.hours: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(36000000), config.Servers["test"].Value.Bindings.Kafka.LogRollMs)
			},
		},
		{
			name: "log.roll.hours error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.roll.hours: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.roll.hours: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "log.segment.bytes",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.segment.bytes: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.LogSegmentBytes)
			},
		},
		{
			name: "log.segment.bytes error",
			config: `
servers:
  test:
    bindings:
      kafka:
        log.segment.bytes: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid log.segment.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "group.initial.rebalance.delay.ms",
			config: `
servers:
  test:
    bindings:
      kafka:
        group.initial.rebalance.delay.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.GroupInitialRebalanceDelayMs)
			},
		},
		{
			name: "group.initial.rebalance.delay.ms error",
			config: `
servers:
  test:
    bindings:
      kafka:
        group.initial.rebalance.delay.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid group.initial.rebalance.delay.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "group.min.session.timeout.ms",
			config: `
servers:
  test:
    bindings:
      kafka:
        group.min.session.timeout.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Servers["test"].Value.Bindings.Kafka.GroupMinSessionTimeoutMs)
			},
		},
		{
			name: "group.min.session.timeout.ms error",
			config: `
servers:
  test:
    bindings:
      kafka:
        group.min.session.timeout.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid group.min.session.timeout.ms: cannot unmarshal string to int64: foo")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &asyncapi3.Config{}
			err := yaml.Unmarshal([]byte(tc.config), &cfg)

			tc.test(t, cfg, err)
		})
	}
}

func TestKafkaBindingsTopic_Yaml(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, config *asyncapi3.Config, err error)
	}{
		{
			name: "partitions",
			config: `
channels:
  test:
    bindings:
      kafka:
        partitions: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, 10, config.Channels["test"].Value.Bindings.Kafka.Partitions)
			},
		},
		{
			name: "partitions error",
			config: `
channels:
  test:
    bindings:
      kafka:
        partitions: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid partition: cannot unmarshal string to int: foo")
			},
		},
		{
			name: "retention.bytes",
			config: `
channels:
  test:
    bindings:
      kafka:
        retention.bytes: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Channels["test"].Value.Bindings.Kafka.RetentionBytes)
			},
		},
		{
			name: "retention.bytes error",
			config: `
channels:
  test:
    bindings:
      kafka:
        retention.bytes: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid retention.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "retention.ms",
			config: `
channels:
  test:
    bindings:
      kafka:
        retention.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Channels["test"].Value.Bindings.Kafka.RetentionMs)
			},
		},
		{
			name: "retention.ms error",
			config: `
channels:
  test:
    bindings:
      kafka:
        retention.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid retention.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "segment.bytes",
			config: `
channels:
  test:
    bindings:
      kafka:
        segment.bytes: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Channels["test"].Value.Bindings.Kafka.SegmentBytes)
			},
		},
		{
			name: "segment.bytes error",
			config: `
channels:
  test:
    bindings:
      kafka:
        segment.bytes: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid segment.bytes: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "segment.ms",
			config: `
channels:
  test:
    bindings:
      kafka:
        segment.ms: 10
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), config.Channels["test"].Value.Bindings.Kafka.SegmentMs)
			},
		},
		{
			name: "segment.ms error",
			config: `
channels:
  test:
    bindings:
      kafka:
        segment.ms: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid segment.ms: cannot unmarshal string to int64: foo")
			},
		},
		{
			name: "confluent.value.schema.validation",
			config: `
channels:
  test:
    bindings:
      kafka:
        confluent.value.schema.validation: false
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.False(t, config.Channels["test"].Value.Bindings.Kafka.ValueSchemaValidation)
			},
		},
		{
			name: "confluent.value.schema.validation error",
			config: `
channels:
  test:
    bindings:
      kafka:
        confluent.value.schema.validation: foo
`,
			test: func(t *testing.T, config *asyncapi3.Config, err error) {
				require.EqualError(t, err, "invalid confluent.value.schema.validation: cannot unmarshal string to bool: foo")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &asyncapi3.Config{}
			err := yaml.Unmarshal([]byte(tc.config), &cfg)

			tc.test(t, cfg, err)
		})
	}
}
