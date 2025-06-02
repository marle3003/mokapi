package mqtt

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Encoder struct {
	writer io.Writer
	buffer [32]byte
}

func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{writer: writer}
}

func (e *Encoder) writeByte(b byte) {
	e.buffer[0] = b
	_, err := e.writer.Write(e.buffer[:1])
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

func (e *Encoder) Write(b []byte) {
	_, err := e.writer.Write(b)
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeString(s string) {
	e.writeInt16(int16(len(s)))
	for len(s) != 0 {
		n := copy(e.buffer[:], s)
		_, err := e.writer.Write(e.buffer[:n])
		if err != nil {
			panic(err)
		}
		s = s[n:]
	}
}

func encodeBool(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func writeRemainingLength(w io.Writer, length int) error {
	if length < 0 || length > 268435455 {
		return fmt.Errorf("remaining length out of range: %d", length)
	}

	for {
		encodedByte := byte(length % 128)
		length /= 128
		// if there are more digits to encode, set the top bit of this byte
		if length > 0 {
			encodedByte |= 0x80
		}
		if _, err := w.Write([]byte{encodedByte}); err != nil {
			return err
		}
		if length == 0 {
			break
		}
	}
	return nil
}
