package protocol

import (
	"bytes"
	"io"
	"io/ioutil"
	"mokapi/test"
	"testing"
)

func TestPageBuffer_ReadWrite(t *testing.T) {
	pb := newPageBuffer()
	data := bytes.Repeat([]byte("foobar"), 100)
	n, err := pb.Write(data)
	test.Ok(t, err)
	test.Equals(t, len(data), n)
	test.Equals(t, len(data), pb.length)

	f := pb.fragment(30, 36)
	test.Equals(t, 6, f.Len())

	read, err := ioutil.ReadAll(f)
	test.Equals(t, "foobar", string(read))

	page := pb.pages[0]
	pb.unref()
	test.Equals(t, len(data), page.length)
	test.Equals(t, 6, f.Len())

	f.unref()
	test.Equals(t, 0, page.length)
	test.Equals(t, 0, f.Len())
}

func TestPage_WriteAt(t *testing.T) {
	p := newPage(0)
	b := []byte("foobar")
	n, err := p.WriteAt(b, 10)
	test.Ok(t, err)
	test.Equals(t, len(b), n)
	test.Equals(t, len(b)+10, p.length)
}

func TestFragment_ReadWriteOutOfRange(t *testing.T) {
	pb := newPageBuffer()
	f := pb.fragment(0, 10)
	b := make([]byte, 10)
	n, err := f.Read(b)
	test.Equals(t, io.EOF, err)
	test.Equals(t, 0, n)
}

func TestPage_ReadOutOfRange(t *testing.T) {
	p := newPage(0)
	b := make([]byte, 10)
	n, err := p.ReadAt(b, 10)
	test.Ok(t, err)
	test.Equals(t, 0, n)

	defer func() {
		err := recover()
		test.Assert(t, err != nil, "should panic")
	}()
	_, _ = p.ReadAt(b, pageSize+1)
}

func TestWritePageOutOfRange(t *testing.T) {
	p := newPage(0)
	b := []byte("foobar")
	n, err := p.WriteAt(b, pageSize)
	test.Ok(t, err)
	test.Equals(t, 0, n)

	defer func() {
		err := recover()
		test.Assert(t, err != nil, "should panic")
	}()
	_, _ = p.WriteAt(b, pageSize+1)
}
