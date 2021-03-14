package kafka

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/server/kafka/protocol"
	"sync"
	"time"
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
	name       string
	partitions map[int]*partition
}

type partition struct {
	leader        broker
	segments      map[int64]*segment
	activeSegment int64
	offset        int64
	committed     int64
	lock          sync.RWMutex
	config        asyncApi.Log
}

type client struct {
	id            string
	group         *group
	lastHeartbeat time.Time
}

type segment struct {
	head        int64
	tail        int64
	log         []*protocol.RecordBatch
	Size        int64
	lastWritten time.Time
}

func (p *partition) read(offset int64, maxBytes int) (set protocol.RecordSet) {
	size := 0
	set = protocol.RecordSet{Batches: make([]protocol.RecordBatch, 0)}

	for {
		s := p.getSegment(offset)
		if s == nil {
			return
		}

		i := offset - s.head
		for _, b := range s.log[i:] {
			set.Batches = append(set.Batches, *b)
			size += int(b.Size())
			if size > maxBytes {
				return
			}
		}
		offset = s.tail + 1
	}
}

func (p *partition) append(batch *protocol.RecordBatch) {
	if p.segments[p.activeSegment].Size > p.config.Segment.Bytes {
		p.addNewSegment()
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	batch.Offset = p.offset
	p.offset++

	p.segments[p.activeSegment].append(batch)
}

func (p *partition) deleteSegment(key int64) {
	if p.activeSegment == key {
		p.addNewSegment()
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.segments, key)
}

func (p *partition) deleteAllInactiveSegments() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for k := range p.segments {
		if k != p.activeSegment {
			delete(p.segments, k)
		}
	}
}

func (p *partition) addNewSegment() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.activeSegment = p.offset
	p.segments[p.activeSegment] = newSegment(p.offset)
}

func (p *partition) getSegment(offset int64) *segment {
	for _, v := range p.segments {
		if v.head <= offset && offset <= v.tail {
			return v
		}
	}

	return nil
}

func (s *segment) append(batch *protocol.RecordBatch) {
	s.log = append(s.log, batch)
	s.Size += int64(batch.Size())
	s.tail = batch.Offset
	s.lastWritten = time.Now()
}

type groupAssignmentStrategy struct {
	assignmentStrategy string
	metadata           []byte
}

func createGuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func newSegment(offset int64) *segment {
	return &segment{head: offset}
}

func newTopic(name string, leader broker, config asyncApi.Log) *topic {
	return &topic{name: name, partitions: map[int]*partition{
		0: newPartition(leader, config)}}
}

func newPartition(leader broker, config asyncApi.Log) *partition {
	return &partition{leader: leader, config: config, activeSegment: 0, segments: map[int64]*segment{0: newSegment(0)}}
}

func (t *topic) addRecord(pi int, record *protocol.RecordBatch) error {
	if pi >= len(t.partitions) {
		return fmt.Errorf("index %q out of range", pi)
	}

	t.partitions[pi].append(record)
	//log.Infof("received new message to topic %q on partition %v. New offset is %v", t.name, pi, t.partitions[pi].offset)

	return nil
}
