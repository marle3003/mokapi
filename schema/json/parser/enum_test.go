package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestParse_Enum(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		d    interface{}
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "string: value in enum",
			s:    schematest.New("string", schematest.WithEnum([]interface{}{"a", "b", "c"})),
			d:    "a",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "a", v)
			},
		},
		{
			name: "string: value not in enum",
			s:    schematest.New("string", schematest.WithEnum([]interface{}{"a", "b", "c"})),
			d:    "z",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value 'z' does not match one in the enumeration [a, b, c]")
			},
		},
		{
			name: "integer: value in enum",
			s:    schematest.New("integer", schematest.WithEnum([]interface{}{1, 2, 3})),
			d:    2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2), v)
			},
		},
		{
			name: "integer: value not in enum",
			s:    schematest.New("integer", schematest.WithEnum([]interface{}{1, 2, 3})),
			d:    9,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '9' does not match one in the enumeration [1, 2, 3]")
			},
		},
		{
			name: "number: value in enum",
			s:    schematest.New("number", schematest.WithEnum([]interface{}{1.1, 2.2, 3.3})),
			d:    2.2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(2.2), v)
			},
		},
		{
			name: "number: value not in enum",
			s:    schematest.New("number", schematest.WithEnum([]interface{}{1.1, 2.2, 3.3})),
			d:    1.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '1.5' does not match one in the enumeration [1.1, 2.2, 3.3]")
			},
		},
		{
			name: "object: value in enum",
			s: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithEnum([]interface{}{map[string]interface{}{"foo": "bar"}, map[string]interface{}{"foo": "baz"}}),
			),
			d: map[string]interface{}{"foo": "bar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "object: value not in enum",
			s: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithEnum([]interface{}{map[string]interface{}{"foo": "bar"}, map[string]interface{}{"foo": "baz"}}),
			),
			d: map[string]interface{}{"foo": "qux"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '{foo: qux}' does not match one in the enumeration [{foo: bar}, {foo: baz}]")
			},
		},
		{
			name: "array: value in enum",
			s: schematest.New("array",
				schematest.WithItems("string"),
				schematest.WithEnum([]interface{}{[]interface{}{"foo", "bar"}, []interface{}{"foo", "baz"}}),
			),
			d: []interface{}{"foo", "bar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, v)
			},
		},
		{
			name: "array: value not in enum",
			s: schematest.New("array",
				schematest.WithItems("string"),
				schematest.WithEnum([]interface{}{[]interface{}{"foo", "bar"}, []interface{}{"foo", "baz"}}),
			),
			d: []interface{}{"foo", "qux"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '[foo, qux]' does not match one in the enumeration [[foo, bar], [foo, baz]]")
			},
		},
		{
			name: "null: value in enum",
			s:    schematest.New("null", schematest.WithEnum([]interface{}{nil})),
			d:    nil,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "null: value not in enum",
			s:    schematest.New("null", schematest.WithEnum([]interface{}{"123"})),
			d:    nil,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := parser.Parser{Schema: tc.s}
			v, err := p.Parse(tc.d)
			tc.test(t, v, err)
		})
	}
}
