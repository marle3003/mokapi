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
	if strings.HasPrefix(r.Proto, "HTTPS") {
		sb.WriteString("https://")
	} else {
		sb.WriteString("http://")
	}
	sb.WriteString(r.Host)
	sb.WriteString(r.URL.String())
	return sb.String()
}
