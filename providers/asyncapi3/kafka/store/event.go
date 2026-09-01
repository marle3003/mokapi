package store

import (
	"mokapi/engine/common"
	"mokapi/kafka"
)

type Trigger func(topic string, partition int, record *kafka.Record, schemaId int) []*common.Action
