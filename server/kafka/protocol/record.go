package protocol

import (
	"encoding/binary"
	"time"
)

type Attributes int16

type RecordSet struct {
	Batches []RecordBatch
}

func (rs *RecordSet) ReadFrom(d *Decoder) {

}

func (rs *RecordSet) WriteTo(e *Encoder) {
	offset := e.writer.Size()

	e.writeInt32(0) // size: 8

	for _, rb := range rs.Batches {
		rb.WriteTo(e)
	}

	size := e.writer.Size() - offset - 4
	e.writer.WriteSizeAt(size, offset)
}

type RecordBatch struct {
	Offset     int64
	Attributes Attributes
	ProducerId int64
	Records    []Record
}

type Record struct {
	Offset  int64
	Time    time.Time
	Key     []byte
	Value   []byte
	Headers []RecordHeader
}

type RecordHeader struct {
	Key   string
	Value []byte
}

func (rb *RecordBatch) Size() (s int32) {
	for _, r := range rb.Records {
		s += int32(len(r.Key))
		s += int32(len(r.Value))
	}
	return
}

func (rb *RecordBatch) WriteTo(e *Encoder) {
	offset := e.writer.Size()
	buffer := make([]byte, 8)

	e.writeInt64(rb.Offset)              // offset
	e.writeInt32(0)                      // size: 8
	e.writeInt32(-1)                     // leader epoch
	e.writeInt8(2)                       // magic
	e.writeInt32(0)                      // checksum: 17
	e.writeInt16(int16(rb.Attributes))   // 21
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

		keyLength := len(r.Key)
		valueLength := len(r.Value)
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
			e.writeVarNullBytes(h.Value)
		}
	}

	binary.BigEndian.PutUint32(buffer[:4], lastOffSetDetla)
	e.writer.WriteAt(buffer[:4], offset+23)

	binary.BigEndian.PutUint64(buffer[:8], uint64(firstTimestamp))
	e.writer.WriteAt(buffer[:8], offset+27)

	binary.BigEndian.PutUint64(buffer[:8], uint64(maxTimestamp))
	e.writer.WriteAt(buffer[:8], offset+35)

	size := e.writer.Size() - offset - 12 // offset(8) + size(4)
	binary.BigEndian.PutUint32(buffer[:4], uint32(size))
	e.writer.WriteAt(buffer[:4], offset+8)

	checksum := e.writer.Checksum(int64(offset+21), int64(offset+size)) // checksum from attributes to end
	binary.BigEndian.PutUint32(buffer[:4], checksum)
	e.writer.WriteAt(buffer[:4], offset+17)
}

func Timestamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano() / int64(time.Millisecond)
}
