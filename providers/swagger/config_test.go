package swagger

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchema_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		test func(t *testing.T, c *Config, err error)
	}{
		{
			name: "info",
			s:    `{"swagger": "2.0", "info": { "title": "FOO", "description": "BAR" }}`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "FOO", c.Info.Name)
				require.Equal(t, "BAR", c.Info.Description)
			},
		},
		{
			name: "schema",
			s:    `{"definitions": { "Foo": { "type": "string" } }}`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", c.Definitions["Foo"].Type[0])
			},
		},
		{
			name: "wrong type in schema attribute",
			s:    `{"definitions": { "Foo": { "items": [] } }}`,
			test: func(t *testing.T, c *Config, err error) {
				require.EqualError(t, err, "structural error at definitions.Foo.items: expected object but received an array")
				require.Equal(t, int64(38), err.(*dynamic.StructuralError).Offset)
			},
		},
		{
			name: "wrong type in schema properties attribute",
			s:    `{"definitions": { "Foo": { "properties": { "value": { "items": [] } } } }}`,
			test: func(t *testing.T, c *Config, err error) {
				require.EqualError(t, err, "structural error at definitions.Foo.properties.value.items: expected object but received an array")
				require.Equal(t, int64(65), err.(*dynamic.StructuralError).Offset)
			},
		},
		{
			name: "tags",
			s:    `{ "tags": [{ "name": "foo" }]}`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Len(t, c.Tags, 1)
				require.Equal(t, "foo", c.Tags[0].Name)
			},
		},
		{
			name: "tags in operation",
			s:    `{ "paths": {"/foo": { "get": { "tags": ["foo", "bar"] } }} }`,
			test: func(t *testing.T, c *Config, err error) {
				require.NoError(t, err)
				require.Len(t, c.Paths, 1)
				require.Len(t, c.Paths["/foo"].Get.Tags, 2)
				require.Equal(t, []string{"foo", "bar"}, c.Paths["/foo"].Get.Tags)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{}
			err := json.Unmarshal([]byte(tc.s), c)
			tc.test(t, c, err)
		})
	}
}
