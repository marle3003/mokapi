package buffer

import (
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
)

var (
	pageBufferPool = sync.Pool{New: func() interface{} { return new(pageBuffer) }}
)

type refCounter uint32

type pageBuffer struct {
	pages  []*page
	length int
	cursor int
	refs   refCounter
}

type Writer interface {
}

type Buffer interface {
	io.Writer
	WriteTo(io.Writer) (int, error)
	WriteAt([]byte, int)
	WriteSizeAt(size int, offset int)
	Scan(begin, end int, f func([]byte) bool)
	Size() int
	Slice(begin, end int) Fragment
	Unref()
}

type Fragment interface {
	io.ReadCloser
	io.Seeker
	Size() int
}

func (r *refCounter) inc() {
	atomic.AddUint32((*uint32)(r), 1)
}

func (r *refCounter) dec() bool {
	i := atomic.AddUint32((*uint32)(r), ^uint32(0))
	return i == 0
}

func NewPageBuffer() Buffer {
	pb := pageBufferPool.Get().(*pageBuffer)
	pb.refs.inc()
	return pb
}

func (pb *pageBuffer) Unref() {
	if pb.refs.dec() {
		for _, p := range pb.pages {
			p.unref()
		}
		pb.length = 0
		pb.cursor = 0
		pb.pages = nil
		pageBufferPool.Put(pb)
	}
}

func (pb *pageBuffer) Write(b []byte) (n int, err error) {
	n = len(b)
	if len(pb.pages) == 0 {
		pb.addPage()
	}

	for len(b) != 0 {
		tail := pb.pages[len(pb.pages)-1]
		available := pageSize - tail.Size()

		if len(b) <= available {
			tail.Write(b)
			pb.length += len(b)
			break
		}

		tail.Write(b[:available])
		b = b[available:]
		pb.length += available
		pb.addPage()
	}

	return
}

func (pb *pageBuffer) WriteSizeAt(size int, offset int) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(size))
	pb.WriteAt(b[:], offset)
}

func (pb *pageBuffer) WriteAt(b []byte, offset int) {
	for _, p := range pb.slice(offset, offset+len(b)) {
		n, _ := p.WriteAt(b, offset)
		b = b[n:]
		offset += n
	}
}

func (pb *pageBuffer) Size() int {
	return pb.length
}

func (pb *pageBuffer) addPage() {
	p := newPage(pb.length)
	pb.pages = append(pb.pages, p)
}

func (pb *pageBuffer) slice(begin, end int) []*page {
	i := begin / pageSize
	j := end / pageSize
	if j < len(pb.pages) {
		j++
	}
	return pb.pages[i:j]
}

func (pb *pageBuffer) Scan(begin, end int, f func([]byte) bool) {
	for _, p := range pb.slice(begin, end) {
		if !f(p.slice(begin, end)) {
			return
		}
	}
}

func (pb *pageBuffer) WriteTo(w io.Writer) (written int, err error) {
	pb.Scan(pb.cursor, pb.length, func(b []byte) bool {
		var n int
		n, err = w.Write(b)
		written += n
		return err == nil
	})
	pb.cursor += written
	return
}

func (pb *pageBuffer) Slice(begin, end int) Fragment {
	pages := pb.slice(begin, end)
	f := &fragment{
		pages:  pages,
		offset: begin,
		cursor: begin,
		length: end - begin,
	}
	f.ref()
	return f
}
