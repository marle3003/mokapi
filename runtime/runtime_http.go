package runtime

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/logs"
	"mokapi/runtime/monitor"
	"net/http"
	"time"
)

type HttpInfo struct {
	*openapi.Config
	LastRequest time.Time
	Requests    int64
	Errors      int64
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

	log := logs.NewHttpLog(r.Method, r.URL.String())
	h.http.AppendHttp(log)
	ctx = logs.NewHttpLogContext(ctx, log)

	h.next.ServeHTTP(rw, r.WithContext(ctx))
}
