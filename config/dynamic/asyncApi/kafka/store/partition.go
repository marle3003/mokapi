package store

import (
	"mokapi/kafka"
	"mokapi/runtime/events"
	"strconv"
	"sync"
	"time"
)

type Partition struct {
	Index         int
	Segments      map[int64]*Segment
	ActiveSegment int64
	Head          int64
	Tail          int64

	Leader   int
	Replicas []int

	validator *validator
	logger    LogRecord

	m sync.RWMutex
}

type Segment struct {
	Head        int64
	Tail        int64
	Log         []kafka.Record
	Size        int
	Opened      time.Time
	Closed      time.Time
	LastWritten time.Time
}

func newPartition(index int, brokers Brokers, logger LogRecord) *Partition {
	replicas := make([]int, 0, len(brokers))
	for i, _ := range brokers {
		replicas = append(replicas, i)
	}
	p := &Partition{
		Index:    index,
		Head:     0,
		Tail:     -1,
		Segments: make(map[int64]*Segment),
		Replicas: replicas,
		logger:   logger,
	}
	if len(replicas) > 0 {
		p.Leader = replicas[0]
	}
	return p
}

func (p *Partition) Read(offset int64, maxBytes int) (kafka.RecordBatch, kafka.ErrorCode) {
	batch := kafka.NewRecordBatch()
	if p.Tail >= 0 && offset > p.Tail {
		return batch, kafka.OffsetOutOfRange
	}

	size := 0
	for {
		if offset > p.Tail || size > maxBytes {
			return batch, kafka.None
		}
		seg := p.GetSegment(offset)
		if seg == nil {
			return batch, kafka.None
		}

		for seg.contains(offset) && maxBytes > size {
			r := seg.record(offset)
			size += r.Size()
			batch.Records = append(batch.Records, r)
			offset++
		}

		if size >= maxBytes {
			return batch, kafka.None
		}
	}
}

func (p *Partition) Write(batch kafka.RecordBatch) (baseOffset int64, err error) {
	if p.validator != nil {
		for _, r := range batch.Records {
			err := p.validator.Payload(r.Value)
			if err != nil {
				return p.Tail, err
			}
		}
	}

	p.m.Lock()
	defer p.m.Unlock()

	now := time.Now()
	baseOffset = p.Tail + 1
	for _, r := range batch.Records {
		p.Tail++
		r.Offset = p.Tail
		switch {
		case len(p.Segments) == 0:
			p.Segments[p.ActiveSegment] = newSegment(p.Tail)
		}

		seg := p.Segments[p.ActiveSegment]
		if r.Time.IsZero() {
			r.Time = now
		}
		seg.Log = append(seg.Log, r)
		seg.Tail += 1
		seg.LastWritten = now

		p.logger(r, events.NewTraits().With("partition", strconv.Itoa(p.Index)))
	}

	return
}

func (p *Partition) Offset() int64 {
	return p.Tail
}

func (p *Partition) StartOffset() int64 {
	if p.Tail < 0 {
		return -1
	}
	return p.Head
}

func (p *Partition) GetSegment(offset int64) *Segment {
	p.m.RLock()
	defer p.m.RUnlock()

	for _, v := range p.Segments {
		if v.Head <= offset && offset <= v.Tail {
			return v
		}
	}

	return nil
}

func (p *Partition) delete() {
	for _, s := range p.Segments {
		s.delete()
	}
}

func (p *Partition) removeReplica(id int) {
	i := 0
	for _, replica := range p.Replicas {
		if replica != id {
			p.Replicas[i] = replica
			i++
		}
	}
	p.Replicas = p.Replicas[:i]
}

func newSegment(offset int64) *Segment {
	return &Segment{
		Head:   offset,
		Tail:   offset,
		Log:    make([]kafka.Record, 0),
		Opened: time.Now(),
	}
}

func (s *Segment) contains(offset int64) bool {
	return offset >= s.Head && offset < s.Tail
}

func (s *Segment) record(offset int64) kafka.Record {
	index := offset - s.Head
	return s.Log[index]
}

func (s *Segment) delete() {
	for _, r := range s.Log {
		r.Key.Close()
		r.Value.Close()
	}
}
