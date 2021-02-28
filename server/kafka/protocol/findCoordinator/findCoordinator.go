package findCoordinator

import (
	"math"
	"mokapi/server/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.FindCoordinator,
			MinVersion: 0,
			MaxVersion: 3},
		&Request{},
		&Response{},
		3,
		math.MaxInt16,
	)
}

type Request struct {
	Key       string           `kafka:"compact=3"`
	KeyType   int8             `kafka:"min=1"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type Response struct {
	ThrottleTimeMs int32            `kafka:"min=1"`
	ErrorCode      int16            `kafka:""`
	ErrorMessage   string           `kafka:"min=1,nullable,compact=3"`
	NodeId         int32            `kafka:""`
	Host           string           `kafka:"compact=3"`
	Port           int32            `kafka:""`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}
