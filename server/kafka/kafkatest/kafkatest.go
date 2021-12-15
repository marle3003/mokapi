package kafkatest

import (
	"fmt"
	"mokapi/server/kafka/protocol"
	"mokapi/server/kafka/protocol/apiVersion"
	"mokapi/server/kafka/protocol/fetch"
	"mokapi/server/kafka/protocol/findCoordinator"
	"mokapi/server/kafka/protocol/heartbeat"
	"mokapi/server/kafka/protocol/joinGroup"
	"mokapi/server/kafka/protocol/listgroup"
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
	case *listgroup.Request, *listgroup.Response:
		return protocol.ListGroup
	default:
		panic(fmt.Sprintf("unknown type: %v", t))
	}
}

func GetRequest(key protocol.ApiKey) protocol.Message {
	switch key {
	case protocol.Produce:
		return &produce.Request{}
	case protocol.Fetch:
		return &fetch.Request{}
	case protocol.Offset:
		return &offset.Request{}
	case protocol.Metadata:
		return &metaData.Request{}
	case protocol.OffsetCommit:
		return &offsetCommit.Request{}
	case protocol.OffsetFetch:
		return &offsetFetch.Request{}
	case protocol.FindCoordinator:
		return &findCoordinator.Request{}
	case protocol.JoinGroup:
		return &joinGroup.Request{}
	case protocol.Heartbeat:
		return &heartbeat.Request{}
	case protocol.SyncGroup:
		return &syncGroup.Request{}
	case protocol.ApiVersions:
		return &apiVersion.Request{}
	case protocol.ListGroup:
		return &listgroup.Request{}
	default:
		panic(fmt.Sprintf("unknown type: %v", key))
	}
}
