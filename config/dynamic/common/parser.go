package common

import "net/url"

type Parser interface {
	Parse(file *File, reader Reader) error
}

type Reader interface {
	Read(u *url.URL, opts ...FileOptions) (*File, error)
}
