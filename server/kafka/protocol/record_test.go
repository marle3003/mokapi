package protocol

import (
	"bytes"
	"mokapi/test"
	"testing"
	"time"
)

func TestRecord_ReadFrom(t *testing.T) {
	testdata := []struct {
		name string
		b    []byte
		fn   func(*testing.T, *Decoder)
	}{
		{
			"empty",
			[]byte{},
			func(t *testing.T, d *Decoder) {
				record := RecordBatch{}
				err := record.ReadFrom(d)
				test.Ok(t, err)
			},
		},
		{
			"size zero",
			[]byte{0, 0, 0, 0},
			func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				test.Ok(t, err)
			},
		},
		{
			"one record",
			[]byte{
				0, 0, 0, 81, // length
				0, 0, 0, 0, 0, 0, 0, 12, // base offset
				0, 0, 0, 0, // batch length
				0, 0, 0, 0, // leader epoch
				2,          // magic
				0, 0, 0, 0, // crc
				0, 0, // attributes
				0, 0, 0, 0, // last offset delta
				0, 0, 1, 125, 158, 189, 76, 76, // first timestamp
				0, 0, 0, 0, 0, 0, 0, 0, // max timestamp
				0, 0, 0, 0, 0, 0, 0, 0, // producer id
				0, 0, // producer epoch
				0, 0, 0, 0, // base sequence
				0, 0, 0, 1, // number of records
				24,               // record length 12
				0,                // attributes
				0,                // delta timestamp
				2,                // delta offset 1
				6, 'f', 'o', 'o', // key
				6, 'b', 'a', 'r', // value
				0, // header
			},
			func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				test.Ok(t, err)
				test.Equals(t, 1, len(batch.Records))
				record := batch.Records[0]
				test.Equals(t, int64(13), record.Offset)
				test.Equals(t, "foo", string(record.Key))
				test.Equals(t, "bar", string(record.Value))
				y, m, day := record.Time.Date()
				test.Equals(t, 2021, y)
				test.Equals(t, time.December, m)
				test.Equals(t, 9, day)
			},
		},
		{
			"one record",
			[]byte{
				0, 0, 0, 81, // length
				0, 0, 0, 0, 0, 0, 0, 12, // base offset
				0, 0, 0, 0, // batch length
				0, 0, 0, 0, // leader epoch
				2,          // magic
				0, 0, 0, 0, // crc
				0, 0, // attributes
				0, 0, 0, 0, // last offset delta
				0, 0, 1, 125, 158, 189, 76, 76, // first timestamp
				0, 0, 0, 0, 0, 0, 0, 0, // max timestamp
				0, 0, 0, 0, 0, 0, 0, 0, // producer id
				0, 0, // producer epoch
				0, 0, 0, 0, // base sequence
				0, 0, 0, 2, // number of records
				24,               // record length 12
				0,                // attributes
				0,                // delta timestamp
				2,                // delta offset 1
				6, 'f', 'o', 'o', // key
				6, 'b', 'a', 'r', // value
				0,                // header
				24,               // record length 12
				0,                // attributes
				0,                // delta timestamp
				2,                // delta offset 1
				6, 'f', 'o', 'o', // key
				6, 'b', 'a', 'r', // value
				0, // header
			},
			func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				test.Ok(t, err)
				test.Equals(t, 2, len(batch.Records))
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			d := NewDecoder(bytes.NewReader(data.b), len(data.b))
			t.Run(data.name, func(t *testing.T) {
				data.fn(t, d)
			})
		})
	}
}

func TestRecordBatch_WriteTo(t *testing.T) {
	testdata := []struct {
		name  string
		batch RecordBatch
	}{
		{
			"empty batch",
			NewRecordBatch(),
		},
		{
			"single record",
			RecordBatch{Records: []Record{
				{
					Time:    toTime(Timestamp(time.Now())),
					Key:     []byte("foo"),
					Value:   []byte("bar"),
					Headers: nil,
				},
			}},
		},
		{
			"two records",
			RecordBatch{Records: []Record{
				{
					Time:    toTime(Timestamp(time.Now())),
					Key:     []byte("key-1"),
					Value:   []byte("value-1"),
					Headers: nil,
				},
				{
					Offset:  1,
					Time:    toTime(Timestamp(time.Now())),
					Key:     []byte("key-2"),
					Value:   []byte("value-2"),
					Headers: nil,
				},
			}},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			pb := newPageBuffer()

			e := NewEncoder(pb)
			data.batch.WriteTo(e)
			var buf bytes.Buffer
			n, err := pb.WriteTo(&buf)
			test.Ok(t, err)
			test.Assert(t, n > 0, "written should not be 0")

			d := NewDecoder(bytes.NewReader(buf.Bytes()), pb.Size())
			batch := RecordBatch{}
			err = batch.ReadFrom(d)
			test.Ok(t, err)
			test.Equals(t, data.batch, batch)
		})
	}
}

func TestRecord_Size(t *testing.T) {
	testdata := []struct {
		name  string
		size  int
		batch RecordBatch
	}{
		{
			"empty batch",
			0,
			NewRecordBatch(),
		},
		{
			"single record",
			6,
			RecordBatch{Records: []Record{
				{
					Time:    toTime(Timestamp(time.Now())),
					Key:     []byte("foo"),
					Value:   []byte("bar"),
					Headers: nil,
				},
			}},
		},
		{
			"two records",
			24,
			RecordBatch{Records: []Record{
				{
					Time:    toTime(Timestamp(time.Now())),
					Key:     []byte("key-1"),
					Value:   []byte("value-1"),
					Headers: nil,
				},
				{
					Offset:  1,
					Time:    toTime(Timestamp(time.Now())),
					Key:     []byte("key-2"),
					Value:   []byte("value-2"),
					Headers: nil,
				},
			}},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			test.Equals(t, data.size, data.batch.Size())
		})
	}
}
