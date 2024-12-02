package dynamictest

import (
	"errors"
	"mokapi/config/dynamic"
	"net/url"
)

var NotFound = errors.New("config not found")

type Reader struct {
	Data map[*url.URL]*dynamic.Config
}

func (r *Reader) Read(u *url.URL, v any) (*dynamic.Config, error) {
	if r.Data == nil {
		return nil, NotFound
	}
	if c, ok := r.Data[u]; ok {
		return c, nil
	}
	return nil, NotFound
}

type ReaderFunc func(u *url.URL, v any) (*dynamic.Config, error)

func (f ReaderFunc) Read(u *url.URL, v any) (*dynamic.Config, error) {
	return f(u, v)
}
