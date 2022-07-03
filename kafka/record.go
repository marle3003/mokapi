package kafka

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"time"
)

const magicOffset = 16

type Attributes int16

type RecordBatch struct {
	Records []Record
}

func NewRecordBatch() RecordBatch {
	return RecordBatch{Records: make([]Record, 0)}
}

func (rb *RecordBatch) ReadFrom(d *Decoder) error {
	size := d.ReadInt32()
	if size <= 0 {
		return nil
	}

	b := make([]byte, magicOffset+1)
	if _, err := io.ReadFull(d.reader, b); err != nil {
		return err
	}
	magic := b[magicOffset]
	// add read bytes back to reader
	d.reader = io.MultiReader(bytes.NewReader(b), d.reader)

	switch magic {
	case 0:
		return rb.readFromV0(d)
	case 1:
		return rb.readFromV1(d)
	case 2:
		return rb.readFromV2(d)
	default:
		return fmt.Errorf("unsupported record version %v", magic)
	}
}

type Record struct {
	Offset  int64
	Time    time.Time
	Key     Bytes
	Value   Bytes
	Headers []RecordHeader
}

type RecordHeader struct {
	Key   string
	Value []byte
}

func (r *Record) Size() (s int) {
	r.Key.Seek(0, io.SeekStart)
	r.Value.Seek(0, io.SeekStart)
	if r.Key != nil {
		s += r.Key.Len()
	}
	if r.Value != nil {
		s += r.Value.Len()
	}
	return
}

func (rb *RecordBatch) Size() (s int) {
	for _, r := range rb.Records {
		s += r.Size()
	}
	return
}

func (rb *RecordBatch) WriteTo(e *Encoder) {
	if len(rb.Records) == 0 {
		// send only size of records
		e.writeInt32(0) // size
		return
	}

	offsetBatchSize := e.writer.Size()
	e.writeInt32(0) // placeholder length

	offset := e.writer.Size()
	buffer := make([]byte, 8)

	e.writeInt64(0)                      // base offset
	e.writeInt32(0)                      // size: 8
	e.writeInt32(0)                      // leader epoch
	e.writeInt8(2)                       // magic
	e.writeInt32(0)                      // checksum: 17
	e.writeInt16(int16(0))               // 21
	e.writeInt32(0)                      // last offset delta: 23
	e.writeInt64(0)                      // first timestamp: 27
	e.writeInt64(0)                      // max timestamp: 35
	e.writeInt64(-1)                     // Producer Id
	e.writeInt16(-1)                     // producer epoch
	e.writeInt32(-1)                     // base sequence
	e.writeInt32(int32(len(rb.Records))) // num records

	firstTimestamp := int64(0)
	maxTimestamp := int64(0)
	lastOffSetDetla := uint32(0)

	// records must be sorted by time
	for i, r := range rb.Records {
		t := Timestamp(r.Time)
		if t == 0 {
			t = Timestamp(time.Now())
		}
		if i == 0 {
			firstTimestamp = t
		}
		if t > maxTimestamp {
			maxTimestamp = t
		}

		deltaTimestamp := t - firstTimestamp
		deltaOffset := int64(i)
		lastOffSetDetla = uint32(i)

		keyLength := r.Key.Len()
		valueLength := r.Value.Len()
		headerLength := len(r.Headers)

		size := binary.PutUvarint(buffer, uint64(deltaTimestamp)) +
			1 + // attribute
			binary.PutVarint(buffer, deltaOffset) +
			binary.PutVarint(buffer, int64(keyLength)) + keyLength +
			binary.PutVarint(buffer, int64(valueLength)) + valueLength +
			binary.PutVarint(buffer, int64(headerLength))

		for _, h := range r.Headers {
			k := len(h.Key)
			v := len(h.Value)
			size += binary.PutVarint(buffer, int64(k)) + k +
				binary.PutVarint(buffer, int64(v)) + v
		}

		e.writeVarInt(int64(size))
		e.writeInt8(0) // attributes
		e.writeVarInt(deltaTimestamp)
		e.writeVarInt(deltaOffset)

		e.writeVarNullBytes(r.Key)
		e.writeVarNullBytes(r.Value)

		e.writeVarInt(int64(headerLength))

		for _, h := range r.Headers {
			e.writeVarString(h.Key)
			e.writeVarNullBytes_Old(h.Value)
		}
	}

	binary.BigEndian.PutUint32(buffer[:4], lastOffSetDetla)
	e.writer.WriteAt(buffer[:4], offset+23)

	binary.BigEndian.PutUint64(buffer[:8], uint64(firstTimestamp))
	e.writer.WriteAt(buffer[:8], offset+27)

	binary.BigEndian.PutUint64(buffer[:8], uint64(maxTimestamp))
	e.writer.WriteAt(buffer[:8], offset+35)

	totalLength := e.writer.Size() - offset
	batchSize := totalLength - 12 // offset(8) + size(4)
	binary.BigEndian.PutUint32(buffer[:4], uint32(batchSize))
	e.writer.WriteAt(buffer[:4], offset+8)

	checksum := uint32(0)
	crcTable := crc32.MakeTable(crc32.Castagnoli)
	// checksum from attributes to end
	e.writer.Scan(offset+21, offset+totalLength, func(chunk []byte) bool {
		checksum = crc32.Update(checksum, crcTable, chunk)
		return true
	})

	binary.BigEndian.PutUint32(buffer[:4], checksum)
	e.writer.WriteAt(buffer[:4], offset+17)

	size := e.writer.Size() - offsetBatchSize - 4
	e.writer.WriteSizeAt(size, offsetBatchSize)
}
