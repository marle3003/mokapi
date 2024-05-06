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
