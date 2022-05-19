package runtime

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/monitor"
	"net/http"
)

type HttpInfo struct {
	*openapi.Config
}

type HttpHandler struct {
	http *monitor.Http
	next http.Handler
	info *HttpInfo
}

func NewHttpHandler(http *monitor.Http, next http.Handler) *HttpHandler {
	return &HttpHandler{http: http, next: next}
}

func (h *HttpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewHttpContext(r.Context(), h.http)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}
