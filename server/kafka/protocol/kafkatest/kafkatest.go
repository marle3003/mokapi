package kafkatest

import (
	"fmt"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/metaData"
	"mokapi/server/kafka/protocol/offset"
	"mokapi/server/kafka/protocol/offsetCommit"
	"mokapi/server/kafka/protocol/offsetFetch"
	"mokapi/server/kafka/protocol/produce"
	"mokapi/server/kafka/protocol/syncGroup"
)

func NewRequest(clientId string, version int, msg protocol.Message) *protocol.Request {
	return &protocol.Request{
		Header: &protocol.Header{
			ApiKey:     getApiKey(msg),
			ApiVersion: int16(version),
			ClientId:   clientId,
		},
		Message: msg,
	}
}

func getApiKey(msg protocol.Message) protocol.ApiKey {
	switch t := msg.(type) {
	case *produce.Request, *produce.Response:
		return protocol.Produce
	case *fetch.Request, *fetch.Response:
		return protocol.Fetch
	case *offset.Request, *offset.Response:
		return protocol.Offset
	case *metaData.Request, *metaData.Response:
		return protocol.Metadata
	case *offsetCommit.Request, *offsetCommit.Response:
		return protocol.OffsetCommit
	case *offsetFetch.Request, *offsetFetch.Response:
		return protocol.OffsetFetch
	case *findCoordinator.Request, *findCoordinator.Response:
		return protocol.FindCoordinator
	case *joinGroup.Request, *joinGroup.Response:
		return protocol.JoinGroup
	case *heartbeat.Request, *heartbeat.Response:
		return protocol.Heartbeat
	case *syncGroup.Request, *syncGroup.Response:
		return protocol.SyncGroup
	case *apiVersion.Request, *apiVersion.Response:
		return protocol.ApiVersions
	default:
		panic(fmt.Sprintf("unknown type: %v", t))
	}
}
