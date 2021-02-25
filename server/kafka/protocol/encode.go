package protocol

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"reflect"
	"sync"
)

type encodeFunc func(*Encoder, reflect.Value)

type BufferWriter interface {
	io.Writer
	WriteAt([]byte, int)
	WriteSizeAt(size int, offset int)
	Size() int
	Checksum(start, end int64) uint32
}

type WriterTo interface {
	WriteTo(e *Encoder)
}

type page struct {
	offset int
	buffer [65536]byte
}

func (p *page) Write(b []byte) (n int, err error) {
	n = copy(p.buffer[p.offset:], b)
	p.offset += n
	return
}

func (p *page) WriteSizeAt(size int, offset int) {
	binary.BigEndian.PutUint32(p.buffer[offset:offset+4], uint32(size))
}

func (p *page) WriteAt(b []byte, offset int) {
	copy(p.buffer[offset:], b)
}

func (p *page) Size() int {
	return p.offset
}

func (p *page) Checksum(start, end int64) uint32 {
	table := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(p.buffer[start:end], table)
}

var (
	pagePool = sync.Pool{New: func() interface{} { return new(page) }}
	writerTo = reflect.TypeOf((*WriterTo)(nil)).Elem()
)

type Encoder struct {
	writer BufferWriter
	buffer [32]byte
}

func NewEncoder(w BufferWriter) *Encoder {
	return &Encoder{writer: w}
}

func newEncodeFunc(t reflect.Type, version int16, tag kafkaTag) encodeFunc {
	if reflect.PtrTo(t).Implements(writerTo) {
		return func(e *Encoder, v reflect.Value) {
			v.Addr().Interface().(WriterTo).WriteTo(e)
		}
	}

	switch t.Kind() {
	case reflect.Struct:
		return newStructEncodeFunc(t, version, tag)
	case reflect.String:
		if version >= tag.compact && tag.nullable {
			return (*Encoder).encodeCompactNullString
		} else if version >= tag.compact {
			return (*Encoder).encodeCompactString
		} else if tag.nullable {
			return (*Encoder).encodeNullString
		}
		return (*Encoder).encodeString
	case reflect.Int64:
		return (*Encoder).encodeInt64
	case reflect.Int32:
		return (*Encoder).encodeInt32
	case reflect.Int16:
		return (*Encoder).encodeInt16
	case reflect.Int8:
		return (*Encoder).encodeInt8
	case reflect.Bool:
		return (*Encoder).encodeBool
	case reflect.Map:
		if tag.protoType == "TAG_BUFFER" {
			return (*Encoder).encodeTagBuffer
		}
		panic("unsupported map: " + t.String())
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 { // []byte
			return newBytesEncodeFunc(version, tag)
		}
		return newEncodeArray(t, version, tag)
	default:
		panic("unsupported type: " + t.String())
	}
}

func newBytesEncodeFunc(version int16, tag kafkaTag) encodeFunc {
	switch {
	case version >= tag.compact && tag.nullable:
		return (*Encoder).encodeCompactNullBytes
	case version >= tag.compact:
		return (*Encoder).encodeCompactBytes
	case tag.nullable:
		return (*Encoder).encodeNullBytes
	default:
		return (*Encoder).encodeBytes
	}
}

func newEncodeArray(t reflect.Type, version int16, tag kafkaTag) encodeFunc {
	elemType := t.Elem()
	elemFunc := newEncodeFunc(elemType, version, tag)

	switch {
	case tag.compact <= version:
		return func(e *Encoder, v reflect.Value) { e.encodeCompactArray(v, elemFunc) }
	default:
		return func(e *Encoder, v reflect.Value) { e.encodeArray(v, elemFunc) }
	}
}

func newStructEncodeFunc(t reflect.Type, version int16, tag kafkaTag) encodeFunc {
	type field struct {
		index  int
		encode encodeFunc
	}
	fields := make([]*field, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := getTag(f)
		if !tag.isValid(version) {
			continue
		}
		fields = append(fields, &field{i, newEncodeFunc(f.Type, version, tag)})
	}

	return func(e *Encoder, v reflect.Value) {
		for _, f := range fields {
			f.encode(e, v.Field(f.index))
		}
	}
}

func (e *Encoder) encodeCompactArray(v reflect.Value, encodeElem encodeFunc) {
	e.writeUVarInt(uint64(v.Len() + 1))

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		encodeElem(e, item)
	}
}

func (e *Encoder) encodeArray(v reflect.Value, encodeElem encodeFunc) {
	e.writeInt32(int32(v.Len()))

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		encodeElem(e, item)
	}
}

func (e *Encoder) encodeCompactBytes(v reflect.Value) {
	e.writeCompactBytes(v.Bytes())
}

