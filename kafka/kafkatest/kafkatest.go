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
	"mokapi/kafka/initProducerId"
	"mokapi/kafka/joinGroup"
	"mokapi/kafka/listgroup"
	"mokapi/kafka/metaData"
	"mokapi/kafka/offset"
	"mokapi/kafka/offsetCommit"
	"mokapi/kafka/offsetFetch"
	"mokapi/kafka/produce"
	"mokapi/kafka/syncGroup"
)

func NewRequest(clientId string, version int16, msg kafka.Message) *kafka.Request {
	r := &kafka.Request{
		Header: &kafka.Header{
			ApiKey:     getApiKey(msg),
			ApiVersion: version,
			ClientId:   clientId,
		},
		Message: msg,
		Context: kafka.NewClientContext(context.Background(), "127.0.0.1:42424", "127.0.0.1:9092"),
	}
	ctx := kafka.ClientFromContext(r.Context)
	ctx.ClientId = clientId
	return r
}

func BytesToString(bytes kafka.Bytes) string {
	_, _ = bytes.Seek(0, io.SeekStart)
	b := make([]byte, bytes.Size())
	_, _ = bytes.Read(b)
	return string(b)
}

func getApiKey(msg kafka.Message) kafka.ApiKey {
	switch t := msg.(type) {
	case *produce.Request, *produce.Response:
		return kafka.Produce
	case *fetch.Request, *fetch.Response:
		return kafka.Fetch
	case *offset.Request, *offset.Response:
		return kafka.ListOffsets
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
	case *initProducerId.Request, *initProducerId.Response:
		return kafka.InitProducerId
	default:
		panic(fmt.Sprintf("unknown type: %v", t))
	}
}
