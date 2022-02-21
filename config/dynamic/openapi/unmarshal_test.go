package openapi

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestContent_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		s    string
		fn   func(t *testing.T, c Content)
	}{
		{
			"empty",
			"",
			func(t *testing.T, c Content) {
				require.Len(t, c, 0)
			},
		},
		{
			"with ref",
			`
application/json:
  schema:
    $ref: '#/components/schemas/Foo'
`,
			func(t *testing.T, c Content) {
				require.Len(t, c, 1)
				require.Contains(t, c, "application/json")
			},
		},
	}

	for _, test := range testcases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := Content{}
			err := yaml.Unmarshal([]byte(test.s), &c)
			require.NoError(t, err)
			test.fn(t, c)
		})
	}
}
