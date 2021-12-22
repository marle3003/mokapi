package kafkatest

import (
	"fmt"
	"mokapi/kafka/protocol"
	"mokapi/kafka/protocol/apiVersion"
	"mokapi/kafka/protocol/createTopics"
	"mokapi/kafka/protocol/fetch"
	"mokapi/kafka/protocol/findCoordinator"
	"mokapi/kafka/protocol/heartbeat"
	"mokapi/kafka/protocol/joinGroup"
	"mokapi/kafka/protocol/listgroup"
	"mokapi/kafka/protocol/metaData"
	"mokapi/kafka/protocol/offset"
	"mokapi/kafka/protocol/offsetCommit"
	"mokapi/kafka/protocol/offsetFetch"
	"mokapi/kafka/protocol/produce"
	"mokapi/kafka/protocol/syncGroup"
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

func BytesToString(bytes protocol.Bytes) string {
	b := make([]byte, bytes.Len())
	bytes.Read(b)
	return string(b)
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
	case *createTopics.Request, *createTopics.Response:
		return protocol.CreateTopics
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
	case protocol.CreateTopics:
		return &createTopics.Request{}
	default:
		panic(fmt.Sprintf("unknown type: %v", key))
	}
}
