package fetch

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.Fetch,
			MinVersion: 0,
			MaxVersion: 12},
		&Request{},
		&Response{},
		12,
		12,
	)
}

type Request struct {
	ReplicaId       int32            `kafka:""`
	MaxWaitMs       int32            `kafka:""`
	MinBytes        int32            `kafka:""`
	MaxBytes        int32            `kafka:"min=3"`
	IsolationLevel  int8             `kafka:"min=4"`
	SessionId       int32            `kafka:"min=7"`
	SessionEpoch    int32            `kafka:"min=7"`
	Topics          []Topic          `kafka:"compact=12"`
	ForgottenTopics []Topic          `kafka:"min=7,compact=12"`
	RackId          string           `kafka:"min=11,compact=12"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=12"`
}

type Topic struct {
	Name       string             `kafka:"compact=12"`
	Partitions []RequestPartition `kafka:"compact=12"`
	TagFields  map[int64]string   `kafka:"type=TAG_BUFFER,min=12"`
}

type RequestPartition struct {
	Index              int32 `kafka:""`
	CurrentLeaderEpoch int32 `kafka:"min=9"`
	FetchOffset        int64 `kafka:""`
	// only used by followers
	LastFetchedEpoch int32            `kafka:"min=12"`
	LogStartOffset   int64            `kafka:"min=5"`
	MaxBytes         int32            `kafka:""`
	TagFields        map[int64]string `kafka:"type=TAG_BUFFER,min=12"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=1"`
	ErrorCode      kafka.ErrorCode  `kafka:"min=7"`
	SessionId      int32            `kafka:"min=7"`
	Topics         []ResponseTopic  `kafka:"compact=12"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=12"`
}

type ResponseTopic struct {
	Name       string              `kafka:"compact=12"`
	Partitions []ResponsePartition `kafka:"compact=12"`
	TagFields  map[int64]string    `kafka:"type=TAG_BUFFER,min=12"`
}

type ResponsePartition struct {
	Index                int32                `kafka:""`
	ErrorCode            kafka.ErrorCode      `kafka:""`
	HighWatermark        int64                `kafka:""`
	LastStableOffset     int64                `kafka:"min=4"`
	LogStartOffset       int64                `kafka:"min=5"`
	AbortedTransactions  []AbortedTransaction `kafka:"min=4,compact=12"`
	PreferredReadReplica int32                `kafka:"min=11"`
	RecordSet            kafka.RecordBatch    `kafka:"compact=12"`
	TagFields            map[int64]string     `kafka:"type=TAG_BUFFER,min=12"`
}

type AbortedTransaction struct {
	ProducerId  int64            `kafka:"min=4"`
	FirstOffset int64            `kafka:"min=4"`
	TagFields   map[int64]string `kafka:"type=TAG_BUFFER,min=12"`
}
