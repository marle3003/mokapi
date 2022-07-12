package store

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/kafka"
	"mokapi/runtime/events"
	"testing"
	"time"
)

func TestPartition(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(record kafka.Record, traits events.Traits) {}, &Topic{})

	require.Equal(t, 0, p.Index)
	require.Equal(t, int64(0), p.StartOffset())
	require.Equal(t, int64(0), p.Offset())
	require.Equal(t, 1, p.Leader)
	require.Equal(t, []int{1}, p.Replicas)
}

func TestPartition_Write(t *testing.T) {
	var log []kafka.Record
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(record kafka.Record, traits events.Traits) {
			log = append(log, record)
		}, &Topic{})

	offset, err := p.Write(kafka.RecordBatch{
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
		func(_ kafka.Record, _ events.Traits) {}, &Topic{})
	b, errCode := p.Read(0, 1)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Read(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, _ events.Traits) {}, &Topic{})
	offset, err := p.Write(kafka.RecordBatch{
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
	require.Equal(t, int64(0), offset)

	b, errCode := p.Read(1, 1)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Read_OutOfOffset_Empty(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, _ events.Traits) {}, &Topic{})
	b, errCode := p.Read(10, 1)
	require.Equal(t, kafka.None, errCode)
	require.Equal(t, 0, len(b.Records))
}

func TestPartition_Read_OutOfOffset(t *testing.T) {
	p := newPartition(
		0,
		map[int]*Broker{1: {Id: 1}},
		func(_ kafka.Record, _ events.Traits) {}, &Topic{})
	_, _ = p.Write(kafka.RecordBatch{
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
		func(_ kafka.Record, _ events.Traits) {}, &Topic{})
	p.validator = &validator{
		payload:     &schema.Ref{Value: schematest.New("string")},
		contentType: "application/json",
	}

	offset, err := p.Write(kafka.RecordBatch{
		Records: []kafka.Record{
			{
				Time:    time.Now(),
				Key:     kafka.NewBytes([]byte(`"foo-1"`)),
				Value:   kafka.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})

	require.Error(t, err, "expected string got float64")
	require.Equal(t, int64(0), offset)
	require.Equal(t, int64(0), p.Offset())
	require.Equal(t, int64(0), p.StartOffset())
}
