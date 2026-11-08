package server

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfigKey(t *testing.T) {
	testcases := []struct {
		name string
		url  string
		key  string
	}{
		{name: "file with empty host", url: "file:///foo/bar.yml", key: "file:/foo/bar.yml"},
		{name: "file without host", url: "file:/foo/bar.yml", key: "file:/foo/bar.yml"},
		{name: "http", url: "http://localhost:8080/foo?bar=1", key: "http://localhost:8080/foo?bar=1"},
		{name: "https with user", url: "https://user:pw@example.com/foo.yml", key: "https://user:pw@example.com/foo.yml"},
		{name: "git", url: "git+https://github.com/foo/bar?file=a.yml", key: "git+https://github.com/foo/bar?file=a.yml"},
		{name: "npm", url: "npm://foo?file=a.yml", key: "npm://foo?file=a.yml"},
		{name: "relative", url: "foo/bar.yml", key: "foo/bar.yml"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			require.NoError(t, err)

			require.Equal(t, tc.key, getConfigKey(u))
			require.Equal(t, tc.url, u.String(), "URL must stay unchanged")
		})
	}
}
