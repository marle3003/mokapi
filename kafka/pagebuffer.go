package kafka

import (
	"bytes"
	"io"
)

type Bytes interface {
	io.ReadCloser
	io.Seeker
	Size() int
}

type bytesReader struct {
	bytes.Reader
}

func NewBytes(b []byte) Bytes {
	r := new(bytesReader)
	r.Reset(b)
	return r
}

func BytesToString(bytes Bytes) string {
	return string(Read(bytes))
}

func Read(bytes Bytes) []byte {
	if bytes == nil {
		return nil
	}
	_, _ = bytes.Seek(0, io.SeekStart)
	b := make([]byte, bytes.Size())
	_, _ = bytes.Read(b)
	return b
}

func (b *bytesReader) Close() error { return nil }

func (b *bytesReader) Size() int { return int(b.Reader.Size()) }
