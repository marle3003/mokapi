package store

import (
	log "github.com/sirupsen/logrus"
	"mokapi/kafka/protocol"
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
	leader        int
	replicas      []int

	validator *validator
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

func newPartition(index int, brokers map[int]*Broker) *Partition {
	replicas := make([]int, 0, len(brokers))
	for i, _ := range brokers {
		replicas = append(replicas, i)
	}
	p := &Partition{
		index:    index,
		head:     0,
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

		for seg.contains(offset) && maxBytes > size {
			r := seg.record(offset)
			size += r.Size()
			batch.Records = append(batch.Records, r)
			offset++
		}
	}
}

func (p *Partition) Write(batch protocol.RecordBatch) (baseOffset int64, err error) {
	if p.validator != nil {
		for _, r := range batch.Records {
			err := p.validator.Payload(r.Value)
			if err != nil {
				return p.tail, err
			}
		}
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	now := time.Now()
	baseOffset = p.tail + 1
	for _, r := range batch.Records {
		p.tail++
		r.Offset = p.tail
		switch {
		case len(p.segments) == 0:
			p.segments[p.activeSegment] = newSegment(p.tail)
		}

		seg := p.segments[p.activeSegment]
		if r.Time.IsZero() {
			r.Time = now
		}
		seg.log = append(seg.log, r)
		seg.tail += 1
		seg.lastWritten = now

		log.Debugf("new message written to partition=%v offset=%v", p.index, r.Offset)
	}

	return
}

func (p *Partition) Offset() int64 {
	return p.tail
}

func (p *Partition) StartOffset() int64 {
	if p.tail < 0 {
		return -1
	}
	return p.head
}

func (p *Partition) Replicas() []int {
	return p.replicas
}

func (p *Partition) Leader() int {
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

func (p *Partition) delete() {
	for _, s := range p.segments {
		s.delete()
	}
}

func (p *Partition) removeReplica(id int) {
	i := 0
	for _, replica := range p.replicas {
		if replica != id {
			p.replicas[i] = replica
			i++
		}
	}
	p.replicas = p.replicas[:i]
}

func newSegment(offset int64) *Segment {
	return &Segment{
		head:   offset,
		tail:   offset,
		log:    make([]protocol.Record, 0),
		opened: time.Now(),
	}
}

func (s *Segment) contains(offset int64) bool {
	return offset >= s.head && offset < s.tail
}

func (s *Segment) record(offset int64) protocol.Record {
	index := offset - s.head
	return s.log[index]
}

func (s *Segment) delete() {
	for _, r := range s.log {
		r.Key.Close()
		r.Value.Close()
	}
}
