package kafka

import "mokapi/server/kafka/protocol"

type broker struct {
	Id     int
	groups map[string]group
	topics map[string]topic
}

type topic struct {
}

type partition struct {
	offset  int64
	batches []protocol.RecordBatch
}

type group struct {
	consumers []consumer
	leader    consumer
}

type consumer struct {
	clientId string
}
