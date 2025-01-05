package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/schema/avro/schema"
	"testing"
)

func TestParser_Parse_Json(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "string",
			input:  `"foo"`,
			schema: &schema.Schema{Type: []interface{}{"string"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name:   "not string",
			input:  `123`,
			schema: &schema.Schema{Type: []interface{}{"string"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "invalid type, expected string but got float\nschema path #/type")
			},
		},
		{
			name:   "int",
			input:  `123`,
			schema: &schema.Schema{Type: []interface{}{"int"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 123, v)
			},
		},
		{
			name:   "not int",
			input:  `12.34`,
			schema: &schema.Schema{Type: []interface{}{"int"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "invalid type, expected int but got float\nschema path #/type")
			},
		},
		{
			name:   "long",
			input:  `123`,
			schema: &schema.Schema{Type: []interface{}{"long"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(123), v)
			},
		},
		{
			name:   "not long",
			input:  `12.34`,
			schema: &schema.Schema{Type: []interface{}{"long"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "invalid type, expected long but got float\nschema path #/type")
			},
		},
		{
			name:   "float",
			input:  `12.3`,
			schema: &schema.Schema{Type: []interface{}{"float"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(12.3), v)
			},
		},
		{
			name:   "double",
			input:  `12.3`,
			schema: &schema.Schema{Type: []interface{}{"double"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 12.3, v)
			},
		},
		{
			name:   "boolean",
			input:  `true`,
			schema: &schema.Schema{Type: []interface{}{"boolean"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name:   "enum",
			input:  `"foo"`,
			schema: &schema.Schema{Type: []interface{}{"enum"}, Symbols: []string{"foo", "bar"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name:   "enum but not in symbols",
			input:  `"yuh"`,
			schema: &schema.Schema{Type: []interface{}{"enum"}, Symbols: []string{"foo", "bar"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "value 'yuh' does not match one in the symbols [foo, bar]\nschema path #/enum")
			},
		},
		{
			name:  "record",
			input: `{"foo": "bar"}`,
			schema: &schema.Schema{Type: []interface{}{"record"}, Fields: []schema.Schema{
				{Name: "foo", Type: []interface{}{"string"}},
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name:   "wrapped",
			input:  `"foo"`,
			schema: &schema.Schema{Type: []interface{}{&schema.Schema{Type: []interface{}{"string"}}}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name:   "array",
			input:  `["foo", "bar"]`,
			schema: &schema.Schema{Type: []interface{}{"array"}, Items: &schema.Schema{Type: []interface{}{"string"}}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v interface{}
			err := json.Unmarshal([]byte(tc.input), &v)
			require.NoError(t, err)
			p := schema.Parser{
				Schema: tc.schema,
			}
			v, err = p.Parse(v)
			tc.test(t, v, err)
		})
	}
}
