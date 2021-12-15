package memory

import (
	"mokapi/server/kafka"
	"mokapi/server/kafka/protocol"
	"sync"
	"time"
)

type Partition struct {
	index         int
	segments      map[int64]*Segment
	activeSegment int64
	head          int64
	tail          int64
	lock          sync.RWMutex
	leader        *Broker
	replicas      []*Broker
}

type Segment struct {
	head        int64
	tail        int64
	log         []protocol.Record
	Size        int
	opened      time.Time
	closed      time.Time
	lastWritten time.Time
}

func newPartition(index int, replicas []*Broker) *Partition {
	p := &Partition{
		index:    index,
		head:     -1,
		tail:     -1,
		segments: make(map[int64]*Segment),
		replicas: replicas,
	}
	if len(replicas) > 0 {
		p.leader = replicas[0]
	}
	return p
}

func (p *Partition) Index() int {
	return p.index
}

func (p *Partition) Read(offset int64, maxBytes int) (protocol.RecordBatch, protocol.ErrorCode) {
	batch := protocol.NewRecordBatch()
	if p.tail >= 0 && offset > p.tail {
		return batch, protocol.OffsetOutOfRange
	}

	size := 0
	for {
		if offset > p.tail || size > maxBytes {
			return batch, protocol.None
		}
		seg := p.getSegment(offset)
		if seg == nil {
			return batch, protocol.None
		}

		for seg.Contains(offset) && maxBytes > size {
			r := seg.Record(offset)
			size += r.Size()
			batch.Records = append(batch.Records, r)
			offset++
		}
	}
}

func (p *Partition) Write(batch protocol.RecordBatch) {
	p.lock.Lock()
	defer p.lock.Unlock()

	now := time.Now()
	for _, r := range batch.Records {
		p.tail++
		switch {
		case len(p.segments) == 0:
			p.segments[p.activeSegment] = newSegment(p.tail)
		}

		seg := p.segments[p.activeSegment]
		r.Offset = p.tail
		if r.Time.IsZero() {
			r.Time = now
		}
		seg.log = append(seg.log, r)
		seg.tail = r.Offset
		seg.lastWritten = now
	}

}

func (p *Partition) Offset() int64 {
	return p.tail
}

func (p *Partition) StartOffset() int64 {
	return p.head
}

func (p *Partition) Replicas() []kafka.Broker {
	brokers := make([]kafka.Broker, 0, len(p.replicas))
	for _, r := range p.replicas {
		brokers = append(brokers, r)
	}
	return brokers
}

func (p *Partition) Leader() kafka.Broker {
	return p.leader
}

func (p *Partition) getSegment(offset int64) *Segment {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, v := range p.segments {
		if v.head <= offset && offset <= v.tail {
			return v
		}
	}

	return nil
}

func newSegment(offset int64) *Segment {
	return &Segment{
		head:   offset,
		tail:   offset,
		log:    make([]protocol.Record, 0),
		opened: time.Now(),
	}
}

func (s *Segment) Contains(offset int64) bool {
	return offset >= s.head && offset <= s.tail
}

func (s *Segment) Record(offset int64) protocol.Record {
	index := offset - s.head
	return s.log[index]
}
