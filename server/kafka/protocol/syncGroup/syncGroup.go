package syncGroup

import (
	"math"
	"mokapi/server/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.SyncGroup,
			MinVersion: 0,
			MaxVersion: 5},
		&Request{},
		&Response{},
		4,
		math.MaxInt16,
	)
}

type Request struct {
	GroupId          string            `kafka:"compact=4"`
	GenerationId     int32             `kafka:""`
	MemberId         string            `kafka:"compact=4"`
	GroupInstanceId  string            `kafka:"min=3,compact=4,nullable"`
	ProtocolType     string            `kafka:"min=5,compact=5,nullable"`
	ProtocolName     string            `kafka:"min=5,compact=5,nullable"`
	GroupAssignments []GroupAssignment `kafka:""`
	TagFields        map[int64]string  `kafka:"type=TAG_BUFFER,min=4"`
}

type GroupAssignment struct {
	MemberId   string           `kafka:"compact=4"`
	Assignment []byte           `kafka:"compact=4"`
	TagFields  map[int64]string `kafka:"type=TAG_BUFFER,min=4"`
}

type Response struct {
	ThrottleTimeMs int32              `kafka:"min=1"`
	ErrorCode      protocol.ErrorCode `kafka:""`
	ProtocolType   string             `kafka:"min=5,compact=5,nullable"`
	ProtocolName   string             `kafka:"min=5,compact=5,nullable"`
	Assignment     []byte             `kafka:"compact=4"`
	TagFields      map[int64]string   `kafka:"type=TAG_BUFFER,min=4"`
}
