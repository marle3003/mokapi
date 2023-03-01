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
				ct := c["application/json"].ContentType
				require.Equal(t, "application", ct.Type)
				require.Equal(t, "json", ct.Subtype)
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

func TestResponses_UnmarshalYAML(t *testing.T) {
	s := `"200": {
"description": "Success"
}`
	r := &Responses{}
	err := yaml.Unmarshal([]byte(s), &r)
	require.NoError(t, err)
	require.Equal(t, "Success", r.GetResponse(200).Description)
}
