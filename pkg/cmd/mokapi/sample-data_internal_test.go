package mokapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFileOrUrl(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		scheme   string
		opaque   string
		path     string
		host     string
		fragment string
	}{
		{name: "drive letter with backslash", input: `C:\x.yaml`, scheme: "file", opaque: `C:\x.yaml`},
		{name: "drive letter with fragment", input: `C:/x.yaml#/a`, scheme: "file", opaque: `C:/x.yaml`, fragment: "/a"},
		{name: "relative path", input: "./x.json", scheme: "file", path: "./x.json"},
		{name: "absolute path with fragment", input: "/abs/x.yaml#/a", scheme: "file", path: "/abs/x.yaml", fragment: "/a"},
		{name: "UNC path", input: `\\server\share\x.yaml`, scheme: "file", path: `\\server\share\x.yaml`},
		{name: "https url with fragment", input: "https://example.com/x.yaml#/a", scheme: "https", path: "/x.yaml", host: "example.com", fragment: "/a"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := parseFileOrUrl(tc.input)

			require.NoError(t, err)
			require.Equal(t, tc.scheme, u.Scheme)
			require.Equal(t, tc.opaque, u.Opaque)
			require.Equal(t, tc.path, u.Path)
			require.Equal(t, tc.host, u.Host)
			require.Equal(t, tc.fragment, u.Fragment)
		})
	}

	t.Run("invalid url", func(t *testing.T) {
		_, err := parseFileOrUrl("%zz")

		require.Error(t, err)
	})
}
