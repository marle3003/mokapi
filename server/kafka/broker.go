package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/server/kafka/protocol"
)

var (
	logs = make(map[string]batchLog)
)

type broker struct {
	id   int
	host string
	port int
}

func newBroker(id int, host string, port int) broker {
	return broker{id, host, port}
}

type topic struct {
	partitions map[int]*partition
}

type partition struct {
	leader broker
	log    *batchLog
}

type consumer struct {
	id string
}

type batchLog struct {
	offset    int64
	committed int64
	batches   []*protocol.RecordBatch
}

func (l *batchLog) Append(batch *protocol.RecordBatch) {
	l.batches = append(l.batches, batch)
	batch.Offset = l.offset
	l.offset++
}

type groupAssignmentStrategy struct {
	assignmentStrategy string
	metadata           []byte
}

var (
	RebalanceInProgressCode = 27
)

func createGuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
