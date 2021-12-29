package store

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/kafka/protocol"
	"mokapi/test"
	"testing"
	"time"
)

func TestPartition(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {id: 1}})

	test.Equals(t, 0, p.Index())
	test.Equals(t, int64(-1), p.StartOffset())
	test.Equals(t, int64(-1), p.Offset())
	test.Equals(t, 1, p.Leader())
	test.Equals(t, []int{1}, p.Replicas())
}

func TestPartition_Write(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {id: 1}})

	offset, err := p.Write(protocol.RecordBatch{
		Records: []protocol.Record{
			{
				Time:    time.Now(),
				Key:     protocol.NewBytes([]byte(`"foo-1"`)),
				Value:   protocol.NewBytes([]byte(`"bar-1"`)),
				Headers: nil,
			},
			{
				Offset:  0,
				Key:     protocol.NewBytes([]byte(`"foo-2"`)),
				Value:   protocol.NewBytes([]byte(`"bar-2"`)),
				Headers: nil,
			},
		},
	})

	test.Ok(t, err)
	test.Equals(t, int64(0), offset)
	test.Equals(t, int64(1), p.Offset())
	test.Equals(t, int64(0), p.StartOffset())

	b, errCode := p.Read(0, 1)
	test.Equals(t, protocol.None, errCode)
	test.Equals(t, 1, len(b.Records))

	b, errCode = p.Read(0, 100)
	test.Equals(t, protocol.None, errCode)
	test.Equals(t, 2, len(b.Records))

	test.Assert(t, !b.Records[0].Time.IsZero(), "time is set")
}

func TestPartition_Read_Empty(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {id: 1}})
	b, errCode := p.Read(0, 1)
	test.Equals(t, protocol.None, errCode)
	test.Equals(t, 0, len(b.Records))
}

func TestPartition_Read_OutOfOffset_Empty(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {id: 1}})
	b, errCode := p.Read(10, 1)
	test.Equals(t, protocol.None, errCode)
	test.Equals(t, 0, len(b.Records))
}

func TestPartition_Read_OutOfOffset(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {id: 1}})
	_, _ = p.Write(protocol.RecordBatch{
		Records: []protocol.Record{
			{
				Time:    time.Now(),
				Key:     protocol.NewBytes([]byte(`"foo-1"`)),
				Value:   protocol.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})

	b, errCode := p.Read(10, 1)
	test.Equals(t, protocol.OffsetOutOfRange, errCode)
	test.Equals(t, 0, len(b.Records))
}

func TestPartition_Write_Value_Validator(t *testing.T) {
	p := newPartition(0, map[int]*Broker{1: {id: 1}})
	p.validator = &validator{
		payload:     &openapi.SchemaRef{Value: openapitest.NewSchema("string")},
		contentType: "application/json",
	}

	offset, err := p.Write(protocol.RecordBatch{
		Records: []protocol.Record{
			{
				Time:    time.Now(),
				Key:     protocol.NewBytes([]byte(`"foo-1"`)),
				Value:   protocol.NewBytes([]byte(`12`)),
				Headers: nil,
			},
		},
	})

	test.EqualError(t, "expected string got float64", err)
	test.Equals(t, int64(-1), offset)
	test.Equals(t, int64(-1), p.Offset())
	test.Equals(t, int64(-1), p.StartOffset())
}
