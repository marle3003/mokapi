package apiVersion

import (
	"mokapi/kafka"
)

func init() {
	kafka.Register(
		kafka.ApiReg{
			ApiKey:     kafka.ApiVersions,
			MinVersion: 0,
			MaxVersion: 3},
		&Request{},
		&Response{},
		3,
		// https://github.com/a0x8o/kafka/blob/master/clients/src/main/resources/common/message/ApiVersionsResponse.json
		// Tagged fields are only supported in the body but
		// not in the header
		4,
	)
}

type Request struct {
	ClientSwName    string           `kafka:"min=3,compact=3"`
	ClientSwVersion string           `kafka:"min=3,compact=3"`
	TagFields       map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type Response struct {
	ErrorCode      kafka.ErrorCode  `kafka:""`
	ApiKeys        []ApiKeyResponse `kafka:"compact=3"`
	ThrottleTimeMs int32            `kafka:"min=1"`
	TagFields      map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

type ApiKeyResponse struct {
	ApiKey     kafka.ApiKey     `kafka:""`
	MinVersion int16            `kafka:""`
	MaxVersion int16            `kafka:""`
	TagFields  map[int64]string `kafka:"type=TAG_BUFFER,min=3"`
}

func NewApiKeyResponse(k kafka.ApiKey, t kafka.ApiType) ApiKeyResponse {
	return ApiKeyResponse{
		ApiKey:     k,
		MinVersion: t.MinVersion,
		MaxVersion: t.MaxVersion,
	}
}
