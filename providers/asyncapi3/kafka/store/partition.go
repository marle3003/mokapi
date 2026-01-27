package store

import (
	"fmt"
	"mokapi/kafka"
	"mokapi/kafka/produce"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"slices"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Partition struct {
	Index         int
	Segments      map[int64]*Segment
	ActiveSegment int64
	Head          int64
	Tail          int64
	Topic         *Topic

	// only for log cleaner
	leader *Broker

	validator *validator
	logger    LogRecord
	trigger   Trigger

	producers map[int64]*PartitionProducerState

	m sync.RWMutex
}

type Segment struct {
	Head        int64
	Tail        int64
	Log         []*record
	Size        int
	Opened      time.Time
	Closed      time.Time
	LastWritten time.Time
}

type record struct {
	Data *kafka.Record
	Log  *KafkaMessageLog
}

type WriteOptions struct {
	SkipValidation bool
	ClientId       string
	ScriptFile     string
}

type WriteResult struct {
	BaseOffset   int64
	Records      []produce.RecordError
	ErrorCode    kafka.ErrorCode
	ErrorMessage string
}

type PartitionProducerState struct {
	ProducerId   int64
	Epoch        int16
	LastSequence int32
}

func newPartition(index int, brokers Brokers, logger LogRecord, trigger Trigger, topic *Topic) *Partition {
	brokerList := make([]*Broker, 0, len(brokers))
	for _, b := range brokers {
		if topic.Config != nil && len(topic.Config.Servers) > 0 {
			if slices.ContainsFunc(topic.Config.Servers, func(s *asyncapi3.ServerRef) bool {
				return s.Value == b.config
			}) {
				brokerList = append(brokerList, b)
			}
		} else {
			brokerList = append(brokerList, b)
		}
	}
	p := &Partition{
		Index:     index,
		Head:      0,
		Tail:      0,
		Segments:  make(map[int64]*Segment),
		logger:    logger,
		trigger:   trigger,
		Topic:     topic,
		producers: make(map[int64]*PartitionProducerState),
	}
	if len(brokerList) > 0 {
		p.leader = brokerList[0]
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

func (p *Partition) WriteSkipValidation(batch kafka.RecordBatch) (WriteResult, error) {
	return p.write(batch, WriteOptions{SkipValidation: true})
}

func (p *Partition) Write(batch kafka.RecordBatch) (WriteResult, error) {
	return p.write(batch, WriteOptions{SkipValidation: false})
}

func (p *Partition) WriteWithOptions(batch kafka.RecordBatch, opts WriteOptions) (WriteResult, error) {
	return p.write(batch, opts)
}

func (p *Partition) write(batch kafka.RecordBatch, opts WriteOptions) (WriteResult, error) {
	if p == nil {
		return WriteResult{}, fmt.Errorf("partition is nil")
	}

	p.m.Lock()
	defer p.m.Unlock()

	result := WriteResult{}

	var writeFuncs []func()
	var producer *ProducerState
	sequenceNumber := int32(-1)

	now := time.Now()
	result.BaseOffset = p.Tail
	var baseTime time.Time
	for i, r := range batch.Records {
		// validate producer idempotence
		if r.ProducerId > 0 {
			if producer == nil {
				producer = p.Topic.s.producers[r.ProducerId]
			}
			state, ok := p.producers[r.ProducerId]
			if ok {
				sequenceNumber = state.LastSequence
			}

			if producer == nil {
				result.fail(i, kafka.InvalidProducerIdMapping, "unknown producer id")
				return result, nil
			} else if producer.ProducerEpoch != r.ProducerEpoch {
				// this is without transactional produce not possible
				// due to producer will always get a new producer id
				// should return PRODUCER_FENCED when record epoch is lower as current known
				// this should be adjusted when transactional is implemented
				result.fail(i, kafka.InvalidProducerEpoch, "producer epoch does not match")
				return result, nil
			} else if r.SequenceNumber != sequenceNumber+1 {
				var msg string
				if r.SequenceNumber <= sequenceNumber {
					msg = fmt.Sprintf("message sequence number already received: %d", r.SequenceNumber)
					result.fail(i, kafka.DuplicateSequenceNumber, msg)

				} else {
					msg = fmt.Sprintf("expected sequence number %d but got %d", sequenceNumber+1, r.SequenceNumber)
					result.fail(i, kafka.OutOfOrderSequenceNumber, msg)
				}
				return result, nil
			}
			sequenceNumber++
		}

		kLog, err := p.validator.Validate(r)
		if err != nil && !opts.SkipValidation {
			result.fail(i, kafka.InvalidRecord, err.Error())
			return result, nil
		}
		if p.trigger(r, kLog.SchemaId) && !opts.SkipValidation {
			// validate again
			kLog, err = p.validator.Validate(r)
			if err != nil {
				result.fail(i, kafka.InvalidRecord, err.Error())
				return result, nil
			}
		}

		if r.Time.IsZero() {
			r.Time = now
		}
		if baseTime.IsZero() {
			baseTime = r.Time
		}

		kLog.ClientId = opts.ClientId
		kLog.ScriptFile = opts.ScriptFile

		writeFuncs = append(writeFuncs, func() {
			r.Offset = p.Tail

			if len(p.Segments) == 0 {
				p.Segments[p.ActiveSegment] = newSegment(p.Tail)
			}

			segment, ok := p.Segments[p.ActiveSegment]
			if !ok {
				segment = p.addSegment()
			}

			segment.Log = append(segment.Log, &record{Data: r, Log: kLog})
			segment.Tail++
			segment.LastWritten = now
			segment.Size += r.Size(result.BaseOffset, baseTime)
			p.Tail++

			kLog.Partition = p.Index
			kLog.Offset = r.Offset
			p.logger(kLog, events.NewTraits().With("partition", strconv.Itoa(p.Index)))
		})
	}

	if len(result.Records) > 0 && p.Topic.Config.Bindings.Kafka.ValueSchemaValidation {
		return result, nil
	}

	for _, writeFunc := range writeFuncs {
		writeFunc()
	}

	if sequenceNumber >= 0 && producer != nil {
		state, ok := p.producers[producer.ProducerId]
		if !ok {
			state = &PartitionProducerState{LastSequence: sequenceNumber}
			p.producers[producer.ProducerId] = state
		} else {
			state.LastSequence = sequenceNumber
		}
	}

	return result, nil
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

func (p *Partition) addSegment() *Segment {
	p.m.RLock()
	defer p.m.RUnlock()

	now := time.Now()

	if active, ok := p.Segments[p.ActiveSegment]; ok && active.Closed.IsZero() {
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
		Log:    make([]*record, 0),
		Opened: time.Now(),
	}
}

func (s *Segment) contains(offset int64) bool {
	return offset >= s.Head && offset < s.Tail
}

func (s *Segment) record(offset int64) *kafka.Record {
	index := offset - s.Head
	return s.Log[index].Data
}

func (s *Segment) delete() {
	for _, r := range s.Log {
		log.Debugf("delete record: %v", r.Data.Offset)
		if r.Data.Key != nil {
			_ = r.Data.Key.Close()
		}
		if r.Data.Value != nil {
			_ = r.Data.Value.Close()
		}
		r.Log.Deleted = true
	}
}

func (r *WriteResult) fail(index int, code kafka.ErrorCode, msg string) {
	r.ErrorCode = code
	r.ErrorMessage = msg
	r.Records = append(r.Records, produce.RecordError{
		BatchIndex:             int32(index),
		BatchIndexErrorMessage: msg,
	})
}
