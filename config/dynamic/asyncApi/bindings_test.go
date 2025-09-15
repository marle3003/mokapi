package asyncApi_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/asyncApi"
	"testing"
)

func TestBrokerBindings_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected *asyncApi.BrokerBindings
	}{
		{
			name:     "log.retention.bytes",
			input:    `{"log.retention.bytes":10}`,
			expected: &asyncApi.BrokerBindings{LogRetentionBytes: 10},
		},
		{
			name:     "log.retention.bytes",
			input:    `{"log.retention.ms":10}`,
			expected: &asyncApi.BrokerBindings{LogRetentionMs: 10},
		},
		{
			name:     "log.retention.minutes",
			input:    `{"log.retention.minutes":10}`,
			expected: &asyncApi.BrokerBindings{LogRetentionMs: 600000},
		},
		{
			name:     "log.retention.hours",
			input:    `{"log.retention.hours":1}`,
			expected: &asyncApi.BrokerBindings{LogRetentionMs: 3600000},
		},
		{
			name:     "log.retention.check.interval.ms",
			input:    `{"log.retention.check.interval.ms":10}`,
			expected: &asyncApi.BrokerBindings{LogRetentionCheckIntervalMs: 10},
		},
		{
			name:     "log.segment.delete.delay.ms",
			input:    `{"log.segment.delete.delay.ms":10}`,
			expected: &asyncApi.BrokerBindings{LogSegmentDeleteDelayMs: 10},
		},
		{
			name:     "log.roll.ms",
			input:    `{"log.roll.ms":10}`,
			expected: &asyncApi.BrokerBindings{LogRollMs: 10},
		},
		{
			name:     "log.roll.minutes",
			input:    `{"log.roll.minutes":10}`,
			expected: &asyncApi.BrokerBindings{LogRollMs: 600000},
		},
		{
			name:     "log.roll.hours",
			input:    `{"log.roll.hours":10}`,
			expected: &asyncApi.BrokerBindings{LogRollMs: 36000000},
		},
		{
			name:     "log.segment.bytes",
			input:    `{"log.segment.bytes":10}`,
			expected: &asyncApi.BrokerBindings{LogSegmentBytes: 10},
		},
		{
			name:     "group.initial.rebalance.delay.ms",
			input:    `{"group.initial.rebalance.delay.ms":10}`,
			expected: &asyncApi.BrokerBindings{GroupInitialRebalanceDelayMs: 10},
		},
		{
			name:     "group.min.session.timeout.ms",
			input:    `{"group.min.session.timeout.ms":10}`,
			expected: &asyncApi.BrokerBindings{GroupMinSessionTimeoutMs: 10},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := &asyncApi.BrokerBindings{}
			err := json.Unmarshal([]byte(tc.input), target)
			require.NoError(t, err)
			require.Equal(t, tc.expected, target)
		})
	}
}

