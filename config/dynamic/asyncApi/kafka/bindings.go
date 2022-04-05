package kafka

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/openapi/schema"
	"strconv"
)

type BrokerBindings struct {
	Config map[string]string
}

type Operation struct {
	GroupId *schema.Schema `yaml:"groupId" json:"groupId"`
}

type MessageBinding struct {
	Key *schema.Ref
}

type TopicBindings struct {
	Config map[string]string
}

// LogRetentionBytes the maximum size of the log before deleting it
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

func (b *BrokerBindings) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]string)
	err := value.Decode(m)
	if err != nil {
		return err
	}
	b.Config = m

	return nil
}

// RetentionBytes This configuration controls the maximum size a partition (which consists of log segments) can grow
// to before we will discard old log segments to free up space if we are using the "delete" retention policy.
// By default, there is no size limit only a time limit. Since this limit is enforced at the partition level, multiply
// it by the number of partitions to compute the topic retention in bytes.
func (t TopicBindings) RetentionBytes() (int64, bool) {
	if s, ok := t.Config["retention.bytes"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'retention.bytes' to long, using default instead: %v", err)
		} else {
			return i, true
		}
	}
	return -1, false
}

// RetentionMs This configuration controls the maximum time we will retain a log before we will discard old log
// segments to free up space if we are using the "delete" retention policy. This represents an SLA on how soon
// consumers must read their data. If set to -1, no time limit is applied.
func (t TopicBindings) RetentionMs() (int64, bool) {
	if s, ok := t.Config["retention.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'retention.ms' to long, using default instead: %v", err)
		} else {
			return i, true
		}
	}
	return -1, false
}

func (t TopicBindings) SegmentBytes() (int, bool) {
	if s, ok := t.Config["segment.bytes"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'segment.bytes' to int, using default instead: %v", err)
		} else {
			return i, true
		}
	}
	return -1, false
}

func (t TopicBindings) SegmentMs() (int64, bool) {
	if s, ok := t.Config["segment.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'segment.ms' to long, using default instead: %v", err)
		} else {
			return i, true
		}
	}

	return -1, false
}

func (t TopicBindings) Partitions() int {
	if s, ok := t.Config["partitions"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'partitions' to int, using default instead: %v", err)
			return 1
		} else {
			return i
		}
	}
	return 1
}

func (t *TopicBindings) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]string)
	err := value.Decode(m)
	if err != nil {
		return err
	}
	t.Config = m

	return nil
}
