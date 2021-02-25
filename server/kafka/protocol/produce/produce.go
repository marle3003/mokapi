package produce

import (
	"math"
	"mokapi/server/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.Produce,
			MinVersion: 0,
			MaxVersion: 8},
		&Request{},
		&Response{},
		math.MaxInt16,
	)
}

type Request struct {
	TransactionalId string         `kafka:"min=3,nullable"`
	Acks            int16          `kafka:""`
	Timeout         int32          `kafka:""`
	Topics          []RequestTopic `kafka:""`
}

type RequestTopic struct {
	Name string           `kafka:""`
	Data RequestPartition `kafka:""`
}

type RequestPartition struct {
	Partition int32 `kafka:""`
	Record    protocol.RecordSet
}

type Response struct {
	Topics         []ResponseTopic
	ThrottleTimeMs int32 `kafka:"min=1"`
}

type ResponseTopic struct {
	Name string `kafka:""`
}

type ResponsePartition struct {
	Index          int32         `kafka:""`
	ErrorCode      int16         `kafka:""`
	Offset         int64         `kafka:""`
	LogAppendTime  int64         `kafka:"min=2"`
	LogStartOffset int64         `kafka:"min=5"`
	RecordErrors   []RecordError `kafka:"min=8"`
	ErrorMessage   string        `kafka:"min=8,nullable"`
}

type RecordError struct {
	BatchIndex             int32  `kafka:"min=8"`
	BatchIndexErrorMessage string `kafka:"min=8,nullable"`
}
