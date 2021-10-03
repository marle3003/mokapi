package kafka

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"strconv"
)

type TopicBindings struct {
	Partitions int
	Config     map[string]string
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

func (b TopicBindings) SegmentMs() (int64, bool) {
	if s, ok := b.Config["segment.ms"]; ok {
		if i, err := strconv.ParseInt(s, 10, 64); err != nil {
			log.Errorf("unable to convert 'segment.ms' to long, using default instead: %v", err)
		} else {
			return i, true
		}
	}

	return -1, false
}

func (t *TopicBindings) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]string)
	err := value.Decode(m)
	if err != nil {
		return err
	}
	t.Config = m

	if s, ok := m["partitions"]; ok {
		if i, err := strconv.Atoi(s); err != nil {
			log.Errorf("unable to convert 'partitions' to int, using default instead: %v", err)
			t.Partitions = 1
		} else {
			t.Partitions = i
		}
	}

	return nil
}