func TestBrokerBindings_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected *asyncApi.BrokerBindings
	}{
		{
			name:     "log.retention.bytes",
			input:    `log.retention.bytes: 10`,
			expected: &asyncApi.BrokerBindings{LogRetentionBytes: 10},
		},
		{
			name:     "log.retention.bytes",
			input:    `log.retention.ms: 10`,
			expected: &asyncApi.BrokerBindings{LogRetentionMs: 10},
		},
		{
			name:     "log.retention.minutes",
			input:    `log.retention.minutes: 10`,
			expected: &asyncApi.BrokerBindings{LogRetentionMs: 600000},
		},
		{
			name:     "log.retention.hours",
			input:    `log.retention.hours: 1`,
			expected: &asyncApi.BrokerBindings{LogRetentionMs: 3600000},
		},
		{
			name:     "log.retention.check.interval.ms",
			input:    `log.retention.check.interval.ms: 10`,
			expected: &asyncApi.BrokerBindings{LogRetentionCheckIntervalMs: 10},
		},
		{
			name:     "log.segment.delete.delay.ms",
			input:    `log.segment.delete.delay.ms: 10`,
			expected: &asyncApi.BrokerBindings{LogSegmentDeleteDelayMs: 10},
		},
		{
			name:     "log.roll.ms",
			input:    `log.roll.ms: 10`,
			expected: &asyncApi.BrokerBindings{LogRollMs: 10},
		},
		{
			name:     "log.roll.minutes",
			input:    `log.roll.minutes: 10`,
			expected: &asyncApi.BrokerBindings{LogRollMs: 600000},
		},
		{
			name:     "log.roll.hours",
			input:    `log.roll.hours: 10`,
			expected: &asyncApi.BrokerBindings{LogRollMs: 36000000},
		},
		{
			name:     "log.segment.bytes",
			input:    `log.segment.bytes: 10`,
			expected: &asyncApi.BrokerBindings{LogSegmentBytes: 10},
		},
		{
			name:     "group.initial.rebalance.delay.ms",
			input:    `group.initial.rebalance.delay.ms: 10`,
			expected: &asyncApi.BrokerBindings{GroupInitialRebalanceDelayMs: 10},
		},
		{
			name:     "group.min.session.timeout.ms",
			input:    `group.min.session.timeout.ms: 10`,
			expected: &asyncApi.BrokerBindings{GroupMinSessionTimeoutMs: 10},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := &asyncApi.BrokerBindings{}
			err := yaml.Unmarshal([]byte(tc.input), target)
			require.NoError(t, err)
			require.Equal(t, tc.expected, target)
		})
	}
}

func TestTopicBindings_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected *asyncApi.TopicBindings
	}{
		{
			name:     "default",
			input:    `{}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true},
		},
		{
			name:     "partitions",
			input:    `{"partitions":10}`,
			expected: &asyncApi.TopicBindings{Partitions: 10, ValueSchemaValidation: true, KeySchemaValidation: true},
		},
		{
			name:     "retention.bytes",
			input:    `{"retention.bytes":10}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, RetentionBytes: 10},
		},
		{
			name:     "retention.ms",
			input:    `{"retention.ms":10}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, RetentionMs: 10},
		},
		{
			name:     "segment.bytes",
			input:    `{"segment.bytes":10}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, SegmentBytes: 10},
		},
		{
			name:     "segment.ms",
			input:    `{"segment.ms":10}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, SegmentMs: 10},
		},
		{
			name:     "confluent.value.schema.validation",
			input:    `{"confluent.value.schema.validation":false}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: false, KeySchemaValidation: true},
		},
		{
			name:     "confluent.value.schema.validation",
			input:    `{"confluent.key.schema.validation":false}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: false},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := &asyncApi.TopicBindings{}
			err := json.Unmarshal([]byte(tc.input), target)
			require.NoError(t, err)
			require.Equal(t, tc.expected, target)
		})
	}
}

func TestTopicBindings_UnmarshalYaml(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected *asyncApi.TopicBindings
	}{
		{
			name:     "default",
			input:    `{}`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true},
		},
		{
			name:     "partitions",
			input:    `partitions: 10`,
			expected: &asyncApi.TopicBindings{Partitions: 10, ValueSchemaValidation: true, KeySchemaValidation: true},
		},
		{
			name:     "retention.bytes",
			input:    `retention.bytes: 10`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, RetentionBytes: 10},
		},
		{
			name:     "retention.ms",
			input:    `retention.ms: 10`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, RetentionMs: 10},
		},
		{
			name:     "segment.bytes",
			input:    `segment.bytes: 10`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, SegmentBytes: 10},
		},
		{
			name:     "segment.ms",
			input:    `segment.ms: 10`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: true, SegmentMs: 10},
		},
		{
			name:     "confluent.value.schema.validation",
			input:    `confluent.value.schema.validation: false`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: false, KeySchemaValidation: true},
		},
		{
			name:     "confluent.value.schema.validation",
			input:    `confluent.key.schema.validation: false`,
			expected: &asyncApi.TopicBindings{Partitions: 1, ValueSchemaValidation: true, KeySchemaValidation: false},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := &asyncApi.TopicBindings{}
			err := yaml.Unmarshal([]byte(tc.input), target)
			require.NoError(t, err)
			require.Equal(t, tc.expected, target)
		})
	}
}
