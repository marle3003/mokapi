package dynamic_test

import (
	"encoding/json"
	"mokapi/config/dynamic"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type foo struct {
	Foo string `yaml:"foo"`
}
type ref struct {
	dynamic.Reference[*ref]
	Value *foo
}

func TestReference_Unmarshal(t *testing.T) {

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "JSON with $ref",
			test: func(t *testing.T) {
				var r *ref
				err := json.Unmarshal([]byte(`{"$ref":"foo","summary":"summary","description":"description"}`), &r)
				require.NoError(t, err)
				require.Equal(t, &ref{
					Reference: dynamic.Reference[*ref]{
						Ref:         "foo",
						Summary:     "summary",
						Description: "description",
					},
				}, r)
			},
		},
		{
			name: "JSON value",
			test: func(t *testing.T) {
				var r *ref
				err := json.Unmarshal([]byte(`{"foo":"bar"}`), &r)
				require.NoError(t, err)
				require.Equal(t, &ref{
					Value: &foo{Foo: "bar"},
				}, r)
			},
		},
		{
			name: "YAML with $ref",
			test: func(t *testing.T) {
				var r *ref
				err := yaml.Unmarshal([]byte(`
$ref: foo
summary: summary
description: description
`), &r)
				require.NoError(t, err)
				require.Equal(t, &ref{
					Reference: dynamic.Reference[*ref]{
						Ref:         "foo",
						Summary:     "summary",
						Description: "description",
					},
				}, r)
			},
		},
		{
			name: "YAML value",
			test: func(t *testing.T) {
				var r *ref
				err := yaml.Unmarshal([]byte("foo: bar"), &r)
				require.NoError(t, err)
				require.Equal(t, &ref{
					Value: &foo{Foo: "bar"},
				}, r)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func (r *ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ref) UnmarshalJSON(b []byte) error {
	return r.UnmarshalJson(b, &r.Value)
}
