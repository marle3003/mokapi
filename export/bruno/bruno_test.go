package bruno

import (
	"mokapi/version"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBruno_MarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		c    Collection
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "empty",
			c:    Collection{Version: &version.Version{Major: 1}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `opencollection: 1.0.0
bundled: false
`, s)
			},
		},
		{
			name: "with info",
			c: Collection{
				Version: &version.Version{Major: 1},
				Info: Info{
					Name:    "name",
					Summary: "summary",
					Version: "1.0",
				},
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `opencollection: 1.0.0
info:
    name: name
    summary: summary
    version: "1.0"
bundled: false
`, s)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := yaml.Marshal(tc.c)
			tc.test(t, string(b), err)
		})
	}
}
