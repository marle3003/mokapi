package dynamictest

import (
	"errors"
	"mokapi/config/dynamic"
	"net/url"
)

var NotFound = errors.New("TestReader: config not found")

type Reader struct {
	Data map[string]*dynamic.Config
}

func (r *Reader) Read(u *url.URL, v any) (*dynamic.Config, error) {
	if r.Data == nil {
		return nil, NotFound
	}
	if c, ok := r.Data[u.String()]; ok {
		if p, isParser := c.Data.(dynamic.Parser); isParser {
			if err := p.Parse(c, r); err != nil {
				return nil, err
			}
			return c, nil
		}

		return c, nil
	}
	return nil, NotFound
}

type ReaderFunc func(u *url.URL, v any) (*dynamic.Config, error)

func (f ReaderFunc) Read(u *url.URL, v any) (*dynamic.Config, error) {
	return f(u, v)
}
