package common

import (
	"mokapi/safe"
	"net/url"
)

type Config struct {
	Url          *url.URL
	Data         []byte
	ProviderName string
}

type Provider interface {
	Read(u *url.URL) (*Config, error)
	Start(chan *Config, *safe.Pool) error
}
