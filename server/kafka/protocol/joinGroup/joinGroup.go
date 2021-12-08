package joinGroup

import (
	"math"
	"mokapi/server/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.JoinGroup,
			MinVersion: 0,
			MaxVersion: 7},
		&Request{},
		&Response{},
		6,
		math.MaxInt16,
	)
}

type Request struct {
	GroupId            string           `kafka:"compact=6"`
	SessionTimeoutMs   int32            `kafka:""`
	RebalanceTimeoutMs int32            `kafka:"min=1"`
	MemberId           string           `kafka:"compact=6"`
	GroupInstanceId    string           `kafka:"min=5,compact=6,nullable"`
	ProtocolType       string           `kafka:"compact=6"`
	Protocols          []Protocol       `kafka:"compact=6"`
	TagFields          map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type Protocol struct {
	Name      string           `kafka:"compact=6"`
	MetaData  []byte           `kafka:"compact=6"`
	TagFields map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type Response struct {
	ThrottleTimeMs int32              `kafka:"min=2"`
	ErrorCode      protocol.ErrorCode `kafka:""`
	GenerationId   int32              `kafka:""`
	ProtocolName   string             `kafka:"compact=6"`
	Leader         string             `kafka:"compact=6"`
	MemberId       string             `kafka:"compact=6"`
	Members        []Member           `kafka:""`
	TagFields      map[int64]string   `kafka:"type=TAG_BUFFER,min=6"`
}

type Member struct {
	MemberId        string           `kafka:"compact=6"`
	GroupInstanceId string           `kafka:"min=5,compact=6,nullable"`
	MetaData        []byte           `kafka:"compact=6"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=6"`
}

type MemberGroupMetadata struct {
	MemberId string        `kafka:""`
	Metadata GroupMetadata `kafka:""`
}

type GroupMetadata struct {
	Version  int16    `kafka:""`
	Topics   []string `kafka:""`
	UserData []byte   `kafka:""`
}
