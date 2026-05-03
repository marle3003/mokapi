package mqtt

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Decoder struct {
	reader          io.Reader
	buffer          [8]byte
	err             error
	leftSize        int
	protocolVersion byte
}

func NewDecoder(reader io.Reader, size int, protocolVersion byte) *Decoder {
	return &Decoder{reader: reader, leftSize: size, protocolVersion: protocolVersion}
}

func (d *Decoder) IsV5() bool {
	return d.protocolVersion == 5
}

func (d *Decoder) ReadByte() byte {
	if d.err != nil {
		return 0
	}
	n, err := d.reader.Read(d.buffer[:1])
	if err != nil {
		d.err = err
		return 0
	}
	d.leftSize -= n
	return d.buffer[0]
}

func (d *Decoder) ReadString() string {
	if n := d.ReadInt16(); n == 0 {
		return ""
	} else {
		b := make([]byte, n)
		if d.readFull(b) {
			return string(b[0:])
		}
	}
	return ""
}

func (d *Decoder) ReadBytes() []byte {
	if n := d.ReadInt16(); n < 0 {
		return nil
	} else {
		b := make([]byte, n)
		if d.readFull(b) {
			return b[0:]
		}
	}
	return nil
}

func (d *Decoder) ReadInt16() int16 {
	if d.readFull(d.buffer[:2]) {
		i := binary.BigEndian.Uint16(d.buffer[:2])
		return int16(i)
	} else {
		return 0
	}
}

func (d *Decoder) ReadUInt16() uint16 {
	if d.readFull(d.buffer[:2]) {
		i := binary.BigEndian.Uint16(d.buffer[:2])
		return uint16(i)
	} else {
		return 0
	}
}

func (d *Decoder) ReadInt32() int32 {
	if d.readFull(d.buffer[:4]) {
		i := binary.BigEndian.Uint16(d.buffer[:4])
		return int32(i)
	} else {
		return 0
	}
}

func (d *Decoder) readFull(b []byte) bool {
	if d.err != nil {
		return false
	}

	n, err := io.ReadFull(d.reader, b)
	if err != nil {
		d.err = err
		return false
	}
	d.leftSize -= n

	return true
}

func (d *Decoder) readRemainingLength() int {
	var (
		multiplier = 1
		value      = 0
		digit      byte
	)

	for i := 0; i < 4; i++ { // max 4 bytes
		n, err := d.reader.Read(d.buffer[:1])
		if err != nil {
			d.err = err
			return 0
		}
		d.leftSize -= n
		digit = d.buffer[0]
		value += int(digit&127) * multiplier
		if digit&128 == 0 {
			return value
		}
		multiplier *= 128
	}

	d.err = fmt.Errorf("malformed Remaining Length: exceeds 4 bytes")
	return 0
}

func (d *Decoder) ReadVariableInt() int {
	var multiplier = 1
	var value = 0

	for {
		b := d.ReadByte()

		value += int(b&127) * multiplier

		if multiplier > 128*128*128 {
			panic("Malformed Variable Byte Integer")
		}
		multiplier *= 128

		if b&128 == 0 {
			break
		}
	}

	return value
}
