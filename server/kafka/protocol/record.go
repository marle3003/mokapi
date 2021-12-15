package protocol

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"time"
)

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

	// partition base offset of following records
	baseOffset := d.ReadInt64()
	d.ReadInt32()        // batchLength
	d.ReadInt32()        // leader epoch
	d.ReadInt8()         // magic
	crc := d.ReadInt32() // checksum
	attributes := Attributes(d.ReadInt16())
	d.ReadInt32() // lastOffsetDelta
	firstTimestamp := d.ReadInt64()
	d.ReadInt64() // maxTimestamp
	producerId := d.ReadInt64()
	producerEpoch := d.ReadInt16()
	d.ReadInt32() // baseSequence
	numRecords := d.ReadInt32()

	_ = crc
	_ = producerId
	_ = producerEpoch

	if attributes.Compression() != 0 {
		return fmt.Errorf("compression currently not supported")
	}

	rb.Records = make([]Record, numRecords)
	for i := range rb.Records {
		r := &rb.Records[i]
		l := d.ReadVarInt() // length
		_ = l
		d.ReadInt8() // attributes

		timestampDelta := d.ReadVarInt()
		offsetDelta := d.ReadVarInt()
		r.Offset = baseOffset + offsetDelta
		r.Time = toTime(firstTimestamp + timestampDelta)

		r.Key = d.ReadVarNullBytes()
		r.Value = d.ReadVarNullBytes()

		headerLen := d.ReadVarInt()
		if headerLen > 0 {
			r.Headers = make([]RecordHeader, headerLen)
			for i := range r.Headers {
				r.Headers[i] = RecordHeader{
					Key:   d.ReadString(),
					Value: d.ReadBytes(),
				}
			}
		}
	}

	return nil
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

func (r *Record) Size() int {
	return len(r.Key) + len(r.Value)
}

func (rb *RecordBatch) Size() (s int) {
	for _, r := range rb.Records {
		s += r.Size()
	}
	return
}

func (rb *RecordBatch) WriteTo(e *Encoder) {
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
