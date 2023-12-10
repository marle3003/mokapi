package common

import (
	"net/url"
	"strings"
)

type ConfigInfo struct {
	Provider string
	Url      *url.URL
	inner    *ConfigInfo
}

func (ci *ConfigInfo) Path() string {
	if len(ci.Url.Opaque) > 0 {
		return ci.Url.Opaque
	}
	u := ci.Url
	path, _ := url.PathUnescape(ci.Url.Path)
	query, _ := url.QueryUnescape(ci.Url.RawQuery)
	var sb strings.Builder
	if len(u.Scheme) > 0 {
		sb.WriteString(u.Scheme + ":")
	}
	if len(u.Scheme) > 0 || len(u.Host) > 0 {
		sb.WriteString("//")
	}
	if len(u.Host) > 0 {
		sb.WriteString(u.Host)
	}
	sb.WriteString(path)
	if len(query) > 0 {
		sb.WriteString("?" + query)
	}
	return sb.String()
}

func (ci *ConfigInfo) Inner() *ConfigInfo {
	return ci.inner
}
