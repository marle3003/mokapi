package kafka

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func bytesToString(bytes Bytes) string {
	var b []byte
	bytes.Read(b)
	return string(b)
}

func TestRecord_ReadFrom(t *testing.T) {
	testcases := []struct {
		name string
		data []byte
		test func(*testing.T, *Decoder)
	}{
		{
			name: "empty",
			data: []byte{},
			test: func(t *testing.T, d *Decoder) {
				record := RecordBatch{}
				err := record.ReadFrom(d)
				require.NoError(t, err)
			},
		},
		{
			name: "size zero",
			data: []byte{0, 0, 0, 0},
			test: func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				require.NoError(t, err)
			},
		},
		{
			name: "one record v2",
			data: []byte{
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
				0, // header length
			},
			test: func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				require.NoError(t, err)
				require.Len(t, batch.Records, 1)
				record := batch.Records[0]
				require.Equal(t, int64(13), record.Offset)
				var b [3]byte
				record.Key.Read(b[:])
				require.Equal(t, "foo", string(b[:]))
				record.Value.Read(b[:])
				require.Equal(t, "bar", string(b[:]))
				y, m, day := record.Time.Date()
				require.Equal(t, 2021, y)
				require.Equal(t, time.December, m)
				require.Equal(t, 9, day)
			},
		},
		{
			name: "two record",
			data: []byte{
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
				0, // header length
			},
			test: func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				require.NoError(t, err)
				require.Len(t, batch.Records, 2)
			},
		},
		{
			name: "with header v2",
			data: []byte{
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
				6,                                                         // header length
				22, 't', 'r', 'a', 'c', 'e', 'p', 'a', 'r', 'e', 'n', 't', // first header name
				6, 'f', 'o', 'o', // first header value
				24, 'x', '-', 'm', 'e', 's', 's', 'a', 'g', 'e', '-', 'i', 'd', // 2nd header name
				6, 'b', 'a', 'r', // 2nd header value
				28, 'x', '-', 'm', 'e', 's', 's', 'a', 'g', 'e', '-', 't', 'y', 'p', 'e', // 2nd header name
				6, 'o', 'n', 'e', // 2nd header value
			},
			test: func(t *testing.T, d *Decoder) {
				batch := RecordBatch{}
				err := batch.ReadFrom(d)
				require.NoError(t, err)
				require.Len(t, batch.Records[0].Headers, 3)
				require.Equal(t, "traceparent", batch.Records[0].Headers[0].Key)
				require.Equal(t, []byte("foo"), batch.Records[0].Headers[0].Value)
				require.Equal(t, "x-message-id", batch.Records[0].Headers[1].Key)
				require.Equal(t, []byte("bar"), batch.Records[0].Headers[1].Value)
				require.Equal(t, "x-message-type", batch.Records[0].Headers[2].Key)
				require.Equal(t, []byte("one"), batch.Records[0].Headers[2].Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := bufio.NewReader(bytes.NewReader(tc.data))
			d := NewDecoder(r, len(tc.data))
			tc.test(t, d)
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
					Time:    ToTime(Timestamp(time.Now())),
					Key:     NewBytes([]byte("foo")),
					Value:   NewBytes([]byte("bar")),
					Headers: nil,
				},
			}},
		},
		{
			"two records",
			RecordBatch{Records: []Record{
				{
					Offset:  2,
					Time:    ToTime(Timestamp(time.Now())),
					Key:     NewBytes([]byte("key-1")),
					Value:   NewBytes([]byte("value-1")),
					Headers: nil,
				},
				{
					Offset:  3,
					Time:    ToTime(Timestamp(time.Now())),
					Key:     NewBytes([]byte("key-2")),
					Value:   NewBytes([]byte("value-2")),
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
			require.NoError(t, err)
			require.Greater(t, n, 0, "written should not be 0")

			r := bufio.NewReader(bytes.NewReader(buf.Bytes()))
			d := NewDecoder(r, pb.Size())
			batch := RecordBatch{}
			err = batch.ReadFrom(d)
			require.NoError(t, err)

			require.Equal(t, len(data.batch.Records), len(batch.Records))
			for i := 0; i < len(data.batch.Records); i++ {
				require.Equal(t, data.batch.Records[i].Time, batch.Records[i].Time)
				require.Equal(t, data.batch.Records[i].Offset, batch.Records[i].Offset)
				require.Equal(t, bytesToString(data.batch.Records[i].Key), bytesToString(batch.Records[i].Key))
				require.Equal(t, bytesToString(data.batch.Records[i].Value), bytesToString(batch.Records[i].Value))
			}
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
					Time:    ToTime(Timestamp(time.Now())),
					Key:     NewBytes([]byte("foo")),
					Value:   NewBytes([]byte("bar")),
					Headers: nil,
				},
			}},
		},
		{
			"two records",
			24,
			RecordBatch{Records: []Record{
				{
					Time:    ToTime(Timestamp(time.Now())),
					Key:     NewBytes([]byte("key-1")),
					Value:   NewBytes([]byte("value-1")),
					Headers: nil,
				},
				{
					Offset:  1,
					Time:    ToTime(Timestamp(time.Now())),
					Key:     NewBytes([]byte("key-2")),
					Value:   NewBytes([]byte("value-2")),
					Headers: nil,
				},
			}},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			require.Equal(t, data.size, data.batch.Size())
		})
	}
}

func TestRecordBatch_WriteTo_Bytes_Compare(t *testing.T) {
	records := RecordBatch{Records: []Record{
		{
			Offset:  0,
			Time:    ToTime(1657010762684),
			Key:     NewBytes([]byte("foo")),
			Value:   NewBytes([]byte("bar")),
			Headers: nil,
		},
	}}

	pb := newPageBuffer()

	e := NewEncoder(pb)
	records.WriteTo(e)
	var buf bytes.Buffer
	_, err := pb.WriteTo(&buf)
	require.NoError(t, err)

	b := buf.Bytes()

	expected := []byte{
		0, 0, 0, 74, // length: len - 4
		0, 0, 0, 0, 0, 0, 0, 0, //  base offset
		0, 0, 0, 62, // message size: length - base offset - message size
		0, 0, 0, 0, // leader epoch
		2,                // magic
		119, 89, 114, 22, // crc32
		0, 0, // attributes
		0, 0, 0, 0, // last offset delta
		0, 0, 1, 129, 205, 137, 179, 188, // first timestamp
		0, 0, 1, 129, 205, 137, 179, 188, // max timestamp
		255, 255, 255, 255, 255, 255, 255, 255, // producer id
		255, 255, // producer epoch
		255, 255, 255, 255, // base sequence
		0, 0, 0, 1, // number of records
		24,               // record length 12
		0,                // attributes
		0,                // delta timestamp
		0,                // delta offset 1
		6, 'f', 'o', 'o', // key
		6, 'b', 'a', 'r', // value
		0, // header
	}

	require.Equal(t, expected, b)
}
