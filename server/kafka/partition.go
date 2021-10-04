package kafka

import (
	log "github.com/sirupsen/logrus"
	"mokapi/server/kafka/protocol"
	"sync"
	"time"
)

type partition struct {
	index         int
	topic         *topic
	leader        *broker
	segments      map[int64]*segment
	activeSegment int64
	offset        int64
	startOffset   int64
	lock          sync.RWMutex

	committed map[string]int64 // groupName:offset
}

type segment struct {
	head        int64
	tail        int64
	log         []protocol.RecordBatch
	Size        int
	opened      time.Time
	closed      time.Time
	lastWritten time.Time
}

func newPartition(index int, topic *topic, leader *broker) *partition {
	return &partition{
		index:         index,
		topic:         topic,
		leader:        leader,
		activeSegment: 0,
		segments:      map[int64]*segment{0: newSegment(0)},
		startOffset:   0,
		committed:     make(map[string]int64),
	}
}

func newSegment(offset int64) *segment {
	return &segment{head: offset, opened: time.Now()}
}

func (p *partition) read(offset int64, maxBytes int32) (set protocol.RecordSet, size int32) {
	set = protocol.RecordSet{Batches: make([]protocol.RecordBatch, 0)}

	for {
		s := p.getSegment(offset)
		if s == nil {
			return
		}

		i := offset - s.head
		for _, b := range s.log[i:] {
			if newSize := size + b.Size(); newSize > 30000 {
				return
			}
			set.Batches = append(set.Batches, b)
			size += b.Size()
			if size > maxBytes {
				return
			}
		}
		offset = s.tail + 1
	}
}

func (p *partition) deleteSegment(key int64) {
	if p.activeSegment == key {
		p.addNewSegment()
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	seg := p.segments[key]
	if p.startOffset <= seg.tail {
		p.startOffset = seg.tail + 1
	}

	delete(p.segments, key)
}

func (p *partition) deleteClosedSegments() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for key, seg := range p.segments {
		if !seg.closed.IsZero() {
			delete(p.segments, key)
		}
	}
}

func (p *partition) addNewSegment() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.segments[p.activeSegment].closed = time.Now()

	p.activeSegment = p.offset
	p.segments[p.activeSegment] = newSegment(p.offset)

	log.Infof("kafka: added new segment to partition %v, topic %v", p.index, p.topic.name)
}

func (p *partition) getSegment(offset int64) *segment {
	for _, v := range p.segments {
		if v.head <= offset && offset <= v.tail {
			return v
		}
	}

	return nil
}

func (p *partition) getOffset(group string) int64 {
	if offset, ok := p.committed[group]; ok {
		return offset
	}
	return 0
}

func (p *partition) setOffset(group string, offset int64) {
	p.committed[group] = offset
	log.Infof("kafka: group %v committed offset %v, partition %v, topic %v", group, offset, p.index, p.topic.name)
}
