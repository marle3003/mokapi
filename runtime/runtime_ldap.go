package runtime

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/runtime/monitor"
	"net/http"
)

type LdapInfo struct {
	*openapi.Config
}

type LdapHandler struct {
	ldap *monitor.Ldap
	next http.Handler
}

func NewLdapHandler(http *monitor.Http, next http.Handler) *HttpHandler {
	return &HttpHandler{http: http, next: next}
}

func (h *LdapHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := monitor.NewLdapContext(r.Context(), h.ldap)
	h.ldap.RequestCounter.Add(1)

	h.next.ServeHTTP(rw, r.WithContext(ctx))

}
