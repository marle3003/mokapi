package kafka

import (
	"io"
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
	refs   refCounter
}

type fragment struct {
	pages  []*page
	offset int
	cursor int
	length int
}

func newPage(offset int) *page {
	p := pagePool.Get().(*page)
	p.offset = offset
	p.refs.inc()
	return p
}

func (p *page) unref() {
	if p.refs.dec() {
		p.offset = 0
		p.length = 0
		pagePool.Put(p)
	}
}

func (p *page) ReadAt(b []byte, offset int) (int, error) {
	if offset -= p.offset; offset < 0 || offset > pageSize {
		panic("offset out of range")
	}

	if offset > p.length {
		return 0, nil
	}

	n := copy(b, p.buffer[offset:p.length])
	return n, nil
}

func (p *page) Write(b []byte) (n int, err error) {
	n = copy(p.buffer[p.length:], b)
	p.length += n
	return
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

func (f *fragment) Read(b []byte) (int, error) {
	if end := f.offset + f.length; f.cursor >= end {
		return 0, io.EOF
	}
	if len(b) > f.length {
		b = b[:f.length]
	}
	read := 0
	for _, p := range f.pages {
		n, _ := p.ReadAt(b, f.cursor)
		b = b[n:]
		f.cursor += n
		read += n
	}
	if read == 0 {
		return 0, io.EOF
	}
	return read, nil
}

func (f *fragment) Size() int {
	return f.length
}

func (f *fragment) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		f.cursor += int(offset)
	case io.SeekStart:
		f.cursor = f.offset + int(offset)
	case io.SeekEnd:
		f.cursor = f.offset + f.length + int(offset)
	}
	return int64(f.cursor), nil
}

func (f *fragment) ref() {
	for _, p := range f.pages {
		p.refs.inc()
	}
}

func (f *fragment) unref() {
	for _, p := range f.pages {
		p.unref()
	}
	f.pages = nil
	f.length = 0
	f.cursor = 0
	f.offset = 0
}

func (f *fragment) Close() error {
	f.unref()
	return nil
}
