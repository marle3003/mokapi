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
	return t.transportFor(r.URL.String()).RoundTrip(r)
}

// transportFor returns the transport registered for the first matching URL prefix,
// requests to other repositories go to http.DefaultTransport.
func (t *transport) transportFor(url string) http.RoundTripper {
	t.m.RLock()
	defer t.m.RUnlock()

	for prefix, rt := range t.trans {
		if strings.HasPrefix(url, prefix) {
			return rt
		}
	}
	return http.DefaultTransport
}

func (t *transport) addAuth(r *repository) error {
	if r.config.Auth == nil {
		return nil
	}
	if r.config.Auth.GitHub != nil {
		return addGitHubAuth(t, r)
	}
	return fmt.Errorf("not supported auth: %v", r.config.Auth)
}

func newClient(t *transport) *http.Client {
	return &http.Client{Transport: t}
}
