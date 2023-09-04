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

func TestSchemas_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, c *Config)
	}{
		{
			name: "ref",
			data: `components:
  schemas:
    $ref: 'schemas.yml'
`,
			test: func(t *testing.T, c *Config) {
				require.Equal(t, "schemas.yml", c.Components.Schemas.Ref)
			},
		},
		{
			name: "value",
			data: `components:
  schemas:
    Foo:
      type: number
`,
			test: func(t *testing.T, c *Config) {
				require.NotNil(t, c.Components.Schemas.Value)
				foo := c.Components.Schemas.Value.Get("Foo")
				require.NotNil(t, foo)
				require.NotNil(t, foo.Value)
				require.Equal(t, "number", foo.Value.Type)
			},
		},
		{
			name: "schema ref",
			data: `components:
  schemas:
    Foo:
      $ref: 'schemas.yml'
`,
			test: func(t *testing.T, c *Config) {
				require.NotNil(t, c.Components.Schemas.Value)
				foo := c.Components.Schemas.Value.Get("Foo")
				require.NotNil(t, foo)
				require.Equal(t, "schemas.yml", foo.Ref)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{}
			err := yaml.Unmarshal([]byte(tc.data), &c)
			require.NoError(t, err)
		})
	}
}
