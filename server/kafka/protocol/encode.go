package protocol

import (
	"encoding/binary"
	"io"
	"reflect"
)

type encodeFunc func(*Encoder, reflect.Value)

type BufferWriter interface {
	io.Writer
	WriteAt([]byte, int)
	WriteSizeAt(size int, offset int)
	Size() int
	Scan(begin, end int, f func([]byte) bool)
}

type WriterTo interface {
	WriteTo(e *Encoder)
}

var (
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

func newStructEncodeFunc(t reflect.Type, version int16, _ kafkaTag) encodeFunc {
	type field struct {
		index  int
		name   string
		encode encodeFunc
	}
	fields := make([]*field, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := getTag(f)
		if !tag.isValid(version) {
			continue
		}
		fields = append(fields, &field{i, f.Name, newEncodeFunc(f.Type, version, tag)})
	}

	return func(e *Encoder, v reflect.Value) {
		for _, f := range fields {
			fv := v.Field(f.index)
			f.encode(e, fv)
		}
	}
}

func (e *Encoder) encodeCompactArray(v reflect.Value, encodeElem encodeFunc) {
	len := v.Len()
	e.writeUVarInt(uint64(len + 1))

	for i := 0; i < len; i++ {
		item := v.Index(i)
		encodeElem(e, item)
	}
}

func (e *Encoder) encodeArray(v reflect.Value, encodeElem encodeFunc) {
	len := v.Len()
	e.writeInt32(int32(len))

	for i := 0; i < len; i++ {
		item := v.Index(i)
		encodeElem(e, item)
	}
}

func (e *Encoder) encodeCompactBytes(v reflect.Value) {
	b := v.Bytes()
	e.writeCompactBytes(b)
}

func (e *Encoder) encodeCompactNullBytes(v reflect.Value) {
	b := v.Bytes()
	e.writeCompactNullBytes(b)
}

func (e *Encoder) encodeNullBytes(v reflect.Value) {
	b := v.Bytes()
	e.writeNullBytes(b)
}

func (e *Encoder) encodeBytes(v reflect.Value) {
	b := v.Bytes()
	e.writeBytes(b)
}

func (e *Encoder) encodeCompactString(v reflect.Value) {
	s := v.String()
	e.writeCompactString(s)
}

func (e *Encoder) encodeCompactNullString(v reflect.Value) {
	s := v.String()
	e.writeCompactNullString(s)
}

func (e *Encoder) encodeNullString(v reflect.Value) {
	s := v.String()
	e.writeNullString(s)
}

func (e *Encoder) encodeString(v reflect.Value) {
	s := v.String()
	e.writeNullString(s)
}

func (e *Encoder) encodeInt8(v reflect.Value) {
	i := v.Int()
	e.writeInt8(int8(i))
}

func (e *Encoder) encodeInt16(v reflect.Value) {
	i := v.Int()
	e.writeInt16(int16(i))
}

func (e *Encoder) encodeBool(v reflect.Value) {
	b := v.Bool()
	e.writeBool(b)
}

func (e *Encoder) encodeInt32(v reflect.Value) {
	i := v.Int()
	e.writeInt32(int32(i))
}

func (e *Encoder) encodeInt64(v reflect.Value) {
	i := v.Int()
	e.writeInt64(i)
}

func (e *Encoder) writeCompactBytes(b []byte) {
	len := len(b)
	e.writeUVarInt(uint64(len) + 1)
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

func (e *Encoder) encodeTagBuffer(_ reflect.Value) {
	e.writeUVarInt(0)
}
