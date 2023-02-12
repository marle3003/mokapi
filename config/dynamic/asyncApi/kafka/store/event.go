package store

import (
	"mokapi/kafka"
)

type Trigger func(record *kafka.Record)
