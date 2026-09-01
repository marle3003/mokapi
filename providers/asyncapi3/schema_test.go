package asyncapi3_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSchema_Marshal(t *testing.T) {
	testcases := []struct {
		name string
		s    *asyncapi3.SchemaRef
		json string
		yaml string
		test func(t *testing.T, json, yaml string, err error)
	}{
		{
			name: "ref",
			s: &asyncapi3.SchemaRef{
				Reference: dynamic.Reference[*asyncapi3.SchemaRef]{
					Ref: "/foo",
				},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"/foo"}`, json)
				require.Equal(t, "$ref: /foo\n", yaml)
			},
		},
		{
			name: "value",
			s: &asyncapi3.SchemaRef{
				Value: &schema.Schema{Description: "foo"},
			},
			test: func(t *testing.T, json, yaml string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"description":"foo"}`, json)
				require.Equal(t, "description: foo\n", yaml)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb, err := json.Marshal(tc.s)
			if err != nil {
				tc.test(t, "", "", err)
			}
			yb, err := yaml.Marshal(tc.s)
			tc.test(t, string(jb), string(yb), err)
		})
	}
}
