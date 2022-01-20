package runtime

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/monitor"
	"net/http"
)

type HttpInfo struct {
	*openapi.Config
	Status string `json:"status"`
}

type HttpHandler struct {
	http *monitor.Http
	next http.Handler
}

func NewHttpHandler(http *monitor.Http, next http.Handler) *HttpHandler {
	return &HttpHandler{http: http, next: next}
}

func (h *HttpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewHttpContext(r.Context(), h.http)
	h.http.RequestCounter.Add(1)

	h.next.ServeHTTP(rw, r.WithContext(ctx))

}
