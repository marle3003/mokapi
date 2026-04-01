package lib

import (
	"net/http"
	"strings"
)

func GetUrl(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}
	var sb strings.Builder
	if r.TLS != nil {
		sb.WriteString("https://")
	} else {
		sb.WriteString("http://")
	}
	if r.Host != "" {
		sb.WriteString(r.Host)
	} else {
		sb.WriteString("localhost")
	}
	sb.WriteString(r.URL.String())
	return sb.String()
}
