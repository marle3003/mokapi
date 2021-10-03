package kafka

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"strconv"
)

type BrokerBindings struct {
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
