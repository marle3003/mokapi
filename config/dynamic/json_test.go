package dynamic

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSchema_UnmarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "string",
			test: func(t *testing.T) {
				var s string
				err := UnmarshalJSON([]byte(`"foobar"`), &s)
				require.NoError(t, err)
				require.Equal(t, "foobar", s)
			},
		},
		{
			name: "int",
			test: func(t *testing.T) {
				var n int
				err := UnmarshalJSON([]byte(`123`), &n)
				require.NoError(t, err)
				require.Equal(t, 123, n)
			},
		},
		{
			name: "int8",
			test: func(t *testing.T) {
				var n int8
				err := UnmarshalJSON([]byte(`123`), &n)
				require.NoError(t, err)
				require.Equal(t, int8(123), n)
			},
		},
		{
			name: "int8 overflow",
			test: func(t *testing.T) {
				var n int8
				err := UnmarshalJSON([]byte(`1234567890`), &n)
				require.EqualError(t, err, "overflow number 1234567890")
			},
		},
		{
			name: "float32",
			test: func(t *testing.T) {
				var n float32
				err := UnmarshalJSON([]byte(`123`), &n)
				require.NoError(t, err)
				require.Equal(t, float32(123), n)
			},
		},
		{
			name: "float64",
			test: func(t *testing.T) {
				var n float64
				err := UnmarshalJSON([]byte(`123.5`), &n)
				require.NoError(t, err)
				require.Equal(t, 123.5, n)
			},
		},
		{
			name: "bool",
			test: func(t *testing.T) {
				var b bool
				err := UnmarshalJSON([]byte(`true`), &b)
				require.NoError(t, err)
				require.Equal(t, true, b)
			},
		},
		{
			name: "array",
			test: func(t *testing.T) {
				var v []int
				err := UnmarshalJSON([]byte(`[1,2,3,4]`), &v)
				require.NoError(t, err)
				require.Equal(t, []int{1, 2, 3, 4}, v)
			},
		},
		{
			name: "map",
			test: func(t *testing.T) {
				v := map[string]string{}
				err := UnmarshalJSON([]byte(`{"name": "foo"}`), &v)
				require.NoError(t, err)
				require.Equal(t, map[string]string{"name": "foo"}, v)
			},
		},
		{
			name: "struct",
			test: func(t *testing.T) {
				var v struct {
					Name string
				}
				err := UnmarshalJSON([]byte(`{"name": "foo"}`), &v)
				require.NoError(t, err)
				require.Equal(t, "foo", v.Name)
			},
		},
		{
			name: "struct with tag",
			test: func(t *testing.T) {
				var v struct {
					X string `json:"name"`
				}
				err := UnmarshalJSON([]byte(`{"name": "foo"}`), &v)
				require.NoError(t, err)
				require.Equal(t, "foo", v.X)
			},
		},
		{
			name: "struct with syntax error",
			test: func(t *testing.T) {
				var v struct {
					X string `json:"name"`
				}
				err := UnmarshalJSON([]byte(`{"name": []}`), &v)
				require.EqualError(t, err, "structural error at name: expected string but received an array")
				require.Equal(t, int64(10), err.(*StructuralError).Offset)
			},
		},
		{
			name: "skip string field",
			test: func(t *testing.T) {
				var v struct {
				}
				err := UnmarshalJSON([]byte(`{"name": ""}`), &v)
				require.NoError(t, err)
			},
		},
		{
			name: "skip array field",
			test: func(t *testing.T) {
				var v struct {
					Name string
				}
				err := UnmarshalJSON([]byte(`{"value": [ {"foo": "bar" }], "name": "foo"}`), &v)
				require.NoError(t, err)
				require.Equal(t, "foo", v.Name)
			},
		},
		{
			name: "skip struct field",
			test: func(t *testing.T) {
				var v struct {
					Name string
				}
				err := UnmarshalJSON([]byte(`{"value": {"foo": "bar" }, "name": "foo"}`), &v)
				require.NoError(t, err)
				require.Equal(t, "foo", v.Name)
			},
		},
		{
			name: "number to interface{}",
			test: func(t *testing.T) {
				var v struct {
					Value interface{}
				}
				err := UnmarshalJSON([]byte(`{"value": 12}`), &v)
				require.NoError(t, err)
				require.Equal(t, float64(12), v.Value)
			},
		},
		{
			name: "custom type string",
			test: func(t *testing.T) {
				type special string
				var v struct {
					Name special
				}
				err := UnmarshalJSON([]byte(`{"name": "foo"}`), &v)
				require.NoError(t, err)
				require.Equal(t, special("foo"), v.Name)
			},
		},
		{
			name: "interface with array",
			test: func(t *testing.T) {
				var v struct {
					Values interface{}
				}
				err := UnmarshalJSON([]byte(`{"values": [1,2,3,4]}`), &v)
				require.NoError(t, err)
				require.Equal(t, []interface{}{float64(1), float64(2), float64(3), float64(4)}, v.Values)
			},
		},
		{
			name: "interface with object",
			test: func(t *testing.T) {
				var v struct {
					Value interface{}
				}
				err := UnmarshalJSON([]byte(`{"value": { "foo": "bar" }}`), &v)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v.Value)
			},
		},
		{
			name: "map[string]string",
			test: func(t *testing.T) {
				v := make(map[string]string)
				err := UnmarshalJSON([]byte(`{"value": { "name": { "foo": "bar" } }}`), &v)
				require.Error(t, err)
				require.Equal(t, map[string]string{"value": ""}, v)
			},
		},
		{
			name: "string with null value",
			test: func(t *testing.T) {
				v := make(map[string]string)
				err := UnmarshalJSON([]byte(`{"value": null }`), &v)
				require.NoError(t, err)
				require.Equal(t, map[string]string{"value": ""}, v)
			},
		},
		{
			name: "struct with null value",
			test: func(t *testing.T) {
				type t1 struct{}
				v := map[string]*t1{}
				err := UnmarshalJSON([]byte(`{"value": null }`), &v)
				require.NoError(t, err)
				require.Equal(t, map[string]*t1{"value": nil}, v)
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
