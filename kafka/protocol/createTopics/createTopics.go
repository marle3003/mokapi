package createTopics

import (
	"math"
	"mokapi/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.CreateTopics,
			MinVersion: 0,
			MaxVersion: 7},
		&Request{},
		&Response{},
		5,
		math.MaxInt16,
	)
}

type Request struct {
	Topics    []Topic          `kafka:""`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=5"`
}

type Topic struct {
	Name              string           `kafka:"compact=5"`
	NumPartitions     int32            `kafka:""`
	ReplicationFactor int16            `kafka:""`
	Assignments       []Assignment     `kafka:""`
	Configs           []Config         `kafka:""`
	TimeoutMs         int32            `kafka:""`
	ValidateOnly      bool             `kafka:"min=1"`
	TagFields         map[int64]string `kafka:"type=TAG_BUFFER,min=5"`
}

type Assignment struct {
	Index     int32            `kafka:""`
	BrokerIds []int32          `kafka:""`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=5"`
}

type Config struct {
	Name      string           `kafka:"compact=5"`
	Value     string           `kafka:"compact=5,nullable"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=5"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=2"`
	Topics         []TopicResponse  `kafka:""`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=5"`
}

type TopicResponse struct {
	Name              string             `kafka:"compact=5"`
	ErrorCode         protocol.ErrorCode `kafka:""`
	ErrorMessage      string             `kafka:"min=1,compact=5,nullable"`
	NumPartitions     int32              `kafka:"min=5"`
	ReplicationFactor int16              `kafka:"min=5"`
	Configs           []ConfigResponse   `kafka:"min=5"`
	TagFields         map[int64]string   `kafka:"type=TAG_BUFFER,min=5"`
}

type ConfigResponse struct {
	Name         string           `kafka:"compact=5"`
	Value        string           `kafka:"compact=5,nullable"`
	ReadOnly     bool             `kafka:""`
	ConfigSource int8             `kafka:""`
	IsSensitive  bool             `kafka:""`
	TagFields    map[int64]string `kafka:"type=TAG_BUFFER,min=5"`
}
