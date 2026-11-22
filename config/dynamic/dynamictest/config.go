package dynamictest

import (
	"mokapi/config/dynamic"
	"net/url"
	"time"
)

type ConfigInfoOption func(info *dynamic.ConfigInfo)

func NewConfigInfo(opts ...ConfigInfoOption) dynamic.ConfigInfo {
	u, _ := url.Parse("file://foo.yml")

	ci := dynamic.ConfigInfo{
		Provider: "test",
		Url:      u,
		Time:     time.Now(),
	}

	for _, opt := range opts {
		opt(&ci)
	}
	return ci
}

func WithUrl(s string) ConfigInfoOption {
	return func(info *dynamic.ConfigInfo) {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		info.Url = u
	}
}

type ConfigOption func(c *dynamic.Config)

// NewConfig returns a config that already holds the given parsed data.
func NewConfig(data any, opts ...ConfigOption) *dynamic.Config {
	c := &dynamic.Config{}
	c.SetData(data)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithInfo(info dynamic.ConfigInfo) ConfigOption {
	return func(c *dynamic.Config) {
		c.Info = info
	}
}

func WithRaw(raw []byte) ConfigOption {
	return func(c *dynamic.Config) {
		c.Raw = raw
	}
}
