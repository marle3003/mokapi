package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/kafka"
	"mokapi/kafka/produce"
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
	Topic         *Topic

	Leader   int
	Replicas []int

	validator *validator
	logger    LogRecord
	trigger   Trigger

	m sync.RWMutex
}

type Segment struct {
	Head        int64
	Tail        int64
	Log         []*kafka.Record
	Size        int
	Opened      time.Time
	Closed      time.Time
	LastWritten time.Time
}

type WriteOptions func(args *WriteArgs)

type WriteArgs struct {
	SkipValidation bool
}

func newPartition(index int, brokers Brokers, logger LogRecord, trigger Trigger, topic *Topic) *Partition {
	brokerList := make([]int, 0, len(brokers))
	for i, _ := range brokers {
		brokerList = append(brokerList, i)
	}
	p := &Partition{
		Index:    index,
		Head:     0,
		Tail:     0,
		Segments: make(map[int64]*Segment),
		logger:   logger,
		trigger:  trigger,
		Topic:    topic,
	}
	if len(brokerList) > 0 {
		p.Leader = brokerList[0]
	}
	if len(brokerList) > 1 {
		p.Replicas = brokerList[1:]
	} else {
		p.Replicas = make([]int, 0)
	}

	return p
}

func (p *Partition) Read(offset int64, maxBytes int) (kafka.RecordBatch, kafka.ErrorCode) {
	batch := kafka.NewRecordBatch()
	if offset < p.StartOffset() {
		return batch, kafka.OffsetOutOfRange
	}

	size := 0
	var baseOffset int64
	var baseTime time.Time
	for {
		if offset >= p.Tail || size > maxBytes {
			return batch, kafka.None
		}
		seg := p.GetSegment(offset)
		if seg == nil {
			return batch, kafka.None
		}

		for seg.contains(offset) {
			r := seg.record(offset)
			if baseOffset == 0 {
				baseOffset = r.Offset
				baseTime = r.Time
			}

			size += r.Size(baseOffset, baseTime)
			batch.Records = append(batch.Records, r)
			offset++

			if size > maxBytes {
				return batch, kafka.None
			}
		}
	}
}

func (p *Partition) Write(batch kafka.RecordBatch) (baseOffset int64, records []produce.RecordError, err error) {
	if p == nil {
		return 0, nil, fmt.Errorf("partition is nil")
	}

	p.m.Lock()
	defer p.m.Unlock()

	var writeFuncs []func()

	now := time.Now()
	baseOffset = p.Tail
	var baseTime time.Time
	for _, r := range batch.Records {
		var result *KafkaLog
		result, err = p.validator.Validate(r)
		if err != nil {
			records = append(records, produce.RecordError{BatchIndex: int32(r.Offset), BatchIndexErrorMessage: err.Error()})
		}
		if p.trigger(r, result.SchemaId) {
			// validate again
			result, err = p.validator.Validate(r)
		}

		if len(records) > 0 && p.Topic.Config.Bindings.Kafka.ValueSchemaValidation {
			return p.Tail, records, fmt.Errorf("validation error: %w", err)
		}

		if r.Time.IsZero() {
			r.Time = now
		}
		if baseTime.IsZero() {
			baseTime = r.Time
		}

		writeFuncs = append(writeFuncs, func() {
			r.Offset = p.Tail

			if len(p.Segments) == 0 {
				p.Segments[p.ActiveSegment] = newSegment(p.Tail)
			}

			segment, ok := p.Segments[p.ActiveSegment]
			if !ok {
				segment = p.addSegment()
			}

			segment.Log = append(segment.Log, r)
			segment.Tail++
			segment.LastWritten = now
			segment.Size += r.Size(baseOffset, baseTime)
			p.Tail++

			result.Partition = p.Index
			result.Offset = r.Offset
			p.logger(result, events.NewTraits().With("partition", strconv.Itoa(p.Index)))
		})
	}

	for _, writeFunc := range writeFuncs {
		writeFunc()
	}

	return
}

func (p *Partition) Offset() int64 {
	return p.Tail
}

func (p *Partition) StartOffset() int64 {
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
		p.removeSegment(s)
	}
}

func (p *Partition) removeClosedSegments() {
	for _, s := range p.Segments {
		if !s.Closed.IsZero() {
			if p.Head < s.Tail {
				p.Head = s.Tail
			}
			p.removeSegment(s)
		}
	}
}

func (p *Partition) removeSegment(s *Segment) {
	p.m.RLock()
	defer p.m.RUnlock()

	if p.Head < s.Tail {
		p.Head = s.Tail
	}

	s.delete()
	delete(p.Segments, s.Head)
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

func (p *Partition) addSegment() *Segment {
	p.m.RLock()
	defer p.m.RUnlock()

	now := time.Now()

	if active, ok := p.Segments[p.ActiveSegment]; ok {
		active.Closed = now
	}
	p.ActiveSegment = p.Offset()
	s := newSegment(p.Offset())
	p.Segments[p.ActiveSegment] = s
	log.Infof("kafka: added new segment to partition %v, topic %v", p.Index, p.Topic.Name)

	return s
}

func newSegment(offset int64) *Segment {
	return &Segment{
		Head:   offset,
		Tail:   offset,
		Log:    make([]*kafka.Record, 0),
		Opened: time.Now(),
	}
}

func (s *Segment) contains(offset int64) bool {
	return offset >= s.Head && offset < s.Tail
}

func (s *Segment) record(offset int64) *kafka.Record {
	index := offset - s.Head
	return s.Log[index]
}

func (s *Segment) delete() {
	for _, r := range s.Log {
		log.Debugf("delete record: %v", r.Offset)
		r.Key.Close()
		r.Value.Close()
	}
}
