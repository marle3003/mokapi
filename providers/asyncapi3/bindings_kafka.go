package asyncapi3

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/providers/openapi/schema"
)

type BrokerBindings struct {
	Config map[string]string

	// LogRetentionBytes the maximum size of the log before deleting it
	LogRetentionBytes int64

	// LogRetentionMs The number of milliseconds to keep a log file before deleting it (in milliseconds).
	// If set to -1, no time limit is applied.
	LogRetentionMs int64

	// LogRetentionCheckIntervalMs The frequency in milliseconds that the log cleaner checks whether any log is eligible for deletion
	LogRetentionCheckIntervalMs int64

	// LogSegmentDeleteDelayMs The amount of time to wait before deleting a file from the filesystem
	LogSegmentDeleteDelayMs int64

	LogRollMs int64

	LogSegmentBytes int64

	// GroupInitialRebalanceDelayMs The amount of time the group coordinator will wait for more consumers to join
	// a new group before performing the first rebalance. A longer delay means potentially fewer rebalances, but
	// increases the time until processing begins.
	GroupInitialRebalanceDelayMs int64

	GroupMinSessionTimeoutMs int64
}

type KafkaOperation struct {
	GroupId  *schema.Schema `yaml:"groupId" json:"groupId"`
	ClientId *schema.Schema `yaml:"clientId" json:"clientId"`
}

type KafkaMessageBinding struct {
	Key *SchemaRef
}

type TopicBindings struct {
	Partitions int

	// RetentionBytes This configuration controls the maximum size a partition (which consists of log segments) can grow
	// to before we will discard old log segments to free up space if we are using the "delete" retention policy.
	// By default, there is no size limit only a time limit. Since this limit is enforced at the partition level, multiply
	// it by the number of partitions to compute the topic retention in bytes.
	RetentionBytes int64

	// RetentionMs This configuration controls the maximum time we will retain a log before we will discard old log
	// segments to free up space if we are using the "delete" retention policy. This represents an SLA on how soon
	// consumers must read their data. If set to -1, no time limit is applied.
	RetentionMs int64

	// SegmentBytes This configuration controls the segment file size for the log. Retention and cleaning is always
	// done a file at a time so a larger segment size means fewer files but less granular control over retention.
	SegmentBytes int64

	// SegmentMs This configuration controls the period of time after which Kafka will force the log to roll even if
	// the segment file isn’t full to ensure that retention can delete or compact old data.
	SegmentMs int64

	ValueSchemaValidation bool
}

func (b *BrokerBindings) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]interface{})
	err := value.Decode(m)
	if err != nil {
		return err
	}

	b.LogRetentionBytes, err = getInt64(m, "log.retention.bytes")
	if err != nil {
		return fmt.Errorf("invalid log.retention.bytes: %w", err)
	}
	b.LogRetentionMs, err = getMs(m, "log.retention")
	if err != nil {
		return err
	}
	b.LogRetentionCheckIntervalMs, err = getInt64(m, "log.retention.check.interval.ms")
	if err != nil {
		return fmt.Errorf("invalid log.retention.check.interval.ms: %w", err)
	}
	b.LogSegmentDeleteDelayMs, err = getInt64(m, "log.segment.delete.delay.ms")
	if err != nil {
		return fmt.Errorf("invalid log.segment.delete.delay.ms: %w", err)
	}
	b.LogRollMs, err = getMs(m, "log.roll")
	if err != nil {
		return err
	}
	b.LogSegmentBytes, err = getInt64(m, "log.segment.bytes")
	if err != nil {
		return fmt.Errorf("invalid log.segment.bytes: %w", err)
	}
	b.GroupInitialRebalanceDelayMs, err = getInt64(m, "group.initial.rebalance.delay.ms")
	if err != nil {
		return fmt.Errorf("invalid group.initial.rebalance.delay.ms: %w", err)
	}
	b.GroupMinSessionTimeoutMs, err = getInt64(m, "group.min.session.timeout.ms")
	if err != nil {
		return fmt.Errorf("invalid group.min.session.timeout.ms: %w", err)
	}

	return nil
}

func (t *TopicBindings) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]interface{})
	err := value.Decode(m)
	if err != nil {
		return err
	}

	t.Partitions, err = getInt(m, "partitions")
	if err != nil {
		return fmt.Errorf("invalid partition: %w", err)
	}
	t.RetentionBytes, err = getInt64(m, "retention.bytes")
	if err != nil {
		return fmt.Errorf("invalid retention.bytes: %w", err)
	}
	t.RetentionMs, err = getInt64(m, "retention.ms")
	if err != nil {
		return fmt.Errorf("invalid retention.ms: %w", err)
	}
	t.SegmentBytes, err = getInt64(m, "segment.bytes")
	if err != nil {
		return fmt.Errorf("invalid segment.bytes: %w", err)
	}
	t.SegmentMs, err = getInt64(m, "segment.ms")
	if err != nil {
		return fmt.Errorf("invalid segment.ms: %w", err)
	}
	t.ValueSchemaValidation, err = getBool(m, "confluent.value.schema.validation")
	if err != nil {
		return fmt.Errorf("invalid confluent.value.schema.validation: %w", err)
	}

	return nil
}

func getMs(m map[string]interface{}, baseKey string) (int64, error) {
	key := baseKey + ".ms"
	if i, err := getInt64(m, key); err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	} else if i > 0 {
		return i, nil
	}

	key = baseKey + ".minutes"
	if i, err := getInt64(m, key); err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	} else if i > 0 {
		return i * 60 * 1000, nil
	}

	key = baseKey + ".hours"
	if i, err := getInt64(m, key); err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	} else if i > 0 {
		return i * 60 * 60 * 1000, nil
	}

	return 0, nil
}

func getInt(m map[string]interface{}, keys ...string) (int, error) {
	i := getValue(m, keys...)
	if i != nil {
		switch v := i.(type) {
		case int:
			return v, nil
		default:
			return 0, fmt.Errorf("cannot unmarshal %T to int: %v", i, i)
		}
	}
	return 0, nil
}

func getInt64(m map[string]interface{}, keys ...string) (int64, error) {
	i := getValue(m, keys...)
	if i != nil {
		switch v := i.(type) {
		case int:
			return int64(v), nil
		case int64:
			return v, nil
		default:
			return 0, fmt.Errorf("cannot unmarshal %T to int64: %v", i, i)
		}
	}
	return 0, nil
}

func getBool(m map[string]interface{}, keys ...string) (bool, error) {
	i := getValue(m, keys...)
	if i != nil {
		switch v := i.(type) {
		case bool:
			return v, nil
		default:
			return true, fmt.Errorf("cannot unmarshal %T to bool: %v", i, i)
		}
	}
	return true, nil
}

func getValue(m map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if i, ok := m[key]; ok {
			return i
		}
	}
	if v, ok := m["topicConfiguration"]; ok {
		tc, ok := v.(map[string]interface{})
		if ok {
			return getValue(tc, keys...)
		}
	}
	return nil
}
