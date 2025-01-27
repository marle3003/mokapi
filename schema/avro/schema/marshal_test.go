package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/avro/schema"
	"testing"
)

func TestMarshal(t *testing.T) {
	testcases := []struct {
		name   string
		input  interface{}
		schema *schema.Schema
		test   func(t *testing.T, b []byte, err error)
	}{
		{
			name:   "null",
			input:  nil,
			schema: &schema.Schema{Type: []interface{}{"null"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{}, b)
			},
		},
		{
			name:   "boolean true",
			input:  true,
			schema: &schema.Schema{Type: []interface{}{"boolean"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x1}, b)
			},
		},
		{
			name:   "boolean false",
			input:  false,
			schema: &schema.Schema{Type: []interface{}{"boolean"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x0}, b)
			},
		},
		{
			name:   "integer 1",
			input:  1,
			schema: &schema.Schema{Type: []interface{}{"int"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x2}, b)
			},
		},
		{
			name:   "integer -64",
			input:  int64(-64),
			schema: &schema.Schema{Type: []interface{}{"int"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x7f}, b)
			},
		},
		{
			name:   "integer 64",
			input:  int64(64),
			schema: &schema.Schema{Type: []interface{}{"int"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x80, 0x1}, b)
			},
		},
		{
			name:   "float 3.14159",
			input:  float32(3.14159),
			schema: &schema.Schema{Type: []interface{}{"float"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0xD0, 0xF, 0x49, 0x40}, b)
			},
		},
		{
			name:   "double 3.14159",
			input:  3.14159,
			schema: &schema.Schema{Type: []interface{}{"double"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40}, b)
			},
		},
		{
			name:   "string",
			input:  "foo",
			schema: &schema.Schema{Type: []interface{}{"string"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x6, 0x66, 0x6f, 0x6f}, b)
			},
		},
		{
			name:  "record",
			input: map[string]interface{}{"foo": "bar"},
			schema: &schema.Schema{
				Type:   []interface{}{"record"},
				Fields: []schema.Schema{{Type: []interface{}{"string"}, Name: "foo"}},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x06, 0x62, 0x61, 0x72}, b)
			},
		},
		{
			name:  "array",
			input: []interface{}{"foo", "bar"},
			schema: &schema.Schema{
				Type:  []interface{}{"array"},
				Items: &schema.Schema{Type: []interface{}{"string"}},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x4, 0x6, 0x66, 0x6f, 0x6f, 0x6, 0x62, 0x61, 0x72}, b)
			},
		},
		{
			name:   "enum",
			input:  "foo",
			schema: &schema.Schema{Type: []interface{}{"enum"}, Symbols: []string{"foo"}},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x0}, b)
			},
		},
		{
			name:   "enum not found",
			input:  "bar",
			schema: &schema.Schema{Type: []interface{}{"enum"}, Symbols: []string{"foo"}},
			test: func(t *testing.T, b []byte, err error) {
				require.EqualError(t, err, "value 'bar' does not match one in the symbols [foo]\nschema path #/enum")
			},
		},
		{
			name:  "map",
			input: map[string]interface{}{"foo": 9},
			schema: &schema.Schema{
				Type:   []interface{}{"map"},
				Values: &schema.Schema{Type: []interface{}{"long"}},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x2, 0x06, 0x66, 0x6f, 0x6f, 0x12}, b)
			},
		},
		{
			name:  "union [null, string] with value NULL",
			input: nil,
			schema: &schema.Schema{
				Type: []interface{}{"null", "string"},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x0}, b)
			},
		},
		{
			name:  "union [null, string] with value foo",
			input: "foo",
			schema: &schema.Schema{
				Type: []interface{}{"null", "string"},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x2, 0x6, 0x66, 0x6f, 0x6f}, b)
			},
		},
		{
			name:  "union [null, string] with value foo",
			input: "foo",
			schema: &schema.Schema{
				Type: []interface{}{
					&schema.Schema{Type: []interface{}{"null"}},
					&schema.Schema{Type: []interface{}{"string"}},
				},
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x2, 0x6, 0x66, 0x6f, 0x6f}, b)
			},
		},
		{
			name:  "fixed",
			input: "foo",
			schema: &schema.Schema{
				Type: []interface{}{"fixed"},
				Size: 3,
			},
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{0x6, 0x66, 0x6f, 0x6f}, b)
			},
		},
		{
			name:  "fixed but error",
			input: "foo",
			schema: &schema.Schema{
				Type: []interface{}{"fixed"},
				Size: 2,
			},
			test: func(t *testing.T, b []byte, err error) {
				require.EqualError(t, err, "invalid fixed size, expected 2 but got 3\nschema path #/fixed")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.input)
			tc.test(t, b, err)
		})
	}
}
