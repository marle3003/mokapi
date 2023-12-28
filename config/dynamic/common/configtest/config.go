package configtest

import (
	"mokapi/config/dynamic/common"
	"net/url"
)

func NewConfigInfo() common.ConfigInfo {
	u, _ := url.Parse("file://foo.yml")

	return common.ConfigInfo{
		Provider: "test",
		Url:      u,
	}
}
