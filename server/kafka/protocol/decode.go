package protocol

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"io"
	"reflect"
)

type decodeFunc func(*Decoder, reflect.Value)

type Decoder struct {
	reader io.Reader
	buffer [8]byte
	err    error
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{reader: reader}
}

func newDecodeFunc(t reflect.Type, version int16, tag kafkaTag) decodeFunc {
	switch t.Kind() {
	case reflect.Struct:
		return newStructDecodeFunc(t, version, tag)
	case reflect.String:
		if tag.nullable && tag.compact > version {
			return (*Decoder).decodeString
		}
		return (*Decoder).decodeCompactString
	case reflect.Int32:
		return (*Decoder).decodeInt32
	case reflect.Int16:
		return (*Decoder).decodeInt16
	case reflect.Int8:
		return (*Decoder).decodeInt8
	case reflect.Bool:
		return (*Decoder).decodeBool
	case reflect.Map:
		if tag.protoType == "TAG_BUFFER" {
			return (*Decoder).decodeTagBuffer
		}
		panic("unsupported map: " + t.String())
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 { // []byte
			return (*Decoder).decodeCompactBytes
		}
		return newArayDecodeFunc(t, version)
	default:
		panic("unsupported type: " + t.String())
	}
}

func newArayDecodeFunc(t reflect.Type, version int16) decodeFunc {
	elemType := t.Elem()
	elemFunc := newDecodeFunc(elemType, version, kafkaTag{})

	return func(d *Decoder, v reflect.Value) { d.decodeCompactArray(v, elemFunc) }
}

func newStructDecodeFunc(t reflect.Type, version int16, tag kafkaTag) decodeFunc {
	type field struct {
		index  int
		decode decodeFunc
	}
	fields := make([]*field, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := getTag(f)
		if !tag.isValid(version) {
			continue
		}
		fields = append(fields, &field{i, newDecodeFunc(f.Type, version, tag)})
	}

	return func(d *Decoder, v reflect.Value) {
		for _, f := range fields {
			f.decode(d, v.Field(f.index))
		}
	}
}

func (d *Decoder) decodeCompactBytes(v reflect.Value) {
	b := d.readCompactBytes()
	v.Set(reflect.ValueOf(b))
}

func (d *Decoder) decodeCompactString(v reflect.Value) {
	s := d.readCompactString()
	v.Set(reflect.ValueOf(s))
}

func (d *Decoder) decodeString(v reflect.Value) {
	s := d.readString()
	v.Set(reflect.ValueOf(s))
}

func (d *Decoder) decodeBool(v reflect.Value) {
	b := d.readBool()
	v.Set(reflect.ValueOf(b))
}

func (d *Decoder) decodeInt8(v reflect.Value) {
	i := d.readInt8()
	v.Set(reflect.ValueOf(i).Convert(v.Type()))
}

func (d *Decoder) decodeInt16(v reflect.Value) {
	i := d.readInt16()
	v.Set(reflect.ValueOf(i).Convert(v.Type()))
}

func (d *Decoder) decodeInt32(v reflect.Value) {
	i := d.readInt32()
	v.Set(reflect.ValueOf(i))
}

func (d *Decoder) decodeTagBuffer(v reflect.Value) {
	m := d.readTagFields()
	v.Set(reflect.ValueOf(m))
}

func (d *Decoder) readCompactString() string {
	if n := d.readUvarint(); n == 0 {
		return ""
	} else {
		b := make([]byte, n-1)
		if d.readFull(b) {
			return string(b)
		}
	}
	return ""
}

func (d *Decoder) readInt32() int32 {
	if d.readFull(d.buffer[:4]) {
		i := binary.BigEndian.Uint32(d.buffer[:4])
		return int32(i)
	}
	return 0
}

func (d *Decoder) readInt16() int16 {
	if d.readFull(d.buffer[:2]) {
		i := binary.BigEndian.Uint16(d.buffer[:2])
		return int16(i)
	} else {
		return 0
	}
}

func (d *Decoder) readInt8() int8 {
	return int8(d.readByte())
}

func (d *Decoder) readBool() bool {
	return d.readInt8() != 0
}

func (d *Decoder) readString() string {
	if n := d.readInt16(); n == 0 {
		return ""
	} else {
		b := make([]byte, n)
		if d.readFull(b) {
			return string(b[0:])
		}
	}
	return ""
}

func (d *Decoder) readUvarint() uint64 {
	var x uint64
	var s uint
	for i := 0; ; i++ {
		b := d.readByte()
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				d.err = errors.New("kafka: varint overflows a 64-bit integer")
				return x
			}
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}

func (d *Decoder) readCompactBytes() []byte {
	if n := d.readUvarint(); n < 1 {
		return nil
	} else {
		b := make([]byte, n)
		if d.readFull(b) {
			return b
		} else {
			return nil
		}
	}
}

func (d *Decoder) decodeCompactArray(v reflect.Value, decodeElem decodeFunc) {
	if n := d.readUvarint(); n < 1 {
		a := reflect.MakeSlice(v.Type(), 0, 0)
		v.Set(a)
	} else {
		l := int(n - 1)
		a := reflect.MakeSlice(v.Type(), l, l)
		for i := 0; i < l; i++ {
			decodeElem(d, a.Index(i))
		}
		v.Set(a)
	}
}

func (d *Decoder) readByte() byte {
	if d.err != nil {
		return 0
	}
	if _, err := d.reader.Read(d.buffer[:1]); err != nil {
		d.err = err
		return 0
	}
	return d.buffer[0]
}

//func (d *Decoder) read(b []byte) bool {
//	if d.err != nil {
//		return false
//	}
//	if _, err := d.reader.Read(b); err != nil {
//		d.err = err
//		return false
//	}
//	return true
//}

func (d *Decoder) readFull(b []byte) bool {
	if d.err != nil {
		return false
	}

	if _, err := io.ReadFull(d.reader, b); err != nil {
		d.err = err
		return false
	}
	return true
}

func (d *Decoder) readTagFields() map[int64]string {
	if tagCount := d.readUvarint(); tagCount == 0 {
		return nil
	} else {
		tagFields := make(map[int64]string)
		for i := 0; i < int(tagCount); i++ {
			tagId := d.readUvarint()
			if size := d.readUvarint(); size == 0 {
				d.err = errors.New("tag size zero")
				return tagFields
			} else {
				tag := make([]byte, int(size))
				d.readFull(tag)
				tagFields[int64(tagId)] = string(tag)
			}
		}
		return tagFields
	}
}
