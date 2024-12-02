package store

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/kafka"
	"mokapi/runtime/events"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
	"time"
)

func TestPartition(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(record kafka.Record, partition int, traits events.Traits) {}, func(record *kafka.Record) {}, &Topic{})

	require.Equal(t, 0, p.Index)
	require.Equal(t, int64(0), p.StartOffset())
	require.Equal(t, int64(0), p.Offset())
	require.Equal(t, 1, p.Leader)
	require.Equal(t, []int{}, p.Replicas)
}

func TestPartition_Write(t *testing.T) {
	var log []kafka.Record
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(record kafka.Record, partition int, traits events.Traits) {
			log = append(log, record)
		}, func(record *kafka.Record) {}, &Topic{})

	offset, records, err := p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`"bar-1"`)),
				Headers: nil,
			},
			{
				Offset:  0,
				Key:     kafka.NewBytes([]byte(`"foo-2"`)),
				Value:   kafka.NewBytes([]byte(`"bar-2"`)),
				Headers: nil,
			},
		},
	})

	require.NoError(t, err)
	require.Len(t, records, 0)
	require.Equal(t, int64(0), offset)
	require.Equal(t, int64(2), p.Offset())
	require.Equal(t, int64(0), p.StartOffset())

	b, errCode := p.Read(0, 14)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 1, len(b.Records))

	b, errCode = p.Read(0, 100)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 2, len(b.Records))

	require.False(t, b.Records[0].Time.IsZero(), "time is set")

	require.Len(t, log, 2)
}

func TestPartition_Read_Empty(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, partition int, _ events.Traits) {}, func(record *kafka.Record) {}, &Topic{})
	b, errCode := p.Read(0, 1)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Read(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, partition int, _ events.Traits) {}, func(record *kafka.Record) {}, &Topic{})
	offset, records, err := p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, records, 0)
	require.Equal(t, int64(0), offset)

	b, errCode := p.Read(1, 1)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Read_OutOfOffset_Empty(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, partition int, _ events.Traits) {}, func(record *kafka.Record) {}, &Topic{})
	b, errCode := p.Read(10, 1)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Read_OutOfOffset(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, partition int, _ events.Traits) {}, func(record *kafka.Record) {}, &Topic{})
	_, _, _ = p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})

	b, errCode := p.Read(-10, 1)
	require.Equal(t, kafka.OffsetOutOfRange, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Write_Value_Validator(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, partition int, _ events.Traits) {}, func(record *kafka.Record) {}, &Topic{config: asyncApi.TopicBindings{ValueSchemaValidation: true}})
	p.validator = &validator{
		payload:     &schema.Ref{Value: schematest.New("string")},
		contentType: "application/json",
	}

	offset, records, err := p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})

	require.EqualError(t, err, "validation error")
	require.Len(t, records, 1)
	require.Equal(t, int32(0), records[0].BatchIndex)
	require.Equal(t, "parse 12 failed, expected schema type=string", records[0].BatchIndexErrorMessage)
	require.Equal(t, int64(0), offset)
	require.Equal(t, int64(0), p.Offset())
	require.Equal(t, int64(0), p.StartOffset())
}

func TestPatition_Retention(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, partition int, _ events.Traits) {},
		func(record *kafka.Record) {}, &Topic{})
	require.Equal(t, int64(0), p.Head)
	offset, records, err := p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})
	offset, records, err = p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, records, 0)
	require.Equal(t, int64(1), offset)
	require.Equal(t, int64(0), p.Head)
	require.Equal(t, int64(2), p.Tail)

	// rolling
	p.addSegment()
	require.Len(t, p.Segments, 2)
	require.False(t, p.Segments[0].Closed.IsZero())

	// retention
	p.removeClosedSegments()
	require.Len(t, p.Segments, 1)
	require.Equal(t, int64(2), p.Head)
}
