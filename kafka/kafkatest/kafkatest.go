package kafkatest

import (
	"context"
	"fmt"
	"io"
	"mokapi/kafka"
	"mokapi/kafka/apiVersion"
	"mokapi/kafka/createTopics"
	"mokapi/kafka/fetch"
	"mokapi/kafka/findCoordinator"
	"mokapi/kafka/heartbeat"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/listgroup"
	"mokapi/kafka/metaData"
	"mokapi/kafka/offset"
	"mokapi/kafka/offsetCommit"
	"mokapi/kafka/offsetFetch"
	"mokapi/kafka/produce"
	"mokapi/kafka/syncGroup"
)

func NewRequest(clientId string, version int, msg kafka.Message) *kafka.Request {
	return &kafka.Request{
		Header: &kafka.Header{
			ApiKey:     getApiKey(msg),
			ApiVersion: int16(version),
			ClientId:   clientId,
		},
		Message: msg,
		Context: kafka.NewClientContext(context.Background(), "127.0.0.1:42424"),
	}
}

func BytesToString(bytes kafka.Bytes) string {
	bytes.Seek(0, io.SeekStart)
	b := make([]byte, bytes.Len())
	bytes.Read(b)
	return string(b)
}

func getApiKey(msg kafka.Message) kafka.ApiKey {
	switch t := msg.(type) {
	case *produce.Request, *produce.Response:
		return kafka.Produce
	case *fetch.Request, *fetch.Response:
		return kafka.Fetch
	case *offset.Request, *offset.Response:
		return kafka.Offset
	case *metaData.Request, *metaData.Response:
		return kafka.Metadata
	case *offsetCommit.Request, *offsetCommit.Response:
		return kafka.OffsetCommit
	case *offsetFetch.Request, *offsetFetch.Response:
		return kafka.OffsetFetch
	case *findCoordinator.Request, *findCoordinator.Response:
		return kafka.FindCoordinator
	case *joinGroup.Request, *joinGroup.Response:
		return kafka.JoinGroup
	case *heartbeat.Request, *heartbeat.Response:
		return kafka.Heartbeat
	case *syncGroup.Request, *syncGroup.Response:
		return kafka.SyncGroup
	case *apiVersion.Request, *apiVersion.Response:
		return kafka.ApiVersions
	case *listgroup.Request, *listgroup.Response:
		return kafka.ListGroup
	case *createTopics.Request, *createTopics.Response:
		return kafka.CreateTopics
	default:
		panic(fmt.Sprintf("unknown type: %v", t))
	}
}

func GetRequest(key kafka.ApiKey) kafka.Message {
	switch key {
	case kafka.Produce:
		return &produce.Request{}
	case kafka.Fetch:
		return &fetch.Request{}
	case kafka.Offset:
		return &offset.Request{}
	case kafka.Metadata:
		return &metaData.Request{}
	case kafka.OffsetCommit:
		return &offsetCommit.Request{}
	case kafka.OffsetFetch:
		return &offsetFetch.Request{}
	case kafka.FindCoordinator:
		return &findCoordinator.Request{}
	case kafka.JoinGroup:
		return &joinGroup.Request{}
	case kafka.Heartbeat:
		return &heartbeat.Request{}
	case kafka.SyncGroup:
		return &syncGroup.Request{}
	case kafka.ApiVersions:
		return &apiVersion.Request{}
	case kafka.ListGroup:
		return &listgroup.Request{}
	case kafka.CreateTopics:
		return &createTopics.Request{}
	default:
		panic(fmt.Sprintf("unknown type: %v", key))
	}
}
