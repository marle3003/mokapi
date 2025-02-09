package schema

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
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
			s:    &Schema{Type: []interface{}{"null"}},
			b:    []byte{0, 0, 0, 0, 1},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "boolean true",
			s:    &Schema{Type: []interface{}{"boolean"}},
			b:    []byte{0, 0, 0, 0, 1, 1},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name: "boolean false",
			s:    &Schema{Type: []interface{}{"boolean"}},
			b:    []byte{0, 0, 0, 0, 1, 0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, false, v)
			},
		},
		{
			name: "integer 1",
			s:    &Schema{Type: []interface{}{"int"}},
			b:    []byte{0, 0, 0, 0, 1, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1), v)
			},
		},
		{
			name: "integer -64",
			s:    &Schema{Type: []interface{}{"int"}},
			b:    []byte{0, 0, 0, 0, 1, 0x7f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-64), v)
			},
		},
		{
			name: "integer 64",
			s:    &Schema{Type: []interface{}{"int"}},
			b:    []byte{0, 0, 0, 0, 1, 0x80, 0x01},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(64), v)
			},
		},
		{
			name: "float 3.14159",
			s:    &Schema{Type: []interface{}{"float"}},
			b:    []byte{0, 0, 0, 0, 1, 0xD0, 0xF, 0x49, 0x40},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float32(3.14159), v)
			},
		},
		{
			name: "double 3.14163",
			s:    &Schema{Type: []interface{}{"double"}},
			b:    []byte{0, 0, 0, 0, 1, 0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x9, 0x40},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.14159, v)
			},
		},
		{
			name: "string",
			s:    &Schema{Type: []interface{}{"string"}},
			b:    []byte{0, 0, 0, 0, 1, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "string invalid length",
			s:    &Schema{Type: []interface{}{"string"}},
			b:    []byte{0, 0, 0, 0, 1, 0x1, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "invalid string length at offset 6: -1")
			},
		},
		{
			name: "record",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{Type: []interface{}{"string"}, Name: "foo"},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "foo"}, v)
			},
		},
		{
			name: "array zero length",
			s: &Schema{
				Type:  []interface{}{"array"},
				Items: &Schema{Type: []interface{}{"int"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{}, v)
			},
		},
		{
			name: "array",
			s: &Schema{
				Type:  []interface{}{"array"},
				Items: &Schema{Type: []interface{}{"int"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0x06, 0x2, 0x4, 0x6, 0x0},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, v)
			},
		},
		{
			name: "array invalid length",
			s: &Schema{
				Type:  []interface{}{"array"},
				Items: &Schema{Type: []interface{}{"int"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0x01, 0x2, 0x4, 0x6, 0x0},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "invalid array length at offset 6: -1")
			},
		},
		{
			name: "record with array and string - make sure last 0 is read from array",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{Name: "list", Type: []interface{}{"array"}, Items: &Schema{Type: []interface{}{"int"}}},
					{Name: "foo", Type: []interface{}{"string"}},
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
				Type:    []interface{}{"enum"},
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
				Type:    []interface{}{"enum"},
				Symbols: []string{"foo", "bar", "yuh"},
			},
			b: []byte{0, 0, 0, 0, 1, 0x1},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "index -1 out of enum range at offset 6")
			},
		},
		{
			name: "enum index out of range",
			s: &Schema{
				Type:    []interface{}{"enum"},
				Symbols: []string{"foo", "bar", "yuh"},
			},
			b: []byte{0, 0, 0, 0, 1, 0x8},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "index 4 out of enum range at offset 6")
			},
		},
		{
			name: "map",
			s: &Schema{
				Type:   []interface{}{"map"},
				Values: &Schema{Type: []interface{}{"long"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0x2, 0x06, 0x66, 0x6f, 0x6f, 0x12},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(9)}, v)
			},
		},
		{
			name: "map invalid length",
			s: &Schema{
				Type:   []interface{}{"map"},
				Values: &Schema{Type: []interface{}{"long"}},
			},
			b: []byte{0, 0, 0, 0, 1, 0x1, 0x06, 0x66, 0x6f, 0x6f, 0x12},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "invalid map length at offset 6: -1")
			},
		},
		{
			name: "union [null, string] with value NULL",
			s: &Schema{
				Type: []interface{}{"null", "string"},
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
				Type: []interface{}{"null", "string"},
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
				Type: []interface{}{
					&Schema{Type: []interface{}{"null"}},
					&Schema{Type: []interface{}{"string"}},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0x2, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "union index out of range",
			s: &Schema{
				Type: []interface{}{
					&Schema{Type: []interface{}{"null"}},
					&Schema{Type: []interface{}{"string"}},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0x4, 0x06, 0x66, 0x6f, 0x6f},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "index 6 out of range in union at offset 2")
			},
		},
		{
			name: "fixed",
			s: &Schema{
				Type: []interface{}{"fixed"},
				Size: 3,
			},
			b: []byte{0, 0, 0, 0, 1, 0x1, 0x2, 0x3},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte{1, 2, 3}, v)
			},
		},
		{
			name: "named enum empty namespace",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{
						Name: "f1",
						Type: []interface{}{
							&Schema{
								Name:    "foo",
								Type:    []interface{}{"enum"},
								Symbols: []string{"foo", "bar", "yuh"},
							},
						},
					},
					{
						Name: "f2",
						Type: []interface{}{"foo"},
					},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"f1": "foo", "f2": "bar"}, v)
			},
		},
		{
			name: "named enum with namespace",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{
						Namespace: "ns1",
						Name:      "f1",
						Type: []interface{}{
							&Schema{
								Name:    "foo",
								Type:    []interface{}{"enum"},
								Symbols: []string{"foo", "bar", "yuh"},
							},
						},
					},
					{
						Name: "f2",
						Type: []interface{}{"ns1.foo"},
					},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"f1": "foo", "f2": "bar"}, v)
			},
		},
		{
			name: "named enum with namespace in name",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{
						Namespace: "ns1",
						Name:      "f1",
						Type: []interface{}{
							&Schema{
								Name:    "ns2.foo",
								Type:    []interface{}{"enum"},
								Symbols: []string{"foo", "bar", "yuh"},
							},
						},
					},
					{
						Name: "f2",
						Type: []interface{}{"ns2.foo"},
					},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"f1": "foo", "f2": "bar"}, v)
			},
		},
		{
			name: "named enum not found",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{
						Namespace: "ns1",
						Name:      "f1",
						Type: []interface{}{
							&Schema{
								Name:    "foo",
								Type:    []interface{}{"enum"},
								Symbols: []string{"foo", "bar", "yuh"},
							},
						},
					},
					{
						Name: "f2",
						Type: []interface{}{"ns2.foo"},
					},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "unknown schema type 'ns2.foo' at 'ns1.f2'")
			},
		},
		{
			name: "named enum with root namespace",
			s: &Schema{
				Namespace: "ns",
				Type:      []interface{}{"record"},
				Fields: []*Schema{
					{
						Name: "f1",
						Type: []interface{}{
							&Schema{
								Name:    "foo",
								Type:    []interface{}{"enum"},
								Symbols: []string{"foo", "bar", "yuh"},
							},
						},
					},
					{
						Name: "f2",
						Type: []interface{}{"foo"},
					},
				},
			},
			b: []byte{0, 0, 0, 0, 1, 0, 0x2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"f1": "foo", "f2": "bar"}, v)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				table = map[string]*Schema{}
			}()

			err := tc.s.Parse(&dynamic.Config{Info: dynamictest.NewConfigInfo()}, &dynamictest.Reader{})
			if err != nil {
				tc.test(t, nil, err)
			}

			p := Parser{Schema: tc.s}
			v, err := p.Parse(tc.b)
			tc.test(t, v, err)
		})
	}
}

func Test(t *testing.T) {
	s := &Schema{
		Type: []interface{}{"record"},
		Fields: []*Schema{
			{Name: "Name", Type: []interface{}{"string"}},
			{Name: "Age", Type: []interface{}{"int"}},
		},
	}
	b, err := s.Marshal(map[string]interface{}{"Name": "Carol", "Age": 29})
	require.NoError(t, err)
	str := hex.EncodeToString(b)
	_ = str
	str = base64.StdEncoding.EncodeToString(b)
	_ = str
}
