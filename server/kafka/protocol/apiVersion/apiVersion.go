package apiVersion

import (
	"mokapi/server/kafka/protocol"
)

func init() {
	protocol.Register(
		protocol.ApiReg{
			ApiKey:     protocol.ApiVersions,
			MinVersion: 0,
			MaxVersion: 3},
		&Request{},
		&Response{},
		3,
	)
}

type Request struct {
	ClientSwName    string           `kafka:"min=3,compact=3"`
	ClientSwVersion string           `kafka:"min=3,compact=3"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type Response struct {
	ErrorCode      int16            `kafka:""`
	ApiKeys        []ApiKeyResponse `kafka:"compact=3"`
	ThrottleTimeMs int32            `kafka:"min=1"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type ApiKeyResponse struct {
	ApiKey     protocol.ApiKey  `kafka:""`
	MinVersion int16            `kafka:""`
	MaxVersion int16            `kafka:""`
	TagFields  map[int64]string `kafka:"type=TAG_BUFFER"`
}
