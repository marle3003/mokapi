package lib_test

import (
	"fmt"
	"mokapi/lib"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestClientIP(t *testing.T) {
	var s *httptest.Server
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := lib.GetUrl(r)
		require.Equal(t, fmt.Sprintf("%s/foo", s.URL), u)
		require.True(t, strings.HasPrefix(u, "http"))
	}))
	defer s.Close()

	_, err := s.Client().Get(s.URL + "/foo")
	require.NoError(t, err)

	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := lib.ClientIP(r)
		require.Equal(t, "127.0.0.1", ip)
	}))
	defer s.Close()

	_, err = s.Client().Get(s.URL + "/foo")
	require.NoError(t, err)
}
