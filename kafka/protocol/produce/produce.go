package produce

import (
	"math"
	"mokapi/kafka/protocol"
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
		math.MaxInt16,
	)
}

type Request struct {
	TransactionalId string           `kafka:"min=3,compact=9,nullable"`
	Acks            int16            `kafka:""`
	TimeoutMs       int32            `kafka:""`
	Topics          []RequestTopic   `kafka:""`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type RequestTopic struct {
	Name       string             `kafka:"compact=9"`
	Partitions []RequestPartition `kafka:""`
	TagFields  map[int64]string   `kafka:"type=TAG_BUFFER,min=9"`
}

type RequestPartition struct {
	Index     int32                `kafka:""`
	Record    protocol.RecordBatch `kafka:"compact=9"`
	TagFields map[int64]string     `kafka:"type=TAG_BUFFER,min=9"`
}

type Response struct {
	Topics         []ResponseTopic `kafka:""`
	ThrottleTimeMs int32           `kafka:"min=1"`
}

type ResponseTopic struct {
	Name       string              `kafka:"compact=9"`
	Partitions []ResponsePartition `kafka:""`
	TagFields  map[int64]string    `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponsePartition struct {
	Index          int32              `kafka:""`
	ErrorCode      protocol.ErrorCode `kafka:""`
	BaseOffset     int64              `kafka:""`
	LogAppendTime  int64              `kafka:"min=2"`
	LogStartOffset int64              `kafka:"min=5"`
	RecordErrors   []RecordError      `kafka:"min=8"`
	ErrorMessage   string             `kafka:"min=8,compact=9,nullable"`
	TagFields      map[int64]string   `kafka:"type=TAG_BUFFER,min=9"`
}

type RecordError struct {
	BatchIndex             int32            `kafka:"min=8"`
	BatchIndexErrorMessage string           `kafka:"min=8,compact=9,nullable"`
	TagFields              map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}
