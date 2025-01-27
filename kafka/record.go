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
	Records []*Record
}

func NewRecordBatch() RecordBatch {
	return RecordBatch{Records: make([]*Record, 0)}
}

func (rb *RecordBatch) ReadFrom(d *Decoder, version int16, tag kafkaTag) error {
	var size int64
	if version != 0 && version >= tag.compact {
		size = int64(d.ReadUvarint())
	} else {
		size = int64(d.ReadInt32())
	}

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
	Offset  int64          `json:"offset"`
	Time    time.Time      `json:"time"`
	Key     Bytes          `json:"key"`
	Value   Bytes          `json:"value"`
	Headers []RecordHeader `json:"headers"`
}

type RecordHeader struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

func (r *Record) Size(baseOffSet int64, base time.Time) int {
	t := Timestamp(r.Time)
	if t == 0 {
		t = Timestamp(time.Now())
	}
	deltaTimestamp := t - Timestamp(base)
	deltaOffset := r.Offset - baseOffSet
	keyLength := 0
	if r.Key != nil {
		keyLength = r.Key.Size()
	}
	valueLength := 0
	if r.Value != nil {
		valueLength = r.Value.Size()
	}
	size := 1 + // attribute
		sizeVarInt(deltaTimestamp) +
		sizeVarInt(deltaOffset) +
		sizeVarInt(int64(keyLength)) + keyLength +
		sizeVarInt(int64(valueLength)) + valueLength +
		sizeVarInt(int64(len(r.Headers)))

	for _, h := range r.Headers {
		k := len(h.Key)
		v := len(h.Value)
		size += sizeVarInt(int64(k)) + k +
			sizeVarInt(int64(v)) + v
	}

	return size
}

func (rb *RecordBatch) Size() (s int) {
	if len(rb.Records) == 0 {
		return 0
	}

	t := rb.Records[0].Time
	o := rb.Records[0].Offset

	s = 60
	for _, r := range rb.Records {
		s += r.Size(o, t)
	}
	return
}

func (rb *RecordBatch) WriteTo(e *Encoder, version int16, tag kafkaTag) {
	isCompact := version != 0 && version >= tag.compact

	if len(rb.Records) == 0 {
		if isCompact {
			e.writeUVarInt(0)
		} else {
			e.writeInt32(0)
		}
		return
	}

	b := newPageBuffer()
	rb.writeTo(NewEncoder(b))

	if isCompact {
		messageSetSize := b.Size() + 1
		e.writeUVarInt(uint64(messageSetSize))
	} else {
		messageSetSize := b.Size()
		e.writeInt32(int32(messageSetSize))
	}
	_, err := b.WriteTo(e.writer)
	if err != nil {
		panic(err)
	}
}

func (rb *RecordBatch) writeTo(e *Encoder) {
	offset := e.writer.Size()
	buffer := make([]byte, 8)

	e.writeInt64(rb.Records[0].Offset)   // base offset
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

	firstTime := rb.Records[0].Time
	if firstTime.IsZero() {
		firstTime = time.Now()
	}
	firstTimestamp := Timestamp(firstTime)
	maxTimestamp := int64(0)
	lastOffSetDetla := uint32(0)

	// records must be sorted by time
	for i, r := range rb.Records {
		if r.Time.IsZero() {
			r.Time = firstTime
		}
		t := Timestamp(r.Time)
		if t > maxTimestamp {
			maxTimestamp = t
		}

		deltaTimestamp := t - firstTimestamp
		deltaOffset := int64(i)
		lastOffSetDetla = uint32(i)

		e.writeVarInt(int64(r.Size(int64(i), firstTime)))
		e.writeInt8(0) // attributes
		e.writeVarInt(deltaTimestamp)
		e.writeVarInt(deltaOffset)

		e.writeVarNullBytesFrom(r.Key)
		e.writeVarNullBytesFrom(r.Value)

		e.writeVarInt(int64(len(r.Headers)))
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
}

func sizeVarInt(x int64) int {
	// code from binary.PutVarint
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	i := 0
	for x >= 0x80 {
		x >>= 7
		i++
	}
	return i + 1
}
