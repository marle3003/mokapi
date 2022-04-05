package heartbeat

import (
	"math"
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.Heartbeat,
			MinVersion: 0,
			MaxVersion: 4},
		&Request{},
		&Response{},
		4,
		math.MaxInt16,
	)
}

type Request struct {
	GroupId         string           `kafka:"compact=4"`
	GenerationId    int32            `kafka:""`
	MemberId        string           `kafka:"compact=4"`
	GroupInstanceId string           `kafka:"min=3,compact=4,nullable"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=4"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=1"`
	ErrorCode      kafka.ErrorCode  `kafka:""`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=4"`
}
