package dynamic_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"testing"
)

func TestParseYamlPlain(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "string",
			data: `foo`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "int",
			data: `123`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, 123, v)
			},
		},
		{
			name: "float",
			data: `12.3`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, 12.3, v)
			},
		},
		{
			name: "bool",
			data: `true`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name: "time string",
			data: `2025-04-22T14:30:00Z`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "2025-04-22T14:30:00Z", v)
			},
		},
		{
			name: "null",
			data: `null`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "object",
			data: `value: foo`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"value": "foo"}, v)
			},
		},
		{
			name: "object with null value",
			data: `value: null`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"value": nil}, v)
			},
		},
		{
			name: "array",
			data: `- foo`,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{"foo"}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v yamlData
			err := yaml.Unmarshal([]byte(tc.data), &v)
			tc.test(t, v.Value, err)
		})
	}
}

type yamlData struct {
	Value any
}

func (d *yamlData) UnmarshalYAML(node *yaml.Node) error {
	var err error
	d.Value, err = dynamic.ParseYamlPlain(node)
	return err
}
