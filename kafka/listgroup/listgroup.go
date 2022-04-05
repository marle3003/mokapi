package listgroup

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.ListGroup,
			MinVersion: 0,
			MaxVersion: 4},
		&Request{},
		&Response{},
		3,
		3,
	)
}

type Request struct {
	StatesFilter []string         `kafka:"min=4,compact=4"`
	TagFields    map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=1"`
	ErrorCode      kafka.ErrorCode  `kafka:""`
	Groups         []Group          `kafka:""`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type Group struct {
	GroupId      string           `kafka:"compact=3"`
	ProtocolType string           `kafka:"compact=3"`
	GroupState   string           `kafka:"min=4,compact=4"`
	TagFields    map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}
