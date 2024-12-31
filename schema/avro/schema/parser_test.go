package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	testcases := []struct {
		name string
		s    *Schema
		b    []byte
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "null",
			s:    &Schema{Type: []string{"null"}},
			b:    []byte{0, 0, 0, 0, 1},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "boolean true",
			s:    &Schema{Type: []string{"boolean"}},
			b:    []byte{0, 0, 0, 0, 1, 1},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name: "boolean false",
			s:    &Schema{Type: []string{"boolean"}},
			b:    []byte{0, 0, 0, 0, 1, 0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, false, v)
			},
		},
		{
			name: "integer 1",
			s:    &Schema{Type: []string{"int"}},
			b:    []byte{0, 0, 0, 0, 1, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1), v)
			},
		},
		{
			name: "integer -64",
			s:    &Schema{Type: []string{"int"}},
			b:    []byte{0, 0, 0, 0, 1, 0x7f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-64), v)
			},
		},
		{
			name: "integer 64",
			s:    &Schema{Type: []string{"int"}},
			b:    []byte{0, 0, 0, 0, 1, 0x80, 0x01},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(64), v)
			},
		},
		{
			name: "float 3.14159",
			s:    &Schema{Type: []string{"float"}},
			b:    []byte{0, 0, 0, 0, 1, 0xD0, 0xF, 0x49, 0x40},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(3.14159), v)
			},
		},
		{
			name: "double 3.14163",
			s:    &Schema{Type: []string{"double"}},
			b:    []byte{0, 0, 0, 0, 1, 0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.14159, v)
			},
		},
		{
			name: "string",
			s:    &Schema{Type: []string{"string"}},
			b:    []byte{0, 0, 0, 0, 1, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "record",
			s: &Schema{
				Type: []string{"record"},
				Fields: []Schema{
					{Type: []string{"string"}, Name: "foo"},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "foo"}, v)
			},
		},
		{
			name: "array",
			s: &Schema{
				Type:  []string{"array"},
				Items: &Schema{Type: []string{"int"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0x06, 0x2, 0x4, 0x6, 0x0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, v)
			},
		},
		{
			name: "record with array and string - make sure last 0 is read from array",
			s: &Schema{
				Type: []string{"record"},
				Fields: []Schema{
					{Name: "list", Type: []string{"array"}, Items: &Schema{Type: []string{"int"}}},
					{Name: "foo", Type: []string{"string"}},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0x06, 0x2, 0x4, 0x6, 0x0, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "foo", "list": []interface{}{int64(1), int64(2), int64(3)}}, v)
			},
		},
		{
			name: "enum",
			s: &Schema{
				Type:    []string{"enum"},
				Symbols: []string{"foo", "bar", "yuh"},
			},
			b: []byte{0, 0, 0, 0, 1, 0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "enum negative index",
			s: &Schema{
				Type:    []string{"enum"},
				Symbols: []string{"foo", "bar", "yuh"},
			},
			b: []byte{0, 0, 0, 0, 1, 0x1},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "index -1 out of enum range")
			},
		},
		{
			name: "enum index out of range",
			s: &Schema{
				Type:    []string{"enum"},
				Symbols: []string{"foo", "bar", "yuh"},
			},
			b: []byte{0, 0, 0, 0, 1, 0x8},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "index 4 out of enum range")
			},
		},
		{
			name: "map",
			s: &Schema{
				Type:   []string{"map"},
				Values: &Schema{Type: []string{"long"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0x2, 0x06, 0x66, 0x6f, 0x6f, 0x12},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(9)}, v)
			},
		},
		{
			name: "union [null, string] with value NULL",
			s: &Schema{
				Type: []string{"null", "string"},
			},
			b: []byte{0, 0, 0, 0, 1, 0x0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "union [null, string] with value foo",
			s: &Schema{
				Type: []string{"null", "string"},
			},
			b: []byte{0, 0, 0, 0, 1, 0x2, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "union schema [null, string] with value foo",
			s: &Schema{
				Schema: []*Schema{
					{Type: []string{"null"}},
					{Type: []string{"string"}},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0x2, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "fixed",
			s: &Schema{
				Type: []string{"fixed"},
				Size: 3,
			},
			b: []byte{0, 0, 0, 0, 1, 0x1, 0x2, 0x3},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{1, 2, 3}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := Parser{Schema: tc.s}
			v, err := p.Parse(tc.b)
			tc.test(t, v, err)
		})
	}
}
