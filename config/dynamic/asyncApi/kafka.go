package asyncApi

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"math"
	"mokapi/config/dynamic/openapi"
	"strconv"
)

var (
	config = map[string]string{
		"group.initial.rebalance.delay.ms": "3000",
		"log.retention.hours":              "24",
		"log.retention.minutes":            "null",
		"log.retention.ms":                 "null",
		"log.retention.bytes":              "-1",
		"log.retention.check.interval.ms":  "300000",   // 5min
		"log.segment.bytes":                "10485760", // 10 megabytes
		"log.segment.delete.delay.ms":      "60000",    // 1 min
		"log.cleaner.backoff.ms":           "15000",    // 15 sec
	}
)

type KafkaChannelBinding struct {
	Partitions int
}

type KafkaMessageBinding struct {
	Key *openapi.SchemaRef
}

type KafkaOperationBinding struct {
	GroupId  *openapi.SchemaRef
	ClientId *openapi.SchemaRef
}

type Kafka struct {
	Group Group
	Log   Log
	raw   map[string]string
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

func (k *Kafka) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]string)
	for key, val := range config {
		m[key] = val
	}
	k.raw = m
	err := value.Decode(m)
	if err != nil {
		return err
	}

	for key, v := range m {
		switch key {
		case "group.initial.rebalance.delay.ms":
			if i, err := strconv.Atoi(v); err != nil {
				log.Errorf("unable to convert 'group.initial.rebalance.delay.ms' to int, using default instead: %v", err)
			} else {
				k.Group.Initial.Rebalance.Delay = i
			}
		case "log.retention.hours":
			if i, err := strconv.Atoi(v); err != nil {
				log.Errorf("unable to convert 'log.retention.hours' to int, using default instead: %v", err)
			} else {
				k.Log.Retention.Hours = i
			}
		case "log.retention.minutes":
			if v == "null" {
				k.Log.Retention.Minutes = math.MinInt32
			} else {
				if i, err := strconv.Atoi(v); err != nil {
					log.Errorf("unable to convert 'log.retention.minutes' to int, using default instead: %v", err)
				} else {
					k.Log.Retention.Hours = math.MinInt32
					k.Log.Retention.Minutes = i
					k.raw["log.retention.hours"] = "null"
				}
			}
		case "log.retention.ms":
			if v == "null" {
				k.Log.Retention.Ms = math.MinInt32
			} else {
				if i, err := strconv.Atoi(v); err != nil {
					log.Errorf("unable to convert 'log.retention.ms' to int, using default instead: %v", err)
				} else {
					k.Log.Retention.Hours = math.MinInt32
					k.Log.Retention.Ms = i
					k.raw["log.retention.hours"] = "null"
				}
			}
		case "log.retention.bytes":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.retention.bytes' to long, using default instead: %v", err)
			} else {
				k.Log.Retention.Bytes = i
			}
		case "log.retention.check.interval.ms":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.retention.check.interval.ms' to long, using default instead: %v", err)
			} else {
				k.Log.Retention.CheckIntervalMs = i
			}
		case "log.segment.delete.delay.ms":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.segment.delete.delay.ms' to long, using default instead: %v", err)
			} else {
				k.Log.Segment.DeleteDelayMs = i
			}
		case "log.segment.bytes":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.segment.bytes' to long, using default instead: %v", err)
			} else {
				k.Log.Segment.Bytes = i
			}
		case "log.cleaner.backoff.ms":
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				log.Errorf("unable to convert 'log.cleaner.backoff.ms' to long, using default instead: %v", err)
			} else {
				k.Log.CleanerBackoffMs = i
			}
		}
	}

	return nil
}

func (k Kafka) ToMap() map[string]string {
	return k.raw
}
