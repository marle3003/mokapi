package store

import (
	"mokapi/kafka"
)

type Trigger func(record *kafka.Record)

type EventRecord struct {
	Offset  int64
	Key     string
	Value   string
	Headers map[string]string
}