func (e *Encoder) encodeCompactNullBytes(v reflect.Value) {
	e.writeCompactNullBytes(v.Bytes())
}

func (e *Encoder) encodeNullBytes(v reflect.Value) {
	e.writeNullBytes(v.Bytes())
}

func (e *Encoder) encodeBytes(v reflect.Value) {
	e.writeBytes(v.Bytes())
}

func (e *Encoder) encodeCompactString(v reflect.Value) {
	e.writeCompactString(v.String())
}

func (e *Encoder) encodeCompactNullString(v reflect.Value) {
	e.writeCompactNullString(v.String())
}

func (e *Encoder) encodeNullString(v reflect.Value) {
	e.writeNullString(v.String())
}

func (e *Encoder) encodeString(v reflect.Value) {
	e.writeNullString(v.String())
}

func (e *Encoder) encodeInt8(v reflect.Value) {
	e.writeInt8(int8(v.Int()))
}

func (e *Encoder) encodeInt16(v reflect.Value) {
	e.writeInt16(int16(v.Int()))
}

func (e *Encoder) encodeBool(v reflect.Value) {
	e.writeBool(v.Bool())
}

func (e *Encoder) encodeInt32(v reflect.Value) {
	e.writeInt32(int32(v.Int()))
}

func (e *Encoder) encodeInt64(v reflect.Value) {
	e.writeInt64(v.Int())
}

func (e *Encoder) writeCompactBytes(b []byte) {
	e.writeUVarInt(uint64(len(b)) + 1)
	e.write(b)
}

func (e *Encoder) writeNullBytes(b []byte) {
	if b == nil {
		e.writeInt32(-1)
	} else {
		e.writeInt32(int32(len(b)))
		e.write(b)
	}
}

func (e *Encoder) writeCompactNullBytes(b []byte) {
	if b == nil {
		e.writeUVarInt(0)
	} else {
		e.writeUVarInt(uint64(len(b)) + 1)
		e.write(b)
	}
}

func (e *Encoder) writeVarNullBytes(b []byte) {
	if b == nil {
		e.writeVarInt(-1)
	} else {
		e.writeVarInt(int64(len(b)))
		e.write(b)
	}
}

func (e *Encoder) writeBytes(b []byte) {
	e.writeInt32(int32(len(b)))
	e.write(b)
}

func (e *Encoder) writeCompactString(s string) {
	e.writeUVarInt(uint64(len(s)) + 1)
	e.writeString(s)
}

func (e *Encoder) writeVarString(s string) {
	e.writeVarInt(int64(len(s)))
	e.writeString(s)
}

func (e *Encoder) writeInt64(i int64) {
	binary.BigEndian.PutUint64(e.buffer[:8], uint64(i))
	_, err := e.writer.Write(e.buffer[:8])
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeInt32(i int32) {
	binary.BigEndian.PutUint32(e.buffer[:4], uint32(i))
	_, err := e.writer.Write(e.buffer[:4])
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeInt16(i int16) {
	binary.BigEndian.PutUint16(e.buffer[:2], uint16(i))
	_, err := e.writer.Write(e.buffer[:2])
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeInt8(i int8) {
	e.buffer[0] = byte(i)
	_, err := e.writer.Write(e.buffer[:1])
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeBool(b bool) {
	if b {
		e.writeInt8(1)
	} else {
		e.writeInt8(0)
	}
}

func (e *Encoder) writeCompactNullString(s string) {
	if s == "" {
		e.writeUVarInt(0)
	} else {
		e.writeUVarInt(uint64(len(s)) + 1)
		e.writeString(s)
	}
}

func (e *Encoder) writeNullString(s string) {
	if len(s) == 0 {
		e.writeInt16(-1)
	} else {
		e.writeInt16(int16(len(s)))
		e.writeString(s)
	}
}

func (e *Encoder) writeString(s string) {
	for len(s) != 0 {
		n := copy(e.buffer[:], s)
		_, err := e.writer.Write(e.buffer[:n])
		if err != nil {
			panic(err)
		}
		s = s[n:]
	}
}

func (e *Encoder) write(b []byte) {
	_, err := e.writer.Write(b)
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeVarInt(i int64) {
	n := binary.PutVarint(e.buffer[:], i)

	_, err := e.writer.Write(e.buffer[:n])
	if err != nil {
		panic(err)
	}

}

func (e *Encoder) writeUVarInt(i uint64) {
	b := e.buffer[:]
	n := 0

	for i >= 0x80 && n < len(b) {
		b[n] = byte(i) | 0x80
		i >>= 7
		n++
	}

	if n < len(b) {
		b[n] = byte(i)
		n++
	}

	_, err := e.writer.Write(b[:n])
	if err != nil {
		panic(err)
	}

}

func (e *Encoder) encodeTagBuffer(v reflect.Value) {
	e.writeUVarInt(0)
}
