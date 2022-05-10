package kafka

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"testing"
)

func TestPageBuffer_ReadWrite(t *testing.T) {
	pb := newPageBuffer()
	data := bytes.Repeat([]byte("foobar"), 100)
	n, err := pb.Write(data)
	require.NoError(t, err)
	require.Len(t, data, n)
	require.Len(t, data, pb.length)

	f := pb.fragment(30, 36)
	require.Equal(t, 6, f.Len())

	read, err := ioutil.ReadAll(f)
	require.Equal(t, "foobar", string(read))

	page := pb.pages[0]
	pb.unref()
	require.Len(t, data, page.length)
	require.Equal(t, 6, f.Len())

	f.unref()
	require.Equal(t, 0, page.length)
	require.Equal(t, 0, f.Len())
}

func TestPage_WriteAt(t *testing.T) {
	p := newPage(0)
	b := []byte("foobar")
	n, err := p.WriteAt(b, 10)
	require.NoError(t, err)
	require.Len(t, b, n)
	require.Equal(t, len(b)+10, p.length)
}

func TestFragment_ReadWriteOutOfRange(t *testing.T) {
	pb := newPageBuffer()
	f := pb.fragment(0, 10)
	b := make([]byte, 10)
	n, err := f.Read(b)
	require.Equal(t, io.EOF, err)
	require.Equal(t, 0, n)
}

func TestPage_ReadOutOfRange(t *testing.T) {
	p := newPage(0)
	b := make([]byte, 10)
	n, err := p.ReadAt(b, 10)
	require.NoError(t, err)
	require.Equal(t, 0, n)

	defer func() {
		err := recover()
		require.NotNil(t, err, "should panic")
	}()
	_, _ = p.ReadAt(b, pageSize+1)
}

func TestWritePageOutOfRange(t *testing.T) {
	p := newPage(0)
	b := []byte("foobar")
	n, err := p.WriteAt(b, pageSize)
	require.NoError(t, err)
	require.Equal(t, 0, n)

	defer func() {
		err := recover()
		require.NotNil(t, err, "should panic")
	}()
	_, _ = p.WriteAt(b, pageSize+1)
}
