package lib

import (
	"net"
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

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
