package lib_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/lib"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetUrl(t *testing.T) {
	var s *httptest.Server
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := lib.GetUrl(r)
		require.Equal(t, fmt.Sprintf("%s/foo", s.URL), u)
		require.True(t, strings.HasPrefix(u, "http"))
	}))
	defer s.Close()

	_, err := s.Client().Get(s.URL + "/foo")
	require.NoError(t, err)

	s = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := lib.GetUrl(r)
		require.Equal(t, fmt.Sprintf("%s/foo", s.URL), u)
		require.True(t, strings.HasPrefix(u, "https"))
	}))
	defer s.Close()

	_, err = s.Client().Get(s.URL + "/foo")
	require.NoError(t, err)
}
