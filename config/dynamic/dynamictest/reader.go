package dynamictest

import (
	"errors"
	"mokapi/config/dynamic"
	"net/url"
)

var NotFound = errors.New("TestReader: config not found")

type Reader struct {
	Data map[string]*dynamic.Config

	parsed map[string]bool
}

func (r *Reader) Read(u *url.URL, v any) (*dynamic.Config, error) {
	if r.Data == nil {
		return nil, NotFound
	}
	if c, ok := r.Data[u.String()]; ok {
		if c.Data == nil {
			c.Data = v
			err := dynamic.Parse(c, r)
			return c, err
		}

		if p, isParser := c.Data.(dynamic.Parser); isParser {
			if _, alreadyParsed := r.parsed[u.String()]; alreadyParsed {
				return c, nil
			}
			if err := p.Parse(c, r); err != nil {
				return nil, err
			}

			if r.parsed == nil {
				r.parsed = make(map[string]bool)
			}

			r.parsed[u.String()] = true
		}

		return c, nil
	}
	return nil, NotFound
}

type ReaderFunc func(u *url.URL, v any) (*dynamic.Config, error)

func (f ReaderFunc) Read(u *url.URL, v any) (*dynamic.Config, error) {
	return f(u, v)
}
