package dynamictest

import (
	"mokapi/config/dynamic"
	"net/url"
	"time"
)

func NewConfigInfo() dynamic.ConfigInfo {
	u, _ := url.Parse("file://foo.yml")

	return dynamic.ConfigInfo{
		Provider: "test",
		Url:      u,
		Time:     time.Now(),
	}
}
