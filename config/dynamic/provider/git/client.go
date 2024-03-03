package git

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type transport struct {
	trans map[string]http.RoundTripper
	m     sync.RWMutex
}

func newTransport() *transport {
	return &transport{}
}

func (t *transport) Add(prefixUrl string, transport http.RoundTripper) {
	t.m.Lock()
	defer t.m.Unlock()

	if t.trans == nil {
		t.trans = make(map[string]http.RoundTripper)
	}
	t.trans[prefixUrl] = transport
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.trans == nil {
		return http.DefaultTransport.RoundTrip(r)
	}
	t.m.RLock()

	s := r.URL.String()
	for k, v := range t.trans {
		if strings.HasPrefix(s, k) {
			t.m.RUnlock()
			return v.RoundTrip(r)
		}
	}
	t.m.Unlock()
	return http.DefaultTransport.RoundTrip(r)
}

func (t *transport) addAuth(r *repository) error {
	if r.auth == nil {
		return nil
	}
	if r.auth.GitHub != nil {
		return addGitHubAuth(t, r)
	}
	return fmt.Errorf("not supported auth: %v", r.auth)
}

func newClient(t *transport) *http.Client {
	return &http.Client{Transport: t}
}
