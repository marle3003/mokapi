package kafka

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/openapi/schema"
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

type Operation struct {
	GroupId *schema.Schema `yaml:"groupId" json:"groupId"`
}

type MessageBinding struct {
	Key *schema.Ref
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
	RetentionMs  int64
	SegmentBytes int64
	SegmentMs    int64
}

/*// LogRetentionBytes the maximum size of the log before deleting it
func (b BrokerBindings) LogRetentionBytes() int64 {
	if s, ok := b.Config["log.retention.bytes"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'log.retention.bytes' to long, using default instead: %v", err)
		} else {
			return i
		}
	}
	return -1
}

// GroupInitialRebalanceDelayMs The amount of time the group coordinator will wait for more consumers to join
// a new group before performing the first rebalance. A longer delay means potentially fewer rebalances, but
// increases the time until processing begins.
func (b BrokerBindings) GroupInitialRebalanceDelayMs() int {
	if s, ok := b.Config["group.initial.rebalance.delay.ms"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'group.initial.rebalance.delay.ms' to int, using default instead: %v", err)
		} else {
			return i
		}
	}
	return 3000
}

// LogRetentionMs The number of milliseconds to keep a log file before deleting it (in milliseconds).
// If set to -1, no time limit is applied.
func (b BrokerBindings) LogRetentionMs() int64 {
	if s, ok := b.Config["log.retention.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'log.retention.ms' to long, using default instead: %v", err)
		} else {
			return i
		}
	}

	if s, ok := b.Config["log.retention.minutes"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'log.retention.minutes' to int, using default instead: %v", err)
		} else {
			return int64(i) * 60 * 1000
		}
	}

	if s, ok := b.Config["log.retention.hours"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'log.retention.hours' to int, using default instead: %v", err)
		} else {
			return int64(i) * 60 * 60 * 1000
		}
	}
	return 604800000 // 7 days
}

// LogRetentionCheckIntervalMs The frequency in milliseconds that the log cleaner checks whether any log is eligible for deletion
func (b BrokerBindings) LogRetentionCheckIntervalMs() int64 {
	if s, ok := b.Config["log.retention.check.interval.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'log.retention.check.interval.ms' to long, using default instead: %v", err)
		} else {
			return i
		}
	}
	return 300000 // 5 minutes
}

// LogSegmentDeleteDelayMs The amount of time to wait before deleting a file from the filesystem
func (b BrokerBindings) LogSegmentDeleteDelayMs() int64 {
	if s, ok := b.Config["log.segment.delete.delay.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'log.segment.delete.delay.ms' to long, using default instead: %v", err)
		} else {
			return i
		}
	}
	return 60000 // 1 minutes
}

func (b BrokerBindings) LogRollMs() int64 {
	if s, ok := b.Config["log.roll.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'log.roll.ms' to long, using default instead: %v", err)
		} else {
			return i
		}
	}

	if s, ok := b.Config["log.roll.minutes"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'log.roll.minutes' to int, using default instead: %v", err)
		} else {
			return int64(i) * 60 * 1000
		}
	}

	if s, ok := b.Config["log.roll.hours"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'log.roll.hours' to int, using default instead: %v", err)
		} else {
			return int64(i) * 60 * 60 * 1000
		}
	}
	return 604800000 // 7 days
}

func (b BrokerBindings) LogSegmentBytes() int {
	if s, ok := b.Config["log.segment.bytes"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'log.segment.bytes' to int, using default instead: %v", err)
		} else {
			return i
		}
	}
	return 1073741824 // 1gb
}

func (b BrokerBindings) GroupMinSessionTimeoutMs() int {
	if s, ok := b.Config["group.min.session.timeout.ms"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'group.min.session.timeout.ms' to int, using default instead: %v", err)
		} else {
			return i
		}
	}
	return 6000
}*/

func (b *BrokerBindings) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]interface{})
	err := value.Decode(m)
	if err != nil {
		return err
	}

	b.LogRetentionBytes, err = getInt64(m, "log.retention.bytes")
	if err != nil {
		return err
	}
	b.LogRetentionMs, err = getInt64(m, "log.retention.ms")
	if err != nil {
		return err
	}
	b.LogRetentionMs, err = getMs(m, "log.retention")
	if err != nil {
		return err
	}
	b.LogRetentionCheckIntervalMs, err = getInt64(m, "log.retention.check.interval.ms")
	if err != nil {
		return err
	}
	b.LogSegmentDeleteDelayMs, err = getInt64(m, "log.segment.delete.delay.ms")
	if err != nil {
		return err
	}
	b.LogRollMs, err = getMs(m, "log.roll")
	if err != nil {
		return err
	}
	b.LogSegmentBytes, err = getInt64(m, "log.segment.bytes")
	if err != nil {
		return err
	}
	b.GroupInitialRebalanceDelayMs, err = getInt64(m, "group.initial.rebalance.delay.ms")
	if err != nil {
		return err
	}
	b.GroupMinSessionTimeoutMs, err = getInt64(m, "group.min.session.timeout.ms")
	if err != nil {
		return err
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
		return err
	}
	t.RetentionBytes, err = getInt64(m, "retention.bytes")
	if err != nil {
		return err
	}
	t.RetentionMs, err = getInt64(m, "retention.ms")
	if err != nil {
		return err
	}
	t.SegmentBytes, err = getInt64(m, "segment.bytes")
	if err != nil {
		return err
	}
	t.SegmentMs, err = getInt64(m, "segment.ms")
	if err != nil {
		return err
	}

	return nil
}

func getMs(m map[string]interface{}, baseKey string) (int64, error) {
	i, err := getInt64(m, baseKey+".ms")
	if i > 0 || err != nil {
		return i, err
	}
	i, err = getInt64(m, baseKey+".minutes")
	if i > 0 || err != nil {
		return i * 60 * 1000, err
	}
	i, err = getInt64(m, baseKey+".hours")
	if i > 0 || err != nil {
		return i * 60 * 60 * 1000, err
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
			return 0, fmt.Errorf("cannot unmarshal %T to int", i)
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
			return 0, fmt.Errorf("cannot unmarshal %T to int64", i)
		}
	}
	return 0, nil
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
