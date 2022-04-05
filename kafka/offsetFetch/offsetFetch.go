package offsetFetch

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.OffsetFetch,
			MinVersion: 0,
			MaxVersion: 7},
		&Request{},
		&Response{},
		6,
		7,
	)
}

type Request struct {
	GroupId       string           `kafka:"compact=6"`
	Topics        []RequestTopic   `kafka:"compact=6"`
	RequireStable bool             `kafka:"min=7"`
	TagFields     map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type RequestTopic struct {
	Name             string           `kafka:"compact=6"`
	PartitionIndexes []int32          `kafka:"compact=6"`
	TagFields        map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=3"`
	Topics         []ResponseTopic  `kafka:"compact=6"`
	ErrorCode      kafka.ErrorCode  `kafka:"min=2"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type ResponseTopic struct {
	Name       string           `kafka:"compact=6"`
	Partitions []Partition      `kafka:"compact=6"`
	TagFields  map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type Partition struct {
	Index                int32            `kafka:""`
	CommittedOffset      int64            `kafka:""`
	CommittedLeaderEpoch int32            `kafka:"min=5"`
	Metadata             string           `kafka:"compact=6,nullable"`
	ErrorCode            kafka.ErrorCode  `kafka:""`
	TagFields            map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}
