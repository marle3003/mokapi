package asyncApi

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"math"
	"strconv"
)

type KafkaBinding struct {
	Group Group
	Log   Log
}

type Group struct {
	Initial Initial
}

type Initial struct {
	Rebalance Rebalance
}

type Rebalance struct {
	Delay int
}

type Log struct {
	Retention        Retention
	Segment          Segment
	CleanerBackoffMs int64
}

type Segment struct {
	Bytes         int64
	DeleteDelayMs int64
}

type Retention struct {
	Bytes           int64
	Ms              int
	Minutes         int
	Hours           int
	CheckIntervalMs int64
}

func (b *KafkaBinding) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]string)
	value.Decode(m)

	// set default values
	b.Group.Initial.Rebalance.Delay = 3000
	b.Log.Retention.Hours = 24
	b.Log.Retention.Minutes = math.MinInt32
	b.Log.Retention.Ms = math.MinInt32
	b.Log.Retention.Bytes = -1
	b.Log.Retention.CheckIntervalMs = 300000 // 5min
	b.Log.Segment.Bytes = 10485760           // 10 megabytes
	b.Log.Segment.DeleteDelayMs = 60000      // 1 min
	b.Log.CleanerBackoffMs = 15000           // 15 sec

	for k, v := range m {
		switch k {
		case "group.initial.rebalance.delay.ms":
			if i, err := strconv.Atoi(v); err != nil {
				log.Errorf("unable to convert 'group.initial.rebalance.delay.ms' to int, using default instead: %v", err)
			} else {
				b.Group.Initial.Rebalance.Delay = i
			}
		case "log.retention.hours":
			if i, err := strconv.Atoi(v); err != nil {
				log.Errorf("unable to convert 'log.retention.hours' to int, using default instead: %v", err)
			} else {
				b.Log.Retention.Hours = i
			}
		case "log.retention.minutes":
			if i, err := strconv.Atoi(v); err != nil {
				log.Errorf("unable to convert 'log.retention.minutes' to int, using default instead: %v", err)
			} else {
				b.Log.Retention.Hours = math.MinInt32
				b.Log.Retention.Minutes = i
			}
		case "log.retention.ms":
			if i, err := strconv.Atoi(v); err != nil {
				log.Errorf("unable to convert 'log.retention.ms' to int, using default instead: %v", err)
			} else {
				b.Log.Retention.Hours = math.MinInt32
				b.Log.Retention.Ms = i
			}
		case "log.retention.bytes":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.retention.bytes' to long, using default instead: %v", err)
			} else {
				b.Log.Retention.Bytes = i
			}
		case "log.retention.check.interval.ms":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.retention.check.interval.ms' to long, using default instead: %v", err)
			} else {
				b.Log.Retention.CheckIntervalMs = i
			}
		case "log.segment.delete.delay.ms":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.segment.delete.delay.ms' to long, using default instead: %v", err)
			} else {
				b.Log.Segment.DeleteDelayMs = i
			}
		case "log.segment.bytes":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.segment.bytes' to long, using default instead: %v", err)
			} else {
				b.Log.Segment.Bytes = i
			}
		case "log.cleaner.backoff.ms":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.cleaner.backoff.ms' to long, using default instead: %v", err)
			} else {
				b.Log.CleanerBackoffMs = i
			}
		}
	}

	return nil
}
