package protocol

import (
	"encoding/binary"
	"sync"
)

const pageSize = 65536

var (
	pagePool = sync.Pool{New: func() interface{} { return new(page) }}
)

type page struct {
	offset int // absolute offset
	buffer [pageSize]byte
	length int // length of this page
}

func newPage(offset int) *page {
	p, _ := pagePool.Get().(*page)
	p.offset = offset
	return p
}

func (p *page) unref() {
	p.offset = 0
	p.length = 0
	pagePool.Put(p)
}

func (p *page) Write(b []byte) (n int, err error) {
	n = copy(p.buffer[p.length:], b)
	p.length += n
	return
}

func (p *page) WriteSizeAt(size int, offset int) {
	binary.BigEndian.PutUint32(p.buffer[offset:offset+4], uint32(size))
}

func (p *page) WriteAt(b []byte, offset int) (int, error) {
	if offset -= p.offset; offset < 0 || offset > pageSize {
		panic("offset out of range")
	}
	n := copy(p.buffer[offset:], b)
	if end := offset + n; end > p.length {
		p.length += end
	}
	return n, nil
}

func (p *page) Size() int {
	return p.length
}

func (p *page) slice(begin, end int) []byte {
	i, j := begin-p.offset, end-p.offset
	if i < 0 {
		i = 0
	}
	if j > len(p.buffer) {
		j = len(p.buffer)
	}
	return p.buffer[i:j]
}
