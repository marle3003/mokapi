package ref_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/schema/json/ref"
	"testing"
)

func TestReference_Unmarshal(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "json empty",
			test: func(t *testing.T) {
				s := `{}`
				var r *testRef
				err := json.Unmarshal([]byte(s), &r)
				require.NoError(t, err)
				require.Equal(t, "", r.Ref)
			},
		},
		{
			name: "json $ref",
			test: func(t *testing.T) {
				s := `{"$ref": "foo"}`
				var r *testRef
				err := json.Unmarshal([]byte(s), &r)
				require.NoError(t, err)
				require.Equal(t, "foo", r.Ref)
			},
		},
		{
			name: "json value",
			test: func(t *testing.T) {
				s := `{"foo": "bar"}`
				var f *testRef
				err := json.Unmarshal([]byte(s), &f)
				require.NoError(t, err)
				require.Equal(t, "", f.Ref)
				require.Equal(t, "bar", f.Value.Foo)
			},
		},
		{
			name: "json with ref, value should be ignored",
			test: func(t *testing.T) {
				s := `{"$ref": "foo"}`
				var f *testRef
				err := json.Unmarshal([]byte(s), &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Ref)
				require.Nil(t, f.Value)
			},
		},
		{
			name: "yaml empty",
			test: func(t *testing.T) {
				s := `{}`
				var r *testRef
				err := yaml.Unmarshal([]byte(s), &r)
				require.NoError(t, err)
				require.Equal(t, "", r.Ref)
			},
		},
		{
			name: "yaml $ref",
			test: func(t *testing.T) {
				s := `$ref: foo`
				var r *testRef
				err := yaml.Unmarshal([]byte(s), &r)
				require.NoError(t, err)
				require.Equal(t, "foo", r.Ref)
			},
		},
		{
			name: "yaml value",
			test: func(t *testing.T) {
				s := `foo: bar`
				var f *testRef
				err := yaml.Unmarshal([]byte(s), &f)
				require.NoError(t, err)
				require.Equal(t, "", f.Ref)
				require.Equal(t, "bar", f.Value.Foo)
			},
		},
		{
			name: "yaml with ref, value should be ignored",
			test: func(t *testing.T) {
				s := `$ref: foo`
				var f *testRef
				err := yaml.Unmarshal([]byte(s), &f)
				require.NoError(t, err)
				require.Equal(t, "foo", f.Ref)
				require.Nil(t, f.Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

type testRef struct {
	ref.Reference
	Value *value
}

type value struct {
	Foo string `json:"foo" yaml:"foo"`
}

func (r *testRef) UnmarshalYAML(node *yaml.Node) error {
	return r.UnmarshalYaml(node, &r.Value)
}

func (r *testRef) UnmarshalJSON(b []byte) error {
	return r.UnmarshalJson(b, &r.Value)
}
