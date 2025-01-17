package produce

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.Produce,
			MinVersion: 0,
			MaxVersion: 9},
		&Request{},
		&Response{},
		9,
		9,
	)
}

type Request struct {
	TransactionalId string           `kafka:"min=3,compact=9,nullable"`
	Acks            int16            `kafka:""`
	TimeoutMs       int32            `kafka:""`
	Topics          []RequestTopic   `kafka:"compact=9"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type RequestTopic struct {
	Name       string             `kafka:"compact=9"`
	Partitions []RequestPartition `kafka:"compact=9"`
	TagFields  map[int64]string   `kafka:"type=TAG_BUFFER,min=9"`
}

type RequestPartition struct {
	Index     int32             `kafka:""`
	Record    kafka.RecordBatch `kafka:"compact=9"`
	TagFields map[int64]string  `kafka:"type=TAG_BUFFER,min=9"`
}

type Response struct {
	Topics         []ResponseTopic  `kafka:"compact=9"`
	ThrottleTimeMs int32            `kafka:"min=1"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponseTopic struct {
	Name       string              `kafka:"compact=9"`
	Partitions []ResponsePartition `kafka:"compact=9"`
	TagFields  map[int64]string    `kafka:"type=TAG_BUFFER,min=9"`
}

type ResponsePartition struct {
	Index          int32            `kafka:""`
	ErrorCode      kafka.ErrorCode  `kafka:""`
	BaseOffset     int64            `kafka:""`
	LogAppendTime  int64            `kafka:"min=2"`
	LogStartOffset int64            `kafka:"min=5"`
	RecordErrors   []RecordError    `kafka:"min=8,compact=9"`
	ErrorMessage   string           `kafka:"min=8,compact=9,nullable"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}

type RecordError struct {
	BatchIndex             int32            `kafka:"min=8"`
	BatchIndexErrorMessage string           `kafka:"min=8,compact=9,nullable"`
	TagFields              map[int64]string `kafka:"type=TAG_BUFFER,min=9"`
}
